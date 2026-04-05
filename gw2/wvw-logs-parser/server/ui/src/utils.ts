export function fmtMs(ms: number): string {
  const s = ms / 1000
  return s >= 60
    ? `${Math.floor(s / 60)}m ${Math.round(s % 60)}s`
    : `${s.toFixed(1)}s`
}

export function fmtSec(ms: number): string {
  return `${(ms / 1000).toFixed(1)}s`
}

export function fmtDmg(d: number): string {
  if (d >= 1_000_000) return `${(d / 1_000_000).toFixed(2)}M`
  if (d >= 1_000) return `${(d / 1_000).toFixed(1)}k`
  return String(d)
}

export function fmtDps(d: number): string {
  return Math.round(d).toLocaleString()
}

/** Map GW2 profession name → hex colour */
export const PROF_COLORS: Record<string, string> = {
  Necromancer: '#52a76f',
  Guardian: '#72c1d9',
  Elementalist: '#f68a87',
  Mesmer: '#b679d5',
  Ranger: '#8cdc82',
  Warrior: '#ffe666',
  Engineer: '#d09c59',
  Thief: '#c08f95',
  Revenant: '#d16e5a',
  Reaper: '#52a76f',
  Scourge: '#52a76f',
  Dragonhunter: '#72c1d9',
  Firebrand: '#72c1d9',
  Tempest: '#f68a87',
  Weaver: '#f68a87',
  Chronomancer: '#b679d5',
  Mirage: '#b679d5',
  Soulbeast: '#8cdc82',
  Druid: '#8cdc82',
  Berserker: '#ffe666',
  Spellbreaker: '#ffe666',
  Scrapper: '#d09c59',
  Holosmith: '#d09c59',
  Daredevil: '#c08f95',
  Deadeye: '#c08f95',
  Renegade: '#d16e5a',
  Herald: '#d16e5a',
  Ritualist: '#b679d5',
}

export function profColor(prof: string): string {
  return PROF_COLORS[prof] ?? '#9aa0b8'
}
