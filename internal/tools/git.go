package tools

import (
	"fmt"
	"os/exec"
)

// handleGetGitDiff runs `git diff` and returns the output.
func handleGetGitDiff(args string) string {
	cmd := exec.Command("git", "diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error getting git diff: %v", err)
	}
	return string(output)
}
