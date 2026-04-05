package parser

import (
	"bufio"
	"encoding/json"
	"os"
)

func ParseFight(file string) (*Fight, error) {

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fight Fight
	if err := json.NewDecoder(bufio.NewReaderSize(f, 256*1024)).Decode(&fight); err != nil {
		return nil, err
	}

	BuildDamageTimeline(&fight)

	return &fight, nil
}

// BuildDamageTimeline constructs per-second damage data from each player's
// damage1S field. damage1S[0] is a cumulative damage array, so we convert
// it to per-second deltas.
func BuildDamageTimeline(fight *Fight) {

	timeline := map[int]*TimelineEntry{}

	for _, p := range fight.Players {
		if len(p.Damage1S) == 0 || len(p.Damage1S[0]) == 0 {
			continue
		}

		cumulative := p.Damage1S[0]

		for sec := 0; sec < len(cumulative); sec++ {
			dmg := cumulative[sec]
			if sec > 0 {
				dmg = cumulative[sec] - cumulative[sec-1]
			}
			if dmg <= 0 {
				continue
			}

			entry, ok := timeline[sec]
			if !ok {
				entry = &TimelineEntry{
					Time:    sec,
					Players: map[int]*PlayerDamage{},
				}
				timeline[sec] = entry
			}

			entry.Players[p.ID] = &PlayerDamage{
				PlayerID: p.ID,
				Name:     p.Name,
				Damage:   dmg,
			}
		}
	}

	fight.DamageTimeline = timeline
}
