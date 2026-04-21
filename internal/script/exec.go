package script

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var scriptBlockRegex = regexp.MustCompile(`(?is)\x60\x60\x60(?:powershell|ps|ps1)\s*(.*?)\x60\x60\x60`)

func ExtractScripts(text string) []string {
	var scripts []string

	matches := scriptBlockRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) > 1 {
			scripts = append(scripts, strings.TrimSpace(match[1]))
		}
	}

	return scripts
}

func Execute(script string, timeoutSeconds int) string {
	var outBuf bytes.Buffer

	if script == "" {
		return "Error: empty script"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Error: script timed out after %ds. Partial Output:\n%s", timeoutSeconds, outBuf.String())
	}

	result := string(outBuf.String())

	// Cap output at 10KB to avoid blowing up the context window
	const maxLen = 10 * 1024
	if len(result) > maxLen {
		result = result[:maxLen] + "\n... (output truncated)"
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode := exitErr.ExitCode()
			return fmt.Sprintf("Error: %v (exit code: %d)\nOutput:\n%s", err, exitCode, result)
		}
		return fmt.Sprintf("Error: %v\nOutput:\n%s", err, result)
	}
	if strings.TrimSpace(result) == "" {
		return "Script executed successfully (no output)."
	}
	return strings.TrimSpace(result)
}
