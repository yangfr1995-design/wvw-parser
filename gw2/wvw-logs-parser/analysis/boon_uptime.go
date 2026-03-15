package analysis

import "github.com/ryan/wvwlog/parser"

func CalculateGroupBoons(players []parser.Player) map[int]map[int]float64 {

	result := map[int]map[int]float64{}

	for _, p := range players {

		if _, ok := result[p.Group]; !ok {
			result[p.Group] = map[int]float64{}
		}

		for _, b := range p.BuffUptimes {

			result[p.Group][b.ID] += b.Uptime
		}
	}

	return result
}
