package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleReadFileContents(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		args     string
		expected string
		isError  bool
	}{
		{
			name:     "valid file",
			args:     `{"filepath": "` + strings.ReplaceAll(testFile, "\\", "\\\\") + `"}`,
			expected: "hello world",
			isError:  false,
		},
		{
			name:     "invalid json",
			args:     `{"filepath": "`,
			expected: "Error parsing arguments",
			isError:  true,
		},
		{
			name:     "missing file",
			args:     `{"filepath": "/does/not/exist.txt"}`,
			expected: "Error reading file",
			isError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleReadFileContents(tt.args)
			if tt.isError {
				if !strings.HasPrefix(result, "Error") {
					t.Errorf("handleReadFileContents() expected error response, got: %v", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("handleReadFileContents() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}
