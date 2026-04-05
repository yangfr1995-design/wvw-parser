import { useState } from 'react'
import Section from './Section'
import { fmtSec } from '../utils'
import styles from './OutOfSyncSection.module.css'

interface Props {
  oosByPlayer: Record<string, number[]>
}

type SortKey = 'player' | 'count'

export default function OutOfSyncSection({ oosByPlayer }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('count')
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

  const players = Object.keys(oosByPlayer).sort((a, b) => {
    let cmp = 0
    if (sortKey === 'player') cmp = a.localeCompare(b)
    else cmp = oosByPlayer[a].length - oosByPlayer[b].length
    return sortDir === 'asc' ? cmp : -cmp
  })

  if (players.length === 0) {
    return (
      <Section title="Players Out of Sync" subtitle="All players cast wells within sync windows.">
        <p className={styles.allGood}>
          <span className={styles.badgeGreen}>All synced!</span>{' '}
          No out-of-sync casts detected.
        </p>
      </Section>
    )
  }

  return (
    <Section
      title="Players Out of Sync"
      subtitle={`${players.length} player${players.length !== 1 ? 's' : ''} had casts outside sync windows.`}
    >
      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.sortable} onClick={() => handleSort('player')}>Player {sortIndicator('player')}</th>
            <th className={styles.sortable} onClick={() => handleSort('count')}>OOS Casts {sortIndicator('count')}</th>
            <th>Cast Times</th>
          </tr>
        </thead>
        <tbody>
          {players.map(name => (
            <tr key={name}>
              <td className={styles.playerName}>{name}</td>
              <td className={styles.count}>{oosByPlayer[name].length}</td>
              <td>
                {oosByPlayer[name].map((t, i) => (
                  <span key={i} className={styles.timeBadge}>{fmtSec(t)}</span>
                ))}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Section>
  )
}
