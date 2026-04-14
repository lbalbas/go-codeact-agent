package script

import "testing"

func TestClean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no formatting",
			input:    "echo 'hello'",
			expected: "echo 'hello'",
		},
		{
			name:     "whitespace padding",
			input:    "   echo 'hello'   \n\n",
			expected: "echo 'hello'",
		},
		{
			name:     "markdown wrapper",
			input:    "```powershell\necho 'hello'\n```",
			expected: "echo 'hello'",
		},
		{
			name:     "generic markdown wrapper",
			input:    "```\necho 'hello'\n```",
			expected: "echo 'hello'",
		},
		{
			name:     "multiline wrapper",
			input:    "```powershell\necho 'hello'\nls\n```",
			expected: "echo 'hello'\nls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clean(tt.input)
			if result != tt.expected {
				t.Errorf("Clean() = %q, want %q", result, tt.expected)
			}
		})
	}
}
