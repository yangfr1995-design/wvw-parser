package parser

import (
	"encoding/json"
	"os"
)

func ParseFight(file string) (*Fight, error) {

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var fight Fight
	err = json.Unmarshal(data, &fight)

	ExtractDamageEvents := ExtractDamageEvents(fight.CombatData)
	AttachDamageTimeline(&fight, ExtractDamageEvents)

	return &fight, err
}

func ExtractDamageEvents(events []RawCombatEvent) []DamageEvent {

	var damage []DamageEvent

	for _, e := range events {

		if e.IsBuff == 0 && e.Value > 0 {

			damage = append(damage, DamageEvent{
				Time:   e.Time,
				Source: e.Src,
				Target: e.Dst,
				Damage: e.Value,
			})
		}
	}

	return damage
}

func AttachDamageTimeline(fight *Fight, events []DamageEvent) {

	timeline := BuildDamageTimeline(events, fight)

	fight.DamageTimeline = timeline
}

func BuildDamageTimeline(
	events []DamageEvent,
	fight *Fight,
) map[int]*TimelineEntry {

	timeline := map[int]*TimelineEntry{}

	playerNames := map[int]string{}

	for _, p := range fight.Players {
		playerNames[p.ID] = p.Name
	}

	for _, e := range events {

		sec := e.Time / 1000
		name := playerNames[e.Source]

		entry, ok := timeline[sec]
		if !ok {
			entry = &TimelineEntry{
				Time:    sec,
				Players: map[int]*PlayerDamage{},
			}
			timeline[sec] = entry
		}

		pd, ok := entry.Players[e.Source]
		if !ok {
			pd = &PlayerDamage{
				PlayerID: e.Source,
				Name:     name,
				Damage:   0,
			}
			entry.Players[e.Source] = pd
		}

		pd.Damage += e.Damage
	}

	return timeline
}
