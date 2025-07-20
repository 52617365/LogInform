package internal_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/52617365/LogInform/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadExplanations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yamlContent string
		expected    []internal.Explanation
		expectError bool
	}{
		{
			name: "Valid YAML with multiple explanations",
			yamlContent: `- identifier: "ERROR_001"
  internetExplanation: "Network timeout"
  internalExplanation: "Connection failed"
- identifier: "WARN_002"
  internetExplanation: "High memory usage"
  internalExplanation: "Heap above 80%"`,
			expected: []internal.Explanation{
				{
					Identifier:          "ERROR_001",
					InternetExplanation: "Network timeout",
					InternalExplanation: "Connection failed",
				},
				{
					Identifier:          "WARN_002",
					InternetExplanation: "High memory usage",
					InternalExplanation: "Heap above 80%",
				},
			},
			expectError: false,
		},
		{
			name: "Single explanation",
			yamlContent: `- identifier: "INFO_001"
  internetExplanation: "System started"
  internalExplanation: "Boot sequence complete"`,
			expected: []internal.Explanation{
				{
					Identifier:          "INFO_001",
					InternetExplanation: "System started",
					InternalExplanation: "Boot sequence complete",
				},
			},
			expectError: false,
		},
		{
			name:        "Empty YAML array",
			yamlContent: `[]`,
			expected:    []internal.Explanation{},
			expectError: false,
		},
		{
			name:        "Invalid YAML format",
			yamlContent: `invalid yaml content [`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test the function with yaml content directly
			result, err := internal.LoadExplanationsFromReader(strings.NewReader(testCase.yamlContent))

			if testCase.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expected, result)
			}
		})
	}
}

func TestFindAndPrintMatches(t *testing.T) {
	t.Parallel()

	explanations := []internal.Explanation{
		{Identifier: "ERROR_001", InternetExplanation: "Network error", InternalExplanation: "Connection failed"},
		{Identifier: "WARN_002", InternetExplanation: "Memory warning", InternalExplanation: "High usage"},
		{Identifier: "INFO_003", InternetExplanation: "Info message", InternalExplanation: "Status update"},
	}

	tests := []struct {
		name           string
		content        string
		explanations   []internal.Explanation
		expectedOutput string
	}{
		{
			name: "Single match on one line",
			content: `This is a log line
ERROR_001 occurred at 10:30
Another normal line`,
			explanations:   explanations,
			expectedOutput: "Line 2:1-9 -> ERROR_001\n  \033[1;31mERROR_001\033[0m occurred at 10:30\n  -> Internet: Network error\n  -> Internal: Connection failed\n\n",
		},
		{
			name: "Multiple matches on different lines",
			content: `Starting application
ERROR_001 network issue
Processing data
WARN_002 memory high
INFO_003 status ok`,
			explanations: explanations,
			expectedOutput: "Line 2:1-9 -> ERROR_001\n  \033[1;31mERROR_001\033[0m network issue\n  -> Internet: Network error\n  -> Internal: Connection failed\n\n---\n\nLine 4:1-8 -> WARN_002\n  \033[1;31mWARN_002\033[0m memory high\n  -> Internet: Memory warning\n  -> Internal: High usage\n\n---\n\nLine 5:1-8 -> INFO_003\n  \033[1;31mINFO_003\033[0m status ok\n  -> Internet: Info message\n  -> Internal: Status update\n\n",
		},
		{
			name: "Single identifier line",
			content: `ERROR_001 occurred
Normal line`,
			explanations:   explanations,
			expectedOutput: "Line 1:1-9 -> ERROR_001\n  \033[1;31mERROR_001\033[0m occurred\n  -> Internet: Network error\n  -> Internal: Connection failed\n\n",
		},
		{
			name: "No matches",
			content: `This is a normal log
Nothing special here
Just regular text`,
			explanations:   explanations,
			expectedOutput: "",
		},
		{
			name:           "Empty content",
			content:        "",
			explanations:   explanations,
			expectedOutput: "",
		},
		{
			name: "Empty explanations slice",
			content: `ERROR_001 occurred
WARN_002 happened`,
			explanations:   []internal.Explanation{},
			expectedOutput: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Capture output
			var outputBuffer bytes.Buffer

			// Call the function
			err := internal.FindAndPrintMatches(testCase.content, testCase.explanations, &outputBuffer)
			require.NoError(t, err)

			actualOutput := outputBuffer.String()

			assert.Equal(t, testCase.expectedOutput, actualOutput)
		})
	}
}
