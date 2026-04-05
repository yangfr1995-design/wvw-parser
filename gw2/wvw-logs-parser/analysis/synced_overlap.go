package analysis

import "sort"

// SyncedSpikeWell represents a well sync window that coincides with damage spikes.
type SyncedSpikeWell struct {
	SyncWindowStart int                   // ms from fight start (earliest well cast)
	SyncWindowEnd   int                   // ms from fight start (latest well cast)
	WellPlayers     []string              // unique players who cast in this window
	TotalDamage     int                   // sum of damage from all overlapping spikes
	PlayerSpikes    []PlayerSpikeInWindow // sorted descending by damage
}

// PlayerSpikeInWindow is a spike that overlapped a well sync window.
type PlayerSpikeInWindow struct {
	PlayerName string
	Damage     int
	DPS        float64
}

// overlapPaddingMS: how far (in ms) outside the sync window we still count spikes.
const overlapPaddingMS = 5000

// FindSyncedSpikeWells finds every qualifying well sync window whose time range (plus
// padding) overlaps with at least one player's damage spike.
func FindSyncedSpikeWells(wellAnalysis map[int]*WellSyncAnalysis, spikes map[string][]SpikeDamage) []SyncedSpikeWell {
	var results []SyncedSpikeWell

	for _, wellSync := range wellAnalysis {
		for _, syncGroup := range wellSync.SyncWindows {
			lookStart := syncGroup.StartTime - overlapPaddingMS
			lookEnd := syncGroup.EndTime + overlapPaddingMS

			overlap := SyncedSpikeWell{
				SyncWindowStart: syncGroup.StartTime,
				SyncWindowEnd:   syncGroup.EndTime,
			}

			// deduplicate player names in the well sync group
			seen := map[string]bool{}
			for _, p := range syncGroup.Players {
				if !seen[p] {
					overlap.WellPlayers = append(overlap.WellPlayers, p)
					seen[p] = true
				}
			}

			// collect every spike that overlaps the padded window
			for playerName, playerSpikes := range spikes {
				for _, spike := range playerSpikes {
					if spike.StartTime <= lookEnd && spike.EndTime >= lookStart {
						overlap.TotalDamage += spike.DamageAmount
						overlap.PlayerSpikes = append(overlap.PlayerSpikes, PlayerSpikeInWindow{
							PlayerName: playerName,
							Damage:     spike.DamageAmount,
							DPS:        spike.DPS,
						})
					}
				}
			}

			if len(overlap.PlayerSpikes) > 0 {
				sort.Slice(overlap.PlayerSpikes, func(i, j int) bool {
					return overlap.PlayerSpikes[i].Damage > overlap.PlayerSpikes[j].Damage
				})
				results = append(results, overlap)
			}
		}
	}

	// sort by earliest well cast time
	sort.Slice(results, func(i, j int) bool {
		return results[i].SyncWindowStart < results[j].SyncWindowStart
	})

	return results
}
