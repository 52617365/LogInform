// Package internal provides functionality for reading and processing log explanations.
package internal

import (
	"bufio"
	"fmt"
	"io"
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

// FindAndPrintMatches processes a string content line by line and prints lines containing any identifier
// from the provided slice of [Explanation] structs to the given writer.
func FindAndPrintMatches(content string, explanations []Explanation, writer io.Writer) error {
	if len(explanations) == 0 {
		return nil
	}

	// First pass: collect all matches
	type match struct {
		lineNumber  int
		startIndex  int
		endIndex    int
		line        string
		explanation Explanation
	}
	
	var matches []match
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0
	
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check if line contains any identifier
		for _, explanation := range explanations {
			startIndex := strings.Index(line, explanation.Identifier)
			if startIndex != -1 {
				endIndex := startIndex + len(explanation.Identifier) - 1
				matches = append(matches, match{
					lineNumber:  lineNumber,
					startIndex:  startIndex,
					endIndex:    endIndex,
					line:        line,
					explanation: explanation,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// Second pass: print all matches
	for i, m := range matches {
		// Print line info with arrow
		_, err := fmt.Fprintf(writer, "Line %d:%d-%d -> %s\n", 
			m.lineNumber, m.startIndex+1, m.endIndex+1, m.explanation.Identifier)
		if err != nil {
			return fmt.Errorf("failed to write line info: %w", err)
		}

		// Print the formatted content line with highlighting and indentation
		formattedLine := formatContentLine(m.line, m.explanation.Identifier, 70)
		highlightedLine := strings.ReplaceAll(formattedLine, m.explanation.Identifier, 
			fmt.Sprintf("\033[1;31m%s\033[0m", m.explanation.Identifier))
		_, err = fmt.Fprintf(writer, "  %s\n", highlightedLine)
		if err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}

		// Print explanations with arrows and indentation
		_, err = fmt.Fprintf(writer, "  -> Internet: %s\n", m.explanation.InternetExplanation)
		if err != nil {
			return fmt.Errorf("failed to write explanation: %w", err)
		}

		_, err = fmt.Fprintf(writer, "  -> Internal: %s\n", m.explanation.InternalExplanation)
		if err != nil {
			return fmt.Errorf("failed to write internal explanation: %w", err)
		}

		// Add delimiter if this is not the last match
		if i < len(matches)-1 {
			_, err = fmt.Fprintln(writer, "\n---")
			if err != nil {
				return fmt.Errorf("failed to write delimiter: %w", err)
			}
		}
		
		// Add final newline
		_, err = fmt.Fprintln(writer)
		if err != nil {
			return fmt.Errorf("failed to write final newline: %w", err)
		}
	}

	return nil
}
