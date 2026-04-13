package script

import (
	"os/exec"
	"strings"
)

func Clean(script string) string {
	// Trim whitespace first
	script = strings.TrimSpace(script)

	// Check if wrapped in markdown code blocks
	if strings.HasPrefix(script, "```") && strings.HasSuffix(script, "```") {
		// Remove the first line (```powershell or ```)
		lines := strings.Split(script, "\n")
		if len(lines) > 2 {
			// Remove first and last line
			script = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	return strings.TrimSpace(script)
}

func Execute(script string) string {
	if script == "" {
		return "Error: No script found to execute."
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.CombinedOutput()

	// If there is no output but the command succeeded, let the AI know
	if len(output) == 0 && err == nil {
		return "Script executed successfully (no output)."
	}

	return string(output)
}
