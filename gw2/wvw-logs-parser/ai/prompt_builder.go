package ai

import (
	"encoding/json"

	"github.com/ryan/wvwlog/analysis"
)

func BuildPrompt(players []analysis.PlayerSummary) string {

	data, _ := json.Marshal(players)

	return `
Analyze this WvW fight.

Identify:
- top performers
- players dying too often
- squad composition issues
- boon uptime weaknesses

Data:
` + string(data)
}
