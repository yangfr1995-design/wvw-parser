import type { FightMeta } from '../types'
import { fmtMs } from '../utils'
import styles from './Sidebar.module.css'

interface Props {
  fights: FightMeta[]
  selectedId: string | null
  onSelect: (id: string) => void
  onImport: () => void
}

export default function Sidebar({ fights, selectedId, onSelect, onImport }: Props) {
  return (
    <aside className={styles.sidebar}>
      <div className={styles.head}>
        <h1>⚔ WvW Analyzer</h1>
        <p>{fights.length} fight{fights.length !== 1 ? 's' : ''}</p>
      </div>
      <div className={styles.importRow}>
        <button className={styles.importBtn} onClick={onImport}>
          + Import Logs
        </button>
      </div>
      <div className={styles.list}>
        {fights.length === 0 && (
          <p className={styles.empty}>No fights yet — import logs above</p>
        )}
        {fights.map(f => (
          <button
            key={f.id}
            className={`${styles.item} ${f.id === selectedId ? styles.active : ''}`}
            onClick={() => onSelect(f.id)}
          >
            <span className={styles.name}>{f.name || f.id}</span>
            <span className={styles.meta}>
              {fmtMs(f.durationMs)} · {f.playerCount} players
            </span>
          </button>
        ))}
      </div>
    </aside>
  )
}
