import { useState, useEffect, useRef } from 'react'
import { fetchFights, fetchFightDetail } from './api'
import type { FightMeta, FightDetail } from './types'
import Sidebar from './components/Sidebar'
import FightView from './components/FightView'
import LogsPage from './components/LogsPage'
import styles from './App.module.css'

type View = 'logs' | 'fight'

export default function App() {
  const [fights, setFights]         = useState<FightMeta[]>([])
  const [view, setView]             = useState<View>('logs')
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [detail, setDetail]         = useState<FightDetail | null>(null)
  const [loading, setLoading]       = useState(false)
  const [error, setError]           = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Poll /api/fights every 3s to pick up newly processed logs.
  useEffect(() => {
    loadFights()
    pollRef.current = setInterval(loadFights, 3000)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [])

  async function loadFights() {
    try {
      const data = await fetchFights()
      setFights(data)
    } catch (e) {
      // non-fatal; server may still be starting
    }
  }

  async function selectFight(id: string) {
    setSelectedId(id)
    setView('fight')
    setDetail(null)
    setLoading(true)
    setError(null)
    try {
      const d = await fetchFightDetail(id)
      setDetail(d)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.layout}>
      <Sidebar
        fights={fights}
        selectedId={selectedId}
        onSelect={selectFight}
        onImport={() => setView('logs')}
      />
      <main className={styles.main}>
        {view === 'logs' && (
          <LogsPage onDone={() => setView('fight')} />
        )}
        {view === 'fight' && !selectedId && (
          <div className={styles.placeholder}>
            Select a fight from the sidebar to view analysis
          </div>
        )}
        {view === 'fight' && loading && <div className={styles.loading}>Loading…</div>}
        {view === 'fight' && error && <div className={styles.error}>{error}</div>}
        {view === 'fight' && detail && !loading && <FightView data={detail} />}
      </main>
    </div>
  )
}
