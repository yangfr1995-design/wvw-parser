package processor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertLog(cli string, log string) (string, error) {

	configPath := "/Users/ryan/Desktop/projects/gw2/wvw-logs-parser/bin/parser.conf"
	// With SaveAtOut=false EI writes the JSON next to the input log file.
	outputDir := filepath.Dir(log)

	// Snapshot existing JSON files so we can detect the new one by diff.
	before := map[string]struct{}{}
	if entries, err := os.ReadDir(outputDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				before[e.Name()] = struct{}{}
			}
		}
	}

	cmd := exec.Command(
		"dotnet",
		cli,
		"-c", configPath,
		log,
	)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	if len(output) > 0 {
		fmt.Printf("EliteInsights [%s]: %s\n", filepath.Base(log), outputStr)
	}
	if err != nil {
		return "", fmt.Errorf("EliteInsights failed: %w", err)
	}

	// EI exits 0 even on parse failures — detect them from its output.
	if strings.Contains(outputStr, "Parsing Failure") {
		// Extract the reason after the last ": " for a concise message.
		reason := "parsing failure"
		if idx := strings.LastIndex(outputStr, ": "); idx >= 0 {
			reason = strings.TrimSpace(outputStr[idx+2:])
		}
		return "", fmt.Errorf("parsing failure: %s", reason)
	}

	// Find whichever new JSON appeared in the output directory after the run.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			if _, existed := before[entry.Name()]; !existed {
				return filepath.Join(outputDir, entry.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("no JSON file generated for log: %s", filepath.Base(log))
}
