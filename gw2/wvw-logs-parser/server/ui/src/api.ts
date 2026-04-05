import type { FightMeta, FightDetail, LogEntry } from './types'

export async function fetchLogs(): Promise<LogEntry[]> {
  const res = await fetch('/api/logs')
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return (await res.json()) ?? []
}

export async function processLogs(names: string[]): Promise<void> {
  const res = await fetch('/api/logs/process', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ names }),
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
}

export async function fetchFights(): Promise<FightMeta[]> {
  const res = await fetch('/api/fights')
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return (await res.json()) ?? []
}

export async function fetchFightDetail(id: string): Promise<FightDetail> {
  const res = await fetch(`/api/fights/${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}
