package script

import (
	"testing"
)

func TestExtractScripts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "no code blocks",
			input:    "just some text",
			expected: nil,
		},
		{
			name:     "single block",
			input:    "Here's the fix:\n```powershell\necho hello\n```\nDone.",
			expected: []string{"echo hello"},
		},
		{
			name:     "multiple blocks",
			input:    "Step 1:\n```powershell\necho a\n```\nStep 2:\n```powershell\necho b\n```",
			expected: []string{"echo a", "echo b"},
		},
		{
			name:     "ignores non-powershell",
			input:    "```go\nfmt.Println()\n```\n```powershell\necho yes\n```",
			expected: []string{"echo yes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractScripts(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %d scripts, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("script[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
