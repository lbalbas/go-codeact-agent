package script

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func ExtractScripts(text string) []string {
	var scripts []string
	lines := strings.Split(text, "\n")
	var current []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inBlock && (trimmed == "```powershell" || trimmed == "```ps" || trimmed == "```ps1") {
			inBlock = true
			current = nil
			continue
		}
		if inBlock && trimmed == "```" {
			inBlock = false
			if len(current) > 0 {
				scripts = append(scripts, strings.TrimSpace(strings.Join(current, "\n")))
			}
			continue
		}
		if inBlock {
			current = append(current, line)
		}
	}
	return scripts
}

func Execute(script string, timeoutSeconds int) string {
	if script == "" {
		return "Error: empty script"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Error: script timed out after %ds", timeoutSeconds)
	}

	result := string(output)

	// Cap output at 10KB to avoid blowing up the context window
	const maxLen = 10 * 1024
	if len(result) > maxLen {
		result = result[:maxLen] + "\n... (output truncated)"
	}

	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput:\n%s", err, result)
	}
	if strings.TrimSpace(result) == "" {
		return "Script executed successfully (no output)."
	}
	return strings.TrimSpace(result)
}
