package processor

import (
	"fmt"
	"os"

	"github.com/ryan/wvwlog/analysis"
	"github.com/ryan/wvwlog/config"
	"github.com/ryan/wvwlog/output"
	"github.com/ryan/wvwlog/parser"
)

func StartWorkerPool(logs []string, workers int, cli string) {

	jobs := make(chan string)

	for w := 0; w < workers; w++ {

		go func() {
			for log := range jobs {

				jsonFile, err := ConvertLog(cli, log)
				if err != nil {
					fmt.Println("convert error:", err)
					continue
				}

				fight, err := parser.ParseFight(jsonFile)
				if err != nil {
					fmt.Println("parse error:", err)
					// Clean up JSON file even if parsing fails
					os.Remove(jsonFile)
					continue
				}

				// summary := analysis.BuildFightSummary(fight)

				// output.PrintFight(summary)

				// Run spike damage analysis
				spikeDamage := analysis.AnalyzeSpikedDamage(fight, 5, 0, config.AppConfig.SpikeIncreaseThreshold, config.AppConfig.MinSpikeDPS)
				output.PrintSpikeDamageAnalysis(spikeDamage)

				// Run well skill timing analysis
				wellTiming := analysis.AnalyzeWellSkillTiming(fight)
				output.PrintWellSkillTiming(wellTiming)

				// Clean up the JSON file after processing
				err = os.Remove(jsonFile)
				if err != nil {
					fmt.Printf("warning: failed to delete JSON file %s: %v\n", jsonFile, err)
				}
			}
		}()
	}

	for _, log := range logs {
		jobs <- log
	}

	close(jobs)
}
