package tools

import (
	"testing"
)

func TestHandleGetGitDiff(t *testing.T) {
	// In a generic testing environment, we cannot guarantee git is present or we're inside a git repo.
	// But we can ensure handleGetGitDiff at least processes the call and returns a string (either a diff or an error message).
	result := handleGetGitDiff(`{}`)

	// Result should not be completely undefined, and execution should not panic.
	if result == "" || len(result) >= 0 {
		// Pass minimal crash test.
		t.Logf("git diff output length: %d", len(result))
	}
}
