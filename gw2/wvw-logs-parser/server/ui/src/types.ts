export type LogStatus = 'pending' | 'processing' | 'done' | 'error'

export interface LogEntry {
  name: string
  status: LogStatus
  error?: string
}

export interface FightMeta {
  id: string
  name: string
  durationMs: number
  playerCount: number
}

export interface PlayerSummary {
  Name: string
  Profession: string
  Group: number
  DPS: number
  Damage: number
  Downs: number
  Deaths: number
}

export interface WellSkillTiming {
  PlayerName: string
  SkillID: number
  SkillName: string
  CastTime: number
  Duration: number
  Profession: string
  Group: number
}

export interface SyncGroup {
  StartTime: number
  EndTime: number
  Players: string[]
  Diff: number
  CastIndices: number[]
}

export interface WellSyncAnalysis {
  SkillName: string
  SkillID: number
  AllCastTimes: WellSkillTiming[]
  Synchronized: boolean
  MaxTimingDiff: number
  AvgTimingDiff: number
  SyncWindows: SyncGroup[]
  OutOfSyncCasts: WellSkillTiming[]
}

export interface SpikeDamage {
  PlayerName: string
  StartTime: number
  EndTime: number
  Duration: number
  DamageAmount: number
  DPS: number
  PeakDamageInS: number
}

export interface PlayerSpikeInWindow {
  PlayerName: string
  Damage: number
  DPS: number
}

export interface SyncedSpikeWell {
  SyncWindowStart: number
  SyncWindowEnd: number
  WellPlayers: string[]
  TotalDamage: number
  PlayerSpikes: PlayerSpikeInWindow[]
}

export interface FightDetail {
  meta: FightMeta
  players: PlayerSummary[]
  wellTiming: Record<string, WellSyncAnalysis>
  spikeDamage: Record<string, SpikeDamage[]>
  syncedOverlap: SyncedSpikeWell[]
}
