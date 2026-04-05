import {
  ScatterChart, Scatter, XAxis, YAxis, CartesianGrid,
  Tooltip, ReferenceArea, ResponsiveContainer, Cell,
} from 'recharts'
import type { WellSyncAnalysis } from '../types'
import { fmtSec } from '../utils'
import Section from './Section'
import styles from './WellTimingSection.module.css'

interface Props {
  wellTiming: Record<string, WellSyncAnalysis>
  durationMs: number
}

interface CastPoint {
  x: number   // cast time in seconds
  y: number   // player index
  oos: boolean
  label: string
}

export default function WellTimingSection({ wellTiming, durationMs }: Props) {
  const entries = Object.values(wellTiming ?? {})
  if (entries.length === 0) {
    return (
      <Section title="Well Cast Timing" subtitle="No well skills were cast in this fight.">
        <p className={styles.empty}>No well skills detected</p>
      </Section>
    )
  }

  return (
    <>
      {entries.map(analysis => (
        <WellChart
          key={analysis.SkillID}
          analysis={analysis}
          durationMs={durationMs}
        />
      ))}
    </>
  )
}

function WellChart({ analysis, durationMs }: { analysis: WellSyncAnalysis; durationMs: number }) {
  const allCasts  = analysis.AllCastTimes  ?? []
  const syncWins  = analysis.SyncWindows   ?? []
  const oosCasts  = new Set(
    (analysis.OutOfSyncCasts ?? []).map(c => `${c.PlayerName}|${c.CastTime}`)
  )

  // Build ordered player list
  const playerOrder: string[] = []
  const seen = new Set<string>()
  for (const ct of allCasts) {
    if (!seen.has(ct.PlayerName)) { playerOrder.push(ct.PlayerName); seen.add(ct.PlayerName) }
  }

  const playerIndex = Object.fromEntries(playerOrder.map((p, i) => [p, i]))

  const points: CastPoint[] = allCasts.map(ct => ({
    x: ct.CastTime / 1000,
    y: playerIndex[ct.PlayerName],
    oos: oosCasts.has(`${ct.PlayerName}|${ct.CastTime}`),
    label: `${ct.PlayerName} · ${fmtSec(ct.CastTime)}`,
  }))

  const maxSec = durationMs / 1000
  const syncedPoints = points.filter(p => !p.oos)
  const oosPoints    = points.filter(p =>  p.oos)

  const isSynced = analysis.Synchronized
  const oosCastCount = (analysis.OutOfSyncCasts ?? []).length

  return (
    <Section
      title={analysis.SkillName}
      subtitle={`${allCasts.length} casts · ${syncWins.length} sync group${syncWins.length !== 1 ? 's' : ''}${oosCastCount > 0 ? ` · ${oosCastCount} out-of-sync` : ''}`}
    >
      <div className={styles.statusRow}>
        <span className={isSynced ? styles.badgeGreen : styles.badgeRed}>
          {isSynced ? '✓ SYNCED' : '✗ OUT OF SYNC'}
        </span>
        <span className={styles.muted}>
          Max diff: {fmtSec(analysis.MaxTimingDiff)} · Avg diff: {fmtSec(analysis.AvgTimingDiff)}
        </span>
      </div>

      <div className={styles.chartWrap}>
        <ResponsiveContainer width="100%" height={Math.max(120, playerOrder.length * 38 + 40)}>
          <ScatterChart margin={{ top: 8, right: 16, bottom: 0, left: 0 }}>
            <CartesianGrid stroke="#2c3050" strokeDasharray="3 5" />
            <XAxis
              type="number"
              dataKey="x"
              domain={[0, maxSec]}
              tickFormatter={v => `${v}s`}
              tick={{ fill: '#7b839a', fontSize: 10 }}
              tickLine={false}
              axisLine={{ stroke: '#2c3050' }}
            />
            <YAxis
              type="number"
              dataKey="y"
              domain={[-0.5, playerOrder.length - 0.5]}
              ticks={playerOrder.map((_, i) => i)}
              tickFormatter={i => playerOrder[i] ?? ''}
              tick={{ fill: '#9aa0b8', fontSize: 11 }}
              tickLine={false}
              axisLine={false}
              width={140}
            />
            <Tooltip
              cursor={{ stroke: '#6366f1', strokeWidth: 1 }}
              content={({ payload }) => {
                if (!payload?.length) return null
                const d = payload[0].payload as CastPoint
                return (
                  <div className={styles.tooltip}>
                    <div style={{ fontWeight: 600 }}>{playerOrder[d.y]}</div>
                    <div>Cast at {fmtSec(d.x * 1000)}</div>
                    {d.oos && <div style={{ color: 'var(--red)' }}>Out of sync</div>}
                  </div>
                )
              }}
            />

            {/* Sync window bands */}
            {syncWins.map((g, i) => (
              <ReferenceArea
                key={i}
                x1={g.StartTime / 1000}
                x2={g.EndTime / 1000}
                fill="rgba(99,102,241,0.12)"
                stroke="rgba(99,102,241,0.4)"
                strokeDasharray="4 3"
                strokeWidth={1}
                ifOverflow="visible"
              />
            ))}

            {/* Synced casts */}
            <Scatter name="Synced" data={syncedPoints} shape={<CastDot />}>
              {syncedPoints.map((_, i) => <Cell key={i} fill="#22c55e" />)}
            </Scatter>

            {/* Out-of-sync casts */}
            <Scatter name="Out of sync" data={oosPoints} shape={<CastDot oos />}>
              {oosPoints.map((_, i) => <Cell key={i} fill="#ef4444" />)}
            </Scatter>
          </ScatterChart>
        </ResponsiveContainer>
      </div>

      <div className={styles.legend}>
        <span><span className={styles.dotGreen} /> Synced cast</span>
        <span><span className={styles.dotRed} /> Out-of-sync cast</span>
        <span><span className={styles.bandSwatch} /> Sync window</span>
      </div>
    </Section>
  )
}

function CastDot({ cx = 0, cy = 0, oos = false }: { cx?: number; cy?: number; oos?: boolean }) {
  const color = oos ? '#ef4444' : '#22c55e'
  return (
    <g>
      <circle cx={cx} cy={cy} r={9} fill={color} fillOpacity={0.2} />
      <circle cx={cx} cy={cy} r={5} fill={color} />
    </g>
  )
}
