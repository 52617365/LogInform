// Package internal provides functionality for reading and processing log explanations.
package internal

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	ExplanationLocation string // file, github or gitlab
	// TODO: SFTP support
	ExplanationPath string // Local path or GitHub/Gitlab link
	PrivateRemote   bool   // If this is true then we will expect that GITLAB_TOKEN or GITHUB_TOKEN is set.
}

func fetchExplanationsFromRemoteVcs() io.Reader {
	if viper.GetString("ExplanationLocation") == "file" {
		panic("should not come to this func")
	}

	req, err := http.NewRequest("GET", viper.GetString("ExplanationPath"), nil)
	if err != nil {
		slog.Error("there was an error constructing a request", "err", err, viper.Get("ExplanationLocation"), viper.GetString("ExplanationPath"))
	}

	if viper.GetString("ExplanationLocation") == "github" {
		if viper.GetBool("PrivateRemote") {
			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				slog.Error("remote is private but GITHUB_TOKEN not set")
			}
			// Github uses Authorization Bearer <token>
			req.Header.Set("Authorization", "Bearer "+githubToken)
		}
	} else if viper.GetString("ExplanationLocation") == "gitlab" {
		if viper.GetBool("PrivateRemote") {
			gitlabToken := os.Getenv("GITLAB_TOKEN")
			if gitlabToken == "" {
				slog.Error("remote is private but GITLAB_TOKEN not set")
			}

			// Gitlab uses PRIVATE-TOKEN <token>
			req.Header.Set("PRIVATE-TOKEN", gitlabToken)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error fetching explanations from remote", "err", err, "remote", viper.Get("ExplanationLocation"), "private", viper.Get("PrivateRemote"), "path", viper.Get("ExplanationPath"))
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("got non ok error code when fetching explanations from remote", "err", err, "remote", viper.Get("ExplanationLocation"), "private", viper.Get("PrivateRemote"), "path", viper.Get("ExplanationPath"), "code", resp.StatusCode)
		return nil
	}

	// resp.body is base64 encoded, decode first

	bodyContents, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response from remote rest api call", "err", err, "remote", viper.Get("ExplanationLocation"), "private", viper.Get("PrivateRemote"), "path", viper.Get("ExplanationPath"), "code", resp.StatusCode)
	}

	var contents struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal(bodyContents, &contents); err != nil {
		slog.Error("error unmarshaling JSON response", "error", err)
		return nil
	}

	// The actual data is in the content JSON key
	data, err := base64.StdEncoding.DecodeString(contents.Content)
	if err != nil {
		slog.Error("error decoding response", "err", err, "remote", viper.Get("ExplanationLocation"), "private", viper.Get("PrivateRemote"), "path", viper.Get("ExplanationPath"), "code", resp.StatusCode)
	}
	return bytes.NewReader(data)
}

// loadExplanationsFromReader reads YAML content from an io.Reader and returns a slice of [Explanation] structs.
// It uses streaming YAML parsing to avoid loading the entire content into memory at once.
func loadExplanationsFromReader(reader io.Reader) ([]Explanation, error) {
	var explanations []Explanation

	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&explanations)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	return explanations, nil
}

func FetchExplanations() []Explanation {
	// populated from local file, github or gitlab depending on configuration
	var dataReader io.Reader

	if viper.GetString("ExplanationPath") == "" {
		panic("catch in validation")
	}
	if viper.Get("ExplanationLocation") == "file" {
		file, err := os.Open(viper.GetString("ExplanationPath"))
		if err != nil {
			slog.Error("error reading local explanations file", "explanation_location", viper.GetString("ExplanationLocation"), "path", viper.Get("ExplanationPath"), "private", viper.Get("PrivateRemote"))
		}
		dataReader = file
	} else if viper.Get("ExplanationLocation") == "github" || viper.Get("ExplanationLocation") == "gitlab" {
		dataReader = fetchExplanationsFromRemoteVcs()
	} else {
		slog.Error("FetchExplanations was called with ExplanationLocation that is not handled in function", "explanation_location", viper.GetString("ExplanationLocation"), "path", viper.Get("ExplanationPath"), "private", viper.Get("PrivateRemote"))
	}

	explanations, err := loadExplanationsFromReader(dataReader)
	if err != nil {
		slog.Error("the explanations format received from location is in an invalid format", "explanation_location", viper.GetString("ExplanationLocation"), "path", viper.Get("ExplanationPath"), "private", viper.Get("PrivateRemote"))
	}

	return explanations
}

// Explanation represents a log explanation with identifier and description text.
type Explanation struct {
	Identifier          string `yaml:"identifier"`
	InternetExplanation string `yaml:"internetExplanation"`
	InternalExplanation string `yaml:"internalExplanation"`
}

// formatContentLine truncates very long lines and shows context around the identifier
func formatContentLine(line string, identifier string, maxLen int) string {
	if utf8.RuneCountInString(line) <= maxLen {
		return line
	}

	startIndex := strings.Index(line, identifier)
	if startIndex == -1 {
		// If identifier not found, just truncate from the beginning
		runes := []rune(line)
		if maxLen <= 3 {
			return "..."
		}
		return string(runes[:maxLen-3]) + "..."
	}

	// Calculate how much context to show before and after the identifier
	identifierLen := utf8.RuneCountInString(identifier)
	remainingSpace := maxLen - identifierLen - 6 // Reserve space for "..." on both sides

	if remainingSpace <= 0 {
		// If identifier is too long, just truncate the whole line
		runes := []rune(line)
		if maxLen <= 3 {
			return "..."
		}
		return string(runes[:maxLen-3]) + "..."
	}

	contextBefore := remainingSpace / 2
	contextAfter := remainingSpace - contextBefore

	runes := []rune(line)
	startRune := utf8.RuneCountInString(line[:startIndex])
	endRune := startRune + identifierLen

	// Determine the slice boundaries
	lineStart := 0
	lineEnd := len(runes)

	if startRune > contextBefore {
		lineStart = startRune - contextBefore
	}

	if endRune+contextAfter < len(runes) {
		lineEnd = endRune + contextAfter
	}

	result := string(runes[lineStart:lineEnd])

	// Add ellipsis indicators
	if lineStart > 0 {
		result = "..." + result
	}
	if lineEnd < len(runes) {
		result = result + "..."
	}

	return result
}

// FindAndPrintMatches processes content from an io.Reader line by line and prints lines containing any identifier
// from the provided slice of [Explanation] structs to the given writer.
func FindAndPrintMatches(reader io.Reader, explanations []Explanation, writer io.Writer) error {
	if len(explanations) == 0 {
		return nil
	}

	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check if line contains any identifier
		for _, explanation := range explanations {
			startIndex := strings.Index(line, explanation.Identifier)
			if startIndex != -1 {
				endIndex := startIndex + len(explanation.Identifier) - 1
				_, err := fmt.Fprintf(writer, "Line %d:%d-%d -> %s\n",
					lineNumber, startIndex+1, endIndex+1, explanation.Identifier)
				if err != nil {
					return fmt.Errorf("failed to write line info: %w", err)
				}

				formattedLine := formatContentLine(line, explanation.Identifier, 70)

				highlightedLine := strings.ReplaceAll(formattedLine, explanation.Identifier,
					fmt.Sprintf("\033[1;31m%s\033[0m", explanation.Identifier))
				_, err = fmt.Fprintf(writer, "  %s\n", highlightedLine)
				if err != nil {
					return fmt.Errorf("failed to write line: %w", err)
				}

				_, err = fmt.Fprintf(writer, "  -> Internet: %s\n", explanation.InternetExplanation)
				if err != nil {
					return fmt.Errorf("failed to write explanation: %w", err)
				}

				_, err = fmt.Fprintf(writer, "  -> Internal: %s\n", explanation.InternalExplanation)
				if err != nil {
					return fmt.Errorf("failed to write internal explanation: %w", err)
				}
			}
		}
	}
	if scanner.Err() != nil {
		log.Fatalln(scanner.Err())
	}

	return nil
}
