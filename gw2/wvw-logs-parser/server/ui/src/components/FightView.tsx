import type { FightDetail } from '../types'
import { fmtMs, fmtDmg } from '../utils'
import StatCard from './StatCard'
import WellTimingSection from './WellTimingSection'
import OutOfSyncSection from './OutOfSyncSection'
import SpikeDamageSection from './SpikeDamageSection'
import SyncedOverlapSection from './SyncedOverlapSection'
import styles from './FightView.module.css'

interface Props {
  data: FightDetail
}

export default function FightView({ data }: Props) {
  const { meta, wellTiming, spikeDamage, syncedOverlap } = data

  // Count out-of-sync casts across all well skills
  let totalOos = 0
  const oosByPlayer: Record<string, number[]> = {}
  for (const analysis of Object.values(wellTiming ?? {})) {
    for (const ct of analysis.OutOfSyncCasts ?? []) {
      totalOos++
      if (!oosByPlayer[ct.PlayerName]) oosByPlayer[ct.PlayerName] = []
      oosByPlayer[ct.PlayerName].push(ct.CastTime)
    }
  }

  let totalWellCasts = 0
  for (const a of Object.values(wellTiming ?? {})) {
    totalWellCasts += (a.AllCastTimes ?? []).length
  }

  let totalSpikes = 0
  for (const spks of Object.values(spikeDamage ?? {})) {
    totalSpikes += spks.length
  }

  const totalSpikeDmg = (syncedOverlap ?? []).reduce((s, o) => s + o.TotalDamage, 0)

  return (
    <div className={styles.root}>
      <h1 className={styles.heading}>{meta.name || meta.id}</h1>
      <p className={styles.subhead}>
        {fmtMs(meta.durationMs)} · {meta.playerCount} players
      </p>

      <div className={styles.stats}>
        <StatCard label="Players" value={meta.playerCount} />
        <StatCard label="Well Casts" value={totalWellCasts} />
        <StatCard
          label="Out-of-Sync"
          value={totalOos}
          variant={totalOos > 0 ? 'danger' : 'success'}
        />
        <StatCard label="Spike Events" value={totalSpikes} />
        <StatCard label="Sync+Spike Overlaps" value={(syncedOverlap ?? []).length} />
        {totalSpikeDmg > 0 && (
          <StatCard label="Overlap Total Dmg" value={fmtDmg(totalSpikeDmg)} />
        )}
      </div>

      <WellTimingSection wellTiming={wellTiming} durationMs={meta.durationMs} />
      <OutOfSyncSection oosByPlayer={oosByPlayer} />
      <SpikeDamageSection spikeDamage={spikeDamage} players={data.players} wellTiming={wellTiming} durationMs={meta.durationMs} />
      <SyncedOverlapSection syncedOverlap={syncedOverlap} />
    </div>
  )
}
