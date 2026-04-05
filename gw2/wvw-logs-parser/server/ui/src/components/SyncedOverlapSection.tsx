import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Cell,
} from 'recharts'
import type { SyncedSpikeWell } from '../types'
import { fmtSec, fmtDmg, fmtDps } from '../utils'
import Section from './Section'
import styles from './SyncedOverlapSection.module.css'

interface Props {
  syncedOverlap: SyncedSpikeWell[]
}

export default function SyncedOverlapSection({ syncedOverlap }: Props) {
  const overlaps = syncedOverlap ?? []

  if (overlaps.length === 0) {
    return (
      <Section
        title="Synced Spikes + Wells"
        subtitle="No damage spikes overlapped well sync windows."
      >
        <p className={styles.empty}>No overlapping spike + well sync windows detected</p>
      </Section>
    )
  }

  const maxDmg = Math.max(...overlaps.map(o => o.TotalDamage))

  return (
    <Section
      title="Synced Spikes + Wells"
      subtitle="Time windows where a well sync group and damage spikes coincided (±5s window)."
    >
      {/* Summary bar chart */}
      <div className={styles.chartWrap}>
        <ResponsiveContainer width="100%" height={Math.max(80, overlaps.length * 36 + 30)}>
          <BarChart
            data={overlaps.map(o => ({
              window: `${fmtSec(o.SyncWindowStart)}–${fmtSec(o.SyncWindowEnd)}`,
              damage: o.TotalDamage,
            }))}
            layout="vertical"
            margin={{ top: 4, right: 50, bottom: 4, left: 90 }}
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
              dataKey="window"
              width={90}
              tick={{ fill: '#9aa0b8', fontSize: 11 }}
              tickLine={false}
              axisLine={false}
            />
            <Tooltip
              cursor={{ fill: 'rgba(99,102,241,0.08)' }}
              formatter={(v) => [fmtDmg(Number(v)), 'Total Damage']}
              contentStyle={{
                background: 'var(--surface2)',
                border: '1px solid var(--border)',
                borderRadius: 6,
                fontSize: 12,
              }}
              labelStyle={{ color: 'var(--accent2)', fontWeight: 600 }}
            />
            <Bar dataKey="damage" radius={[0, 3, 3, 0]}>
              {overlaps.map((_, i) => (
                <Cell key={i} fill="url(#overlapGrad)" />
              ))}
            </Bar>
            <defs>
              <linearGradient id="overlapGrad" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%"   stopColor="#6366f1" />
                <stop offset="100%" stopColor="#22d3ee" />
              </linearGradient>
            </defs>
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Detail cards */}
      <div className={styles.cards}>
        {overlaps.map((overlap, i) => {
          const maxSpikeDmg = overlap.PlayerSpikes.length
            ? Math.max(...overlap.PlayerSpikes.map(ps => ps.Damage))
            : 1
          return (
            <div key={i} className={styles.card}>
              <div className={styles.cardHead}>
                <span className={styles.windowLabel}>
                  {fmtSec(overlap.SyncWindowStart)} – {fmtSec(overlap.SyncWindowEnd)}
                </span>
                <span className={styles.badgeWell}>WELL SYNC</span>
                <span className={styles.badgeSpike}>SPIKE</span>
                <span className={styles.totalDmg}>
                  Total: <strong>{fmtDmg(overlap.TotalDamage)}</strong>
                </span>
              </div>

              <div className={styles.wellPlayers}>
                Wells cast by:{' '}
                <span className={styles.wellNames}>
                  {(overlap.WellPlayers ?? []).join(', ')}
                </span>
              </div>

              {/* Total damage bar relative to max overlap */}
              <div className={styles.barBg}>
                <div
                  className={styles.barFill}
                  style={{ width: `${maxDmg > 0 ? (overlap.TotalDamage / maxDmg) * 100 : 0}%` }}
                />
              </div>

              {/* Per-player spike breakdown */}
              {overlap.PlayerSpikes.length > 0 && (
                <table className={styles.spikeTable}>
                  <tbody>
                    {overlap.PlayerSpikes.map((ps, j) => (
                      <tr key={j}>
                        <td className={styles.rank}>{j + 1}</td>
                        <td className={styles.spikeName}>{ps.PlayerName}</td>
                        <td className={styles.spikeBar}>
                          <div className={styles.barBg}>
                            <div
                              className={styles.barFillCyan}
                              style={{ width: `${(ps.Damage / maxSpikeDmg) * 100}%` }}
                            />
                          </div>
                        </td>
                        <td className={styles.spikeDmg}>{fmtDmg(ps.Damage)}</td>
                        <td className={styles.spikeDps}>{fmtDps(ps.DPS)} dps</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )
        })}
      </div>
    </Section>
  )
}
