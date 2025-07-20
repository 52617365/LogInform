package internal_test

import (
	"bytes"
	"testing"

	"github.com/52617365/LogInform/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectExplanations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		explanations   []internal.Explanation
		expectedOutput string
	}{
		{
			name: "Multiple explanations",
			explanations: []internal.Explanation{
				{
					Identifier:          "ERROR_001",
					InternetExplanation: "Network timeout error",
					InternalExplanation: "Connection failed after 30 seconds",
				},
				{
					Identifier:          "WARN_002",
					InternetExplanation: "Memory usage warning",
					InternalExplanation: "Heap usage above 80% threshold",
				},
			},
			expectedOutput: `Loaded 2 explanations:

1. Identifier: ERROR_001
   Internet Explanation: Network timeout error
   Internal Explanation: Connection failed after 30 seconds

2. Identifier: WARN_002
   Internet Explanation: Memory usage warning
   Internal Explanation: Heap usage above 80% threshold

`,
		},
		{
			name: "Single explanation",
			explanations: []internal.Explanation{
				{
					Identifier:          "INFO_001",
					InternetExplanation: "System information",
					InternalExplanation: "Status update message",
				},
			},
			expectedOutput: `Loaded 1 explanations:

1. Identifier: INFO_001
   Internet Explanation: System information
   Internal Explanation: Status update message

`,
		},
		{
			name:           "Empty explanations",
			explanations:   []internal.Explanation{},
			expectedOutput: "Loaded 0 explanations:\n\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var outputBuffer bytes.Buffer

			err := internal.InspectExplanations(testCase.explanations, &outputBuffer)
			require.NoError(t, err)

			actualOutput := outputBuffer.String()
			assert.Equal(t, testCase.expectedOutput, actualOutput)
		})
	}
}
