package config

type Config struct {
	LogFolder              string
	EliteInsightsCLI       string
	WorkerCount            int
	SpikeIncreaseThreshold float64
}

var AppConfig = Config{
	LogFolder:              "/Users/ryan/Desktop/projects/gw2/wvw-logs-parser/sampleLogs",
	EliteInsightsCLI:       "/Users/ryan/Desktop/projects/gw2/wvw-logs-parser/bin/GuildWars2EliteInsights-CLI.dll",
	WorkerCount:            4,
	SpikeIncreaseThreshold: 0, // Temporarily disabled
}
