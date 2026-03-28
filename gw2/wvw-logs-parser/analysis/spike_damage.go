package analysis

import (
	"fmt"

	"github.com/ryan/wvwlog/parser"
)

// SpikeDamage represents a burst of damage in a short window
type SpikeDamage struct {
	PlayerName    string
	StartTime     int // milliseconds
	EndTime       int // milliseconds
	Duration      int // milliseconds
	DamageAmount  int
	DPS           float64 // damage per second
	PeakDamageInS int     // highest damage in any single second
}

// AnalyzeSpikedDamage finds periods where a player dealt very high damage
// It looks for windows where cumulative damage exceeds a threshold
func AnalyzeSpikedDamage(fight *parser.Fight, windowSizeSeconds int, minThreshold int, increaseThreshold float64, minDPS int) map[string][]SpikeDamage {
	spikes := make(map[string][]SpikeDamage)
	minThresholdFactor := 0.1 // 10% of average bracket damage
	for _, player := range fight.Players {
		if len(player.DpsAll) == 0 || player.DpsAll[0].Dps <= minDPS {
			continue
		}
		// Calculate average bracket damage for this player
		playerID := player.ID
		maxSec := fight.Duration / 1000
		var damagePerSecond []int
		for sec := 0; sec <= maxSec; sec++ {
			dmg := 0
			if entry, ok := fight.DamageTimeline[sec]; ok {
				if pd, ok := entry.Players[playerID]; ok {
					dmg = pd.Damage
				}
			}
			damagePerSecond = append(damagePerSecond, dmg)
		}
		bracketSize := windowSizeSeconds
		nBrackets := len(damagePerSecond) / bracketSize
		if len(damagePerSecond)%bracketSize != 0 {
			nBrackets++
		}
		var totalBracketDamage int
		for i := 0; i < nBrackets; i++ {
			start := i * bracketSize
			end := start + bracketSize
			if end > len(damagePerSecond) {
				end = len(damagePerSecond)
			}
			sum := 0
			for j := start; j < end; j++ {
				sum += damagePerSecond[j]
			}
			totalBracketDamage += sum
		}
		avgBracketDamage := 1
		if nBrackets > 0 {
			avgBracketDamage = totalBracketDamage / nBrackets
		}
		autoMinThreshold := int(float64(avgBracketDamage) * minThresholdFactor)
		spikes[player.Name] = findSpikes(player, fight, autoMinThreshold, increaseThreshold)
	}
	return spikes
}

// findSpikes analyzes a single player's damage data to find spike windows
func findSpikes(player parser.Player, fight *parser.Fight, minThreshold int, increaseThreshold float64) []SpikeDamage {
	var result []SpikeDamage
	playerID := player.ID
	maxSec := fight.Duration / 1000
	fmt.Printf("Player: %s, Fight duration: %d ms (%d sec)\n", player.Name, fight.Duration, maxSec)
	var damagePerSecond []int
	for sec := 0; sec <= maxSec; sec++ {
		dmg := 0
		if entry, ok := fight.DamageTimeline[sec]; ok {
			if pd, ok := entry.Players[playerID]; ok {
				dmg = pd.Damage
			}
		}
		damagePerSecond = append(damagePerSecond, dmg)
	}

	bracketSize := 4
	nBrackets := len(damagePerSecond) / bracketSize
	if len(damagePerSecond)%bracketSize != 0 {
		nBrackets++
	}

	var prevBracketSum int
	for i := 0; i < nBrackets; i++ {
		start := i * bracketSize
		end := start + bracketSize
		if end > len(damagePerSecond) {
			end = len(damagePerSecond)
		}
		sum := 0
		peak := 0
		firstActive, lastActive := -1, -1
		for j := start; j < end; j++ {
			sum += damagePerSecond[j]
			if damagePerSecond[j] > peak {
				peak = damagePerSecond[j]
			}
			if damagePerSecond[j] > 0 {
				if firstActive == -1 {
					firstActive = j
				}
				lastActive = j
			}
		}
		if firstActive == -1 {
			firstActive = start
			lastActive = end - 1
		}
		if i > 0 && prevBracketSum > 0 {
			increase := float64(sum-prevBracketSum) / float64(prevBracketSum)
			if increase >= increaseThreshold {
				spike := SpikeDamage{
					PlayerName:    player.Name,
					StartTime:     firstActive * 1000,
					EndTime:       lastActive * 1000,
					Duration:      (lastActive - firstActive + 1) * 1000,
					DamageAmount:  sum,
					DPS:           float64(sum) / float64(lastActive-firstActive+1),
					PeakDamageInS: peak,
				}
				result = append(result, spike)
			}
		}
		prevBracketSum = sum
	}
	return result
}

// FindDamageSpikesAdvanced analyzes targetDamage1S time series data
// This requires the raw damage time series which contains per-second damage
func FindDamageSpikesAdvanced(player parser.Player, windowSize int, threshold int) []SpikeDamage {
	var spikes []SpikeDamage

	// This would analyze the damage1S field if it contains per-second data
	// The structure shows damage1S is [[ ...array of damage values... ]]
	// We'd need to parse this to find peaks

	return spikes
}
