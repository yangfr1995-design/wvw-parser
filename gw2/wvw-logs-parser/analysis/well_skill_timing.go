package analysis

import (
	"sort"

	"github.com/ryan/wvwlog/parser"
)

// WellSkillTiming represents when a player cast a Well skill
type WellSkillTiming struct {
	PlayerName string
	SkillID    int
	SkillName  string
	CastTime   int // milliseconds from start
	Duration   int // skill duration in ms
	Profession string
	Group      int
}

// WellSyncAnalysis shows how synchronized Well skils were across the group
type WellSyncAnalysis struct {
	SkillName      string
	SkillID        int
	AllCastTimes   []WellSkillTiming
	Synchronized   bool              // true if all casts are in a qualifying sync group
	MaxTimingDiff  int               // maximum millisecond difference between casts
	AvgTimingDiff  float64           // average difference
	SyncWindows    []SyncGroup       // groups with >half of well-casting players
	OutOfSyncCasts []WellSkillTiming // casts not in any qualifying sync group
}

// SyncGroup represents players who cast within a small time window
type SyncGroup struct {
	StartTime   int // earliest cast time in group
	EndTime     int // latest cast time in group
	Players     []string
	Diff        int   // end - start
	CastIndices []int // indices into AllCastTimes
}

const (
	// WellOfCorruption GW2 skill ID
	WellOfCorruptionID = 10545
	// WellOfSuffering GW2 skill ID
	WellOfSufferingID = 10546

	// SyncWindowMS: maximum span of a sync sub-window within a round window.
	SyncWindowMS = 5000
	// RoundWindowMS: the broad window used to segment casts into rounds.
	RoundWindowMS = 35000
)

// AnalyzeWellSkillTiming extracts and analyzes Well skill casting patterns
func AnalyzeWellSkillTiming(fight *parser.Fight) map[int]*WellSyncAnalysis {
	analysis := make(map[int]*WellSyncAnalysis)

	wellSkills := map[int]string{
		WellOfCorruptionID: "Well of Corruption",
		// WellOfSufferingID:  "Well of Suffering",
	}

	// Extract all Well skill casts from all players
	for skillID, skillName := range wellSkills {
		var castTimes []WellSkillTiming

		for _, player := range fight.Players {
			if player.Rotation != nil {
				for _, rotEntry := range player.Rotation {
					if rotEntry.ID == WellOfCorruptionID || rotEntry.ID == WellOfSufferingID {
						// Found the skill, extract all cast times
						for _, skill := range rotEntry.Skills {
							castTimes = append(castTimes, WellSkillTiming{
								PlayerName: player.Name,
								SkillID:    skillID,
								SkillName:  skillName,
								CastTime:   skill.CastTime,
								Duration:   skill.Duration,
								Profession: player.Profession,
								Group:      player.Group,
							})
						}
						break
					}
				}
			}
		}

		if len(castTimes) > 0 {
			// Sort by cast time
			sort.Slice(castTimes, func(i, j int) bool {
				return castTimes[i].CastTime < castTimes[j].CastTime
			})

			analysis[skillID] = analyzeSync(skillName, skillID, castTimes)
		}
	}

	return analysis
}

// analyzeSync segments casts into 35s round windows, finds the best 5s sync
// sub-window within each round, and validates that at least half the players
// who cast wells in the fight are present in the sync window.
func analyzeSync(skillName string, skillID int, castTimes []WellSkillTiming) *WellSyncAnalysis {
	result := &WellSyncAnalysis{
		SkillName:    skillName,
		SkillID:      skillID,
		AllCastTimes: castTimes,
	}

	if len(castTimes) == 0 {
		result.Synchronized = true
		return result
	}

	// Count distinct players who cast this well at all.
	playerSet := map[string]bool{}
	for _, ct := range castTimes {
		playerSet[ct.PlayerName] = true
	}
	totalWellPlayers := len(playerSet)
	minPlayersForSync := (totalWellPlayers + 1) / 2 // ceil(total/2)

	syncedIndices := map[int]bool{}

	// Walk through casts in 35s round windows.
	for start := 0; start < len(castTimes); {
		roundStart := castTimes[start].CastTime
		// Find end of this 35s round window.
		end := start
		for end < len(castTimes) && castTimes[end].CastTime-roundStart <= RoundWindowMS {
			end++
		}
		roundIndices := make([]int, 0, end-start)
		for i := start; i < end; i++ {
			roundIndices = append(roundIndices, i)
		}

		// Within this round, sliding-window to find the 5s window with the most casts.
		bestLo, bestCount := 0, 0
		for lo := 0; lo < len(roundIndices); lo++ {
			winStart := castTimes[roundIndices[lo]].CastTime
			count := 0
			for hi := lo; hi < len(roundIndices); hi++ {
				if castTimes[roundIndices[hi]].CastTime-winStart <= SyncWindowMS {
					count++
				} else {
					break
				}
			}
			if count > bestCount {
				bestCount = count
				bestLo = lo
			}
		}

		// Collect the indices inside the best 5s window.
		winStart := castTimes[roundIndices[bestLo]].CastTime
		var winIndices []int
		for _, idx := range roundIndices[bestLo:] {
			if castTimes[idx].CastTime-winStart <= SyncWindowMS {
				winIndices = append(winIndices, idx)
			} else {
				break
			}
		}

		// Count distinct players in this sync window.
		winPlayers := map[string]bool{}
		for _, idx := range winIndices {
			winPlayers[castTimes[idx].PlayerName] = true
		}

		if len(winPlayers) >= minPlayersForSync {
			// Valid sync window — mark all casts in it as synced.
			syncStart := castTimes[winIndices[0]].CastTime
			syncEnd := castTimes[winIndices[len(winIndices)-1]].CastTime

			var players []string
			for _, idx := range winIndices {
				syncedIndices[idx] = true
				players = append(players, castTimes[idx].PlayerName)
			}

			result.SyncWindows = append(result.SyncWindows, SyncGroup{
				StartTime:   syncStart,
				EndTime:     syncEnd,
				Players:     players,
				Diff:        syncEnd - syncStart,
				CastIndices: winIndices,
			})
		}
		// All casts in the round that are not in a valid sync window stay unsynced.

		start = end
	}

	// Sort sync windows by start time.
	sort.Slice(result.SyncWindows, func(i, j int) bool {
		return result.SyncWindows[i].StartTime < result.SyncWindows[j].StartTime
	})

	// Collect casts not in any valid sync window.
	for i, ct := range castTimes {
		if !syncedIndices[i] {
			result.OutOfSyncCasts = append(result.OutOfSyncCasts, ct)
		}
	}

	result.Synchronized = len(result.OutOfSyncCasts) == 0

	// Timing statistics: max and avg spread across ordinal groups.
	if len(result.SyncWindows) > 0 {
		var totalDiff int
		for _, g := range result.SyncWindows {
			if g.Diff > result.MaxTimingDiff {
				result.MaxTimingDiff = g.Diff
			}
			totalDiff += g.Diff
		}
		result.AvgTimingDiff = float64(totalDiff) / float64(len(result.SyncWindows))
	}

	return result
}
