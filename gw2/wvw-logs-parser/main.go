package main

import (
	"fmt"

	"github.com/ryan/wvwlog/config"
	"github.com/ryan/wvwlog/processor"
)

func main() {

	fmt.Println("Starting WvW Log Analyzer")

	jobs := processor.ScanFolder(config.AppConfig.LogFolder)

	processor.StartWorkerPool(
		jobs,
		config.AppConfig.WorkerCount,
		config.AppConfig.EliteInsightsCLI,
	)
}
