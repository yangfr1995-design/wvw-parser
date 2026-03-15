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
	outputDir := "/Users/ryan/Desktop/projects/gw2-logs/generated"

	// Create output directory if it doesn't exist
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.Command(
		"dotnet",
		cli,
		"-c", configPath,
		log,
	)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("EliteInsights error output: %s\n", string(output))
		return "", fmt.Errorf("EliteInsights failed: %w", err)
	}

	// Find the generated JSON file in the output directory
	// EliteInsights generates files with format: <filename>_<encounter>.json
	logName := filepath.Base(log)
	logNameWithoutExt := strings.TrimSuffix(logName, filepath.Ext(logName))

	// Search for the JSON file matching the input log name
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), logNameWithoutExt) && strings.HasSuffix(entry.Name(), ".json") {
			return filepath.Join(outputDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no JSON file generated for log: %s", logName)
}
