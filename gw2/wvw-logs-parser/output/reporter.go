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
				fmt.Printf("  Duration: %dms, Damage: %d, Peak DPS: %d, Avg DPS: %.1f\n",
					spike.Duration, spike.DamageAmount, spike.PeakDamageInS, spike.DPS)
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

		fmt.Printf("  Max timing difference: %dms\n", syncAnalysis.MaxTimingDiff)
		fmt.Printf("  Average timing difference: %.0fms\n", syncAnalysis.AvgTimingDiff)

		// Group information
		fmt.Printf("  Sync groups: %d\n", len(syncAnalysis.SyncWindows))
		for _, group := range syncAnalysis.SyncWindows {
			fmt.Printf("    Time(%d - %d) (max diff: %dms): %v\n", group.StartTime, group.EndTime, group.Diff, strings.Join(group.Players, ", "))
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
			fmt.Printf("    %s: %v (%d casts)\n", playerName, times, len(times))
		}
	}
}
