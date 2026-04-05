import { useState, useEffect, useRef } from 'react'
import type { LogEntry } from '../types'
import { fetchLogs, processLogs } from '../api'
import styles from './LogsPage.module.css'

interface Props {
  onDone: () => void // navigate to fights view when user clicks "View Fights"
}

export default function LogsPage({ onDone }: Props) {
  const [entries, setEntries]   = useState<LogEntry[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [error, setError]       = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  function isActive(entries: LogEntry[]) {
    return entries.some(e => e.status === 'processing')
  }

  useEffect(() => {
    load()
    pollRef.current = setInterval(load, 2000)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [])

  async function load() {
    try {
      const data = await fetchLogs()
      setEntries(data)
    } catch (e) {
      setError(String(e))
    }
  }

  function toggleAll(checked: boolean) {
    if (checked) {
      setSelected(new Set(entries.filter(e => e.status === 'pending').map(e => e.name)))
    } else {
      setSelected(new Set())
    }
  }

  function toggle(name: string) {
    setSelected(prev => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  async function handleProcess(names: string[]) {
    if (names.length === 0) return
    setError(null)
    try {
      await processLogs(names)
      setSelected(new Set())
      load()
    } catch (e) {
      setError(String(e))
    }
  }

  const pending    = entries.filter(e => e.status === 'pending')
  const processing = entries.filter(e => e.status === 'processing')
  const done       = entries.filter(e => e.status === 'done')
  const errored    = entries.filter(e => e.status === 'error')
  const allPendingSelected = pending.length > 0 && pending.every(e => selected.has(e.name))

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h2 className={styles.title}>Import Logs</h2>
          <p className={styles.subtitle}>
            {entries.length} log{entries.length !== 1 ? 's' : ''} found
            {processing.length > 0 && ` · ${processing.length} processing…`}
            {done.length > 0 && ` · ${done.length} ready`}
          </p>
        </div>
        <div className={styles.actions}>
          <button
            className={styles.btnSecondary}
            disabled={selected.size === 0 || isActive(entries)}
            onClick={() => handleProcess([...selected])}
          >
            Process Selected ({selected.size})
          </button>
          <button
            className={styles.btnPrimary}
            disabled={pending.length === 0 || isActive(entries)}
            onClick={() => handleProcess(pending.map(e => e.name))}
          >
            Process All Pending ({pending.length})
          </button>
          {done.length > 0 && (
            <button className={styles.btnGreen} onClick={onDone}>
              View Fights →
            </button>
          )}
        </div>
      </div>

      {error && <div className={styles.error}>{error}</div>}

      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.checkCol}>
              <input
                type="checkbox"
                checked={allPendingSelected}
                onChange={e => toggleAll(e.target.checked)}
                disabled={pending.length === 0}
              />
            </th>
            <th>Log File</th>
            <th className={styles.statusCol}>Status</th>
          </tr>
        </thead>
        <tbody>
          {entries.map(e => (
            <tr key={e.name} className={e.status === 'done' ? styles.rowDone : ''}>
              <td className={styles.checkCol}>
                {e.status === 'pending' && (
                  <input
                    type="checkbox"
                    checked={selected.has(e.name)}
                    onChange={() => toggle(e.name)}
                  />
                )}
              </td>
              <td className={styles.filename}>{e.name}</td>
              <td>
                <StatusBadge entry={e} />
                {e.error && <span className={styles.errMsg}>{e.error}</span>}
              </td>
            </tr>
          ))}
          {entries.length === 0 && (
            <tr>
              <td colSpan={3} className={styles.empty}>
                No .zevtc log files found in the configured log folder.
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {processing.length > 0 && (
        <div className={styles.progressBar}>
          <div
            className={styles.progressFill}
            style={{ width: `${(done.length / (done.length + processing.length + pending.length + errored.length)) * 100}%` }}
          />
        </div>
      )}
    </div>
  )
}

function StatusBadge({ entry }: { entry: LogEntry }) {
  switch (entry.status) {
    case 'pending':    return <span className={styles.badgePending}>Pending</span>
    case 'processing': return <span className={styles.badgeProcessing}>⟳ Processing…</span>
    case 'done':       return <span className={styles.badgeDone}>✓ Done</span>
    case 'error':      return <span className={styles.badgeError}>✗ Error</span>
  }
}
