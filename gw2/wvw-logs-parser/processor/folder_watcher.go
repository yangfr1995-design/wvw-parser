package processor

import (
	"os"
	"path/filepath"
)

func ScanFolder(folder string) []string {

	var logs []string

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".zevtc" {
			logs = append(logs, path)
		}

		return nil
	})

	return logs
}

