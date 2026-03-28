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

	// SyncThreshold: casts within this millisecond window are considered synchronized
	SyncThresholdMS = 3000 // 500ms window
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

// analyzeSync evaluates how synchronized casts were
func analyzeSync(skillName string, skillID int, castTimes []WellSkillTiming) *WellSyncAnalysis {
	result := &WellSyncAnalysis{
		SkillName:    skillName,
		SkillID:      skillID,
		AllCastTimes: castTimes,
	}

	if len(castTimes) == 0 {
		result.Synchronized = true // no casts = vacuously true
		return result
	}

	// Count unique players who cast this well across the whole fight
	uniquePlayerSet := map[string]bool{}
	for _, ct := range castTimes {
		uniquePlayerSet[ct.PlayerName] = true
	}
	totalPlayers := len(uniquePlayerSet)

	// Group casts into time windows
	allGroups := groupBySyncWindow(castTimes)

	// A group qualifies as a sync group only if more than half of all
	// well-casting players participated (counted by unique player names)
	syncedIndices := map[int]bool{}
	for _, g := range allGroups {
		uniqueInGroup := map[string]bool{}
		for _, p := range g.Players {
			uniqueInGroup[p] = true
		}
		if len(uniqueInGroup) > totalPlayers/2 {
			result.SyncWindows = append(result.SyncWindows, g)
			for _, idx := range g.CastIndices {
				syncedIndices[idx] = true
			}
		}
	}

	// Collect casts not in any qualifying sync group
	for i, ct := range castTimes {
		if !syncedIndices[i] {
			result.OutOfSyncCasts = append(result.OutOfSyncCasts, ct)
		}
	}

	result.Synchronized = len(result.OutOfSyncCasts) == 0

	// Timing statistics across all casts
	if len(castTimes) > 1 {
		minTime := castTimes[0].CastTime
		maxTime := castTimes[len(castTimes)-1].CastTime
		result.MaxTimingDiff = maxTime - minTime

		var totalDiff int
		for i := 1; i < len(castTimes); i++ {
			totalDiff += castTimes[i].CastTime - castTimes[i-1].CastTime
		}
		result.AvgTimingDiff = float64(totalDiff) / float64(len(castTimes)-1)
	}

	return result
}

// groupBySyncWindow groups casts that occur within the sync threshold
func groupBySyncWindow(castTimes []WellSkillTiming) []SyncGroup {
	if len(castTimes) == 0 {
		return nil
	}

	var groups []SyncGroup
	currentGroup := SyncGroup{
		StartTime:   castTimes[0].CastTime,
		EndTime:     castTimes[0].CastTime,
		Players:     []string{castTimes[0].PlayerName},
		CastIndices: []int{0},
	}

	for i := 1; i < len(castTimes); i++ {
		diff := castTimes[i].CastTime - currentGroup.StartTime
		if diff <= SyncThresholdMS {
			currentGroup.EndTime = castTimes[i].CastTime
			currentGroup.Players = append(currentGroup.Players, castTimes[i].PlayerName)
			currentGroup.Diff = currentGroup.EndTime - currentGroup.StartTime
			currentGroup.CastIndices = append(currentGroup.CastIndices, i)
		} else {
			groups = append(groups, currentGroup)
			currentGroup = SyncGroup{
				StartTime:   castTimes[i].CastTime,
				EndTime:     castTimes[i].CastTime,
				Players:     []string{castTimes[i].PlayerName},
				Diff:        0,
				CastIndices: []int{i},
			}
		}
	}

	groups = append(groups, currentGroup)
	return groups
}
