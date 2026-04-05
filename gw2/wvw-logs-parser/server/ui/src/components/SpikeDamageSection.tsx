import { useState } from 'react'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip,
  Cell, ResponsiveContainer, ScatterChart, Scatter,
  ReferenceArea,
} from 'recharts'
import type { SpikeDamage, PlayerSummary, WellSyncAnalysis } from '../types'
import { fmtSec, fmtDmg, fmtDps, profColor } from '../utils'
import Section from './Section'
import styles from './SpikeDamageSection.module.css'

interface Props {
  spikeDamage: Record<string, SpikeDamage[]>
  players: PlayerSummary[]
  wellTiming?: Record<string, WellSyncAnalysis>
  durationMs?: number
}

interface FlatSpike {
  label: string
  player: string
  profession: string
  start: number
  end: number
  damage: number
  dps: number
  peak: number
}

type SortKey = 'player' | 'start' | 'damage' | 'dps' | 'peak'

export default function SpikeDamageSection({ spikeDamage, players, wellTiming, durationMs }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('damage')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')

  function handleSort(key: SortKey) {
    if (key === sortKey) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir(key === 'player' ? 'asc' : 'desc')
    }
  }

  function sortIndicator(key: SortKey) {
    if (key !== sortKey) return <span className={styles.sortNeutral}>⇅</span>
    return <span className={styles.sortActive}>{sortDir === 'asc' ? '▲' : '▼'}</span>
  }

  const profMap: Record<string, string> = {}
  for (const p of players ?? []) profMap[p.Name] = p.Profession

  // Flatten all spikes
  const flat: FlatSpike[] = []
  for (const [name, spks] of Object.entries(spikeDamage ?? {})) {
    for (const s of spks) {
      flat.push({
        label: `${name} ${fmtSec(s.StartTime)}`,
        player: name,
        profession: profMap[name] ?? '',
        start: s.StartTime,
        end: s.EndTime,
        damage: s.DamageAmount,
        dps: s.DPS,
        peak: s.PeakDamageInS,
      })
    }
  }

  flat.sort((a, b) => {
    let cmp = 0
    if (sortKey === 'player') cmp = a.player.localeCompare(b.player)
    else if (sortKey === 'start')  cmp = a.start  - b.start
    else if (sortKey === 'damage') cmp = a.damage - b.damage
    else if (sortKey === 'dps')    cmp = a.dps    - b.dps
    else if (sortKey === 'peak')   cmp = a.peak   - b.peak
    return sortDir === 'asc' ? cmp : -cmp
  })

  if (flat.length === 0) {
    return (
      <Section title="Spike Damage" subtitle="No damage spikes detected.">
        <p className={styles.empty}>No spike damage detected</p>
      </Section>
    )
  }

  // Top 20 for the chart (readability)
  const chartData = flat.slice(0, 20)

  return (
    <Section
      title="Spike Damage"
      subtitle="Periods where a player's damage significantly exceeded their average bracket."
    >
      <div className={styles.chartWrap}>
        <ResponsiveContainer width="100%" height={280}>
          <BarChart
            data={chartData}
            layout="vertical"
            margin={{ top: 0, right: 40, bottom: 0, left: 130 }}
          >
            <CartesianGrid horizontal={false} stroke="#2c3050" strokeDasharray="3 5" />
            <XAxis
              type="number"
              dataKey="damage"
              tickFormatter={fmtDmg}
              tick={{ fill: '#7b839a', fontSize: 10 }}
              tickLine={false}
              axisLine={{ stroke: '#2c3050' }}
            />
            <YAxis
              type="category"
              dataKey="label"
              tick={{ fill: '#9aa0b8', fontSize: 11 }}
              tickLine={false}
              axisLine={false}
              width={130}
            />
            <Tooltip
              cursor={{ fill: 'rgba(99,102,241,0.08)' }}
              content={({ payload }) => {
                if (!payload?.length) return null
                const d = payload[0].payload as FlatSpike
                return (
                  <div className={styles.tooltip}>
                    <div style={{ fontWeight: 600, color: profColor(d.profession) || 'var(--text)' }}>
                      {d.player}
                      {d.profession && <span style={{ color: 'var(--muted)', marginLeft: 6, fontWeight: 400 }}>{d.profession}</span>}
                    </div>
                    <div>Window: {fmtSec(d.start)} – {fmtSec(d.end)}</div>
                    <div>Damage: <strong>{fmtDmg(d.damage)}</strong></div>
                    <div>Avg DPS: <strong>{fmtDps(d.dps)}</strong></div>
                    <div style={{ color: 'var(--muted)' }}>Peak: {fmtDmg(d.peak)}/s</div>
                  </div>
                )
              }}
            />
            <Bar dataKey="damage" radius={[0, 3, 3, 0]}>
              {chartData.map((entry, i) => (
                <Cell key={i} fill={profColor(entry.profession)} fillOpacity={0.85} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Timeline chart: spike windows + well sync overlay */}
      {durationMs != null && (() => {
        // Collect unique players in spike data
        const playerOrder: string[] = []
        const seenP = new Set<string>()
        for (const s of flat) {
          if (!seenP.has(s.player)) { playerOrder.push(s.player); seenP.add(s.player) }
        }
        const playerIdx = Object.fromEntries(playerOrder.map((p, i) => [p, i]))
        const maxSec = durationMs / 1000

        // Build scatter points (mid-point of each spike window)
        const timelinePoints = flat.map(s => ({
          x: (s.start + s.end) / 2 / 1000,
          y: playerIdx[s.player],
          startSec: s.start / 1000,
          endSec: s.end / 1000,
          player: s.player,
          profession: s.profession,
          damage: s.damage,
          dps: s.dps,
          start: s.start,
          end: s.end,
        }))

        // Collect all well sync windows
        const syncWindows: { start: number; end: number; skill: string }[] = []
        for (const analysis of Object.values(wellTiming ?? {})) {
          for (const g of analysis.SyncWindows ?? []) {
            syncWindows.push({
              start: g.StartTime / 1000,
              end: g.EndTime / 1000,
              skill: analysis.SkillName,
            })
          }
        }

        return (
          <>
            <h4 className={styles.timelineHeading}>Spike Timeline vs Well Sync Windows</h4>
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
                      const d = payload[0].payload as (typeof timelinePoints)[0]
                      return (
                        <div className={styles.tooltip}>
                          <div style={{ fontWeight: 600, color: profColor(d.profession) || 'var(--text)' }}>
                            {d.player}
                            {d.profession && <span style={{ color: 'var(--muted)', marginLeft: 6, fontWeight: 400 }}>{d.profession}</span>}
                          </div>
                          <div>Window: {fmtSec(d.start)} – {fmtSec(d.end)}</div>
                          <div>Damage: <strong>{fmtDmg(d.damage)}</strong></div>
                          <div>Avg DPS: <strong>{fmtDps(d.dps)}</strong></div>
                        </div>
                      )
                    }}
                  />

                  {/* Well sync window bands */}
                  {syncWindows.map((w, i) => (
                    <ReferenceArea
                      key={`well-${i}`}
                      x1={w.start}
                      x2={w.end}
                      fill="rgba(99,102,241,0.12)"
                      stroke="rgba(99,102,241,0.4)"
                      strokeDasharray="4 3"
                      strokeWidth={1}
                      ifOverflow="visible"
                    />
                  ))}

                  {/* Spike windows as reference areas per player row */}
                  {timelinePoints.map((s, i) => (
                    <ReferenceArea
                      key={`spike-${i}`}
                      x1={s.startSec}
                      x2={s.endSec}
                      y1={s.y - 0.3}
                      y2={s.y + 0.3}
                      fill={profColor(s.profession)}
                      fillOpacity={0.25}
                      stroke={profColor(s.profession)}
                      strokeOpacity={0.6}
                      strokeWidth={1}
                      radius={3}
                      ifOverflow="visible"
                    />
                  ))}

                  {/* Scatter dots at midpoint of each spike */}
                  <Scatter name="Spikes" data={timelinePoints}>
                    {timelinePoints.map((entry, i) => (
                      <Cell key={i} fill={profColor(entry.profession)} />
                    ))}
                  </Scatter>
                </ScatterChart>
              </ResponsiveContainer>
            </div>
            <div className={styles.timelineLegend}>
              <span><span className={styles.dotSpike} /> Spike window</span>
              <span><span className={styles.bandSwatch} /> Well sync window</span>
            </div>
          </>
        )
      })()}

      {/* Full table below chart */}
      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.sortable} onClick={() => handleSort('player')}>Player {sortIndicator('player')}</th>
            <th className={styles.sortable} onClick={() => handleSort('start')}>Window {sortIndicator('start')}</th>
            <th className={styles.sortable} onClick={() => handleSort('damage')}>Damage {sortIndicator('damage')}</th>
            <th className={styles.sortable} onClick={() => handleSort('dps')}>Avg DPS {sortIndicator('dps')}</th>
            <th className={styles.sortable} onClick={() => handleSort('peak')}>Peak/s {sortIndicator('peak')}</th>
          </tr>
        </thead>
        <tbody>
          {flat.map((s, i) => (
            <tr key={i}>
              <td>
                <span style={{ color: profColor(s.profession), fontWeight: 500 }}>{s.player}</span>
                {s.profession && (
                  <span className={styles.prof}>{s.profession}</span>
                )}
              </td>
              <td className={styles.mono}>{fmtSec(s.start)} – {fmtSec(s.end)}</td>
              <td className={styles.dmg}>{fmtDmg(s.damage)}</td>
              <td className={styles.dps}>{fmtDps(s.dps)}</td>
              <td className={styles.peak}>{fmtDmg(s.peak)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Section>
  )
}
