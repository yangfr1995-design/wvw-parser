package analysis

import "github.com/ryan/wvwlog/parser"

func BuildFightSummary(fight *parser.Fight) []PlayerSummary {

	var out []PlayerSummary

	for _, p := range fight.Players {

		dps := 0
		damage := 0

		if len(p.DpsAll) > 0 {
			dps = p.DpsAll[0].Dps
			damage = p.DpsAll[0].Damage
		}

		downs := 0
		deaths := 0

		if len(p.Defenses) > 0 {
			downs = p.Defenses[0].DownCount
			deaths = p.Defenses[0].DeadCount
		}

		out = append(out, PlayerSummary{
			Name:       p.Name,
			Profession: p.Profession,
			Group:      p.Group,
			DPS:        dps,
			Damage:     damage,
			Downs:      downs,
			Deaths:     deaths,
		})
	}

	return out
}
