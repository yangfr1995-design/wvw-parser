package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ryan/wvwlog/analysis"
)

func PrintFight(players []analysis.PlayerSummary) {

	fmt.Println("Fight Summary")

	for _, p := range players {

		fmt.Printf(
			"%s (%s) G%d DPS:%d Damage:%d Downs:%d Deaths:%d\n",
			p.Name,
			p.Profession,
			p.Group,
			p.DPS,
			p.Damage,
			p.Downs,
			p.Deaths,
		)
	}
}

// PrintSpikeDamageAnalysis displays spike damage analysis results
func PrintSpikeDamageAnalysis(spikes map[string][]analysis.SpikeDamage) {
	if len(spikes) == 0 {
		return
	}

	fmt.Println("\n=== Spike Damage Analysis ===")

	// Create a sorted list of player names for consistent output
	var players []string
	for p := range spikes {
		players = append(players, p)
	}
	sort.Strings(players)

	for _, playerName := range players {
		playerSpikes := spikes[playerName]
		if len(playerSpikes) > 0 {
			fmt.Printf("\n%s:\n", playerName)
			for _, spike := range playerSpikes {
				fmt.Printf("  %ds - %ds, Damage: %d, Peak DPS: %d, Avg DPS: %.1f\n",
					spike.StartTime/1000, spike.EndTime/1000, spike.DamageAmount, spike.PeakDamageInS, spike.DPS)
			}
		}
	}
}

// PrintWellSkillTiming displays Well skill synchronization analysis
func PrintWellSkillTiming(skillAnalysis map[int]*analysis.WellSyncAnalysis) {
	if len(skillAnalysis) == 0 {
		fmt.Println("\n=== Well Skill Synchronization ===")
		fmt.Println("No Well skills were cast during this fight")
		return
	}

	fmt.Println("\n=== Well Skill Synchronization Analysis ===")

	for skillID, syncAnalysis := range skillAnalysis {
		fmt.Printf("\n%s (ID: %d) - %d casts\n", syncAnalysis.SkillName, skillID, len(syncAnalysis.AllCastTimes))

		if syncAnalysis.Synchronized {
			fmt.Println("  Status: SYNCHRONIZED ✓")
		} else {
			fmt.Println("  Status: OUT OF SYNC ✗")
		}

		fmt.Printf("  Max timing difference: %.1fs\n", float64(syncAnalysis.MaxTimingDiff)/1000)
		fmt.Printf("  Average timing difference: %.1fs\n", syncAnalysis.AvgTimingDiff/1000)

		// Group information
		fmt.Printf("  Sync groups: %d\n", len(syncAnalysis.SyncWindows))
		for _, group := range syncAnalysis.SyncWindows {
			fmt.Printf("    Time(%.1fs - %.1fs) (max diff: %.1fs): %v\n", float64(group.StartTime)/1000, float64(group.EndTime)/1000, float64(group.Diff)/1000, strings.Join(group.Players, ", "))
		}

		// Player details
		fmt.Println("  Cast times by player:")
		playerTimes := make(map[string][]int)
		for _, ct := range syncAnalysis.AllCastTimes {
			playerTimes[ct.PlayerName] = append(playerTimes[ct.PlayerName], ct.CastTime)
		}

		var playerNames []string
		for p := range playerTimes {
			playerNames = append(playerNames, p)
		}
		sort.Strings(playerNames)

		for _, playerName := range playerNames {
			times := playerTimes[playerName]
			secTimes := make([]string, len(times))
			for i, t := range times {
				secTimes[i] = fmt.Sprintf("%.1fs", float64(t)/1000)
			}
			fmt.Printf("    %s: %v (%d casts)\n", playerName, secTimes, len(times))
		}

		// Out of sync casts per player
		if len(syncAnalysis.OutOfSyncCasts) > 0 {
			oosByPlayer := make(map[string][]int)
			for _, ct := range syncAnalysis.OutOfSyncCasts {
				oosByPlayer[ct.PlayerName] = append(oosByPlayer[ct.PlayerName], ct.CastTime)
			}

			var oosPlayers []string
			for p := range oosByPlayer {
				oosPlayers = append(oosPlayers, p)
			}
			sort.Strings(oosPlayers)

			fmt.Printf("  Out of sync casts: %d\n", len(syncAnalysis.OutOfSyncCasts))
			for _, playerName := range oosPlayers {
				times := oosByPlayer[playerName]
				secTimes := make([]string, len(times))
				for i, t := range times {
					secTimes[i] = fmt.Sprintf("%.1fs", float64(t)/1000)
				}
				fmt.Printf("    %s: %d out-of-sync cast(s) at %v\n", playerName, len(times), secTimes)
			}
		}
	}
}
