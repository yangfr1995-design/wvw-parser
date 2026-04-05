package config

type Config struct {
	LogFolder              string
	EliteInsightsCLI       string
	SpikeIncreaseThreshold float64
	MinSpikeDPS            int
}

var AppConfig = Config{
	LogFolder:              "/Users/ryan/Desktop/projects/gw2/wvw-logs-parser/sampleLogs",
	EliteInsightsCLI:       "/Users/ryan/Desktop/projects/gw2/wvw-logs-parser/bin/GuildWars2EliteInsights-CLI.dll",
	SpikeIncreaseThreshold: 0.5,
	MinSpikeDPS:            2000,
}
