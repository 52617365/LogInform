// Package internal provides functionality for reading and processing log explanations.
package internal

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Explanation represents a log explanation with identifier and description text.
type Explanation struct {
	Identifier          string `yaml:"identifier"`
	InternetExplanation string `yaml:"internetExplanation"`
	InternalExplanation string `yaml:"internalExplanation"`
}

// LoadExplanationsFromReader reads YAML content from an io.Reader and returns a slice of [Explanation] structs.
// It uses streaming YAML parsing to avoid loading the entire content into memory at once.
func LoadExplanationsFromReader(reader io.Reader) ([]Explanation, error) {
	var explanations []Explanation

	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&explanations)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	return explanations, nil
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
