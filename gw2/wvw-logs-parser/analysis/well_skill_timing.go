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
	SkillName     string
	SkillID       int
	AllCastTimes  []WellSkillTiming
	Synchronized  bool        // true if all casts within sync window
	MaxTimingDiff int         // maximum millisecond difference between casts
	AvgTimingDiff float64     // average difference
	SyncWindows   []SyncGroup // groups of players who cast in sync
}

// SyncGroup represents players who cast within a small time window
type SyncGroup struct {
	StartTime int // earliest cast time in group
	EndTime   int // latest cast time in group
	Players   []string
	Diff      int // end - start
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

	// Group casts by sync windows
	result.SyncWindows = groupBySyncWindow(castTimes)

	// Calculate timing statistics
	if len(castTimes) > 1 {
		minTime := castTimes[0].CastTime
		maxTime := castTimes[len(castTimes)-1].CastTime
		result.MaxTimingDiff = maxTime - minTime

		// Calculate average difference between consecutive casts
		var totalDiff int
		for i := 1; i < len(castTimes); i++ {
			totalDiff += castTimes[i].CastTime - castTimes[i-1].CastTime
		}
		result.AvgTimingDiff = float64(totalDiff) / float64(len(castTimes)-1)
	}

	// Check if well synchronized (all casts within threshold or multiple small groups)
	if len(result.SyncWindows) == 1 && result.SyncWindows[0].Diff <= SyncThresholdMS {
		result.Synchronized = true
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
		StartTime: castTimes[0].CastTime,
		EndTime:   castTimes[0].CastTime,
		Players:   []string{castTimes[0].PlayerName},
	}

	for i := 1; i < len(castTimes); i++ {
		diff := castTimes[i].CastTime - currentGroup.StartTime
		if diff <= SyncThresholdMS {
			// Add to current group
			currentGroup.EndTime = castTimes[i].CastTime
			currentGroup.Players = append(currentGroup.Players, castTimes[i].PlayerName)
			currentGroup.Diff = currentGroup.EndTime - currentGroup.StartTime
		} else {
			// Start new group
			groups = append(groups, currentGroup)
			currentGroup = SyncGroup{
				StartTime: castTimes[i].CastTime,
				EndTime:   castTimes[i].CastTime,
				Players:   []string{castTimes[i].PlayerName},
				Diff:      0,
			}
		}
	}

	// Don't forget the last group
	groups = append(groups, currentGroup)
	return groups
}
