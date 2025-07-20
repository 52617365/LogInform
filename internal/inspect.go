package internal

import (
	"fmt"
	"io"
)

// InspectExplanations writes a formatted list of explanations to the provided writer.
// It displays the count and details of each [Explanation] in a human-readable format.
func InspectExplanations(explanations []Explanation, writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "Loaded %d explanations:\n\n", len(explanations))
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for i, explanation := range explanations {
		_, err = fmt.Fprintf(writer, "%d. Identifier: %s\n", i+1, explanation.Identifier)
		if err != nil {
			return fmt.Errorf("failed to write identifier: %w", err)
		}

		_, err = fmt.Fprintf(writer, "   Internet Explanation: %s\n", explanation.InternetExplanation)
		if err != nil {
			return fmt.Errorf("failed to write internet explanation: %w", err)
		}

		_, err = fmt.Fprintf(writer, "   Internal Explanation: %s\n\n", explanation.InternalExplanation)
		if err != nil {
			return fmt.Errorf("failed to write internal explanation: %w", err)
		}
	}

	return nil
}
