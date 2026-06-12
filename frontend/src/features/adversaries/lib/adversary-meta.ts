import { Globe2, Server, User, type LucideIcon } from 'lucide-react'
import type { TFunction } from 'i18next'
import type { Adversary, AdversaryKind, ThreatLevel } from '../types/adversary.types'

export const THREAT: Record<ThreatLevel, { dot: string; text: string; ring: string }> = {
  critical: { dot: 'bg-red-500', text: 'text-red-600 dark:text-red-400', ring: 'ring-red-500/40' },
  high: { dot: 'bg-orange-500', text: 'text-orange-600 dark:text-orange-400', ring: 'ring-orange-500/40' },
  medium: { dot: 'bg-amber-500', text: 'text-amber-600 dark:text-amber-400', ring: 'ring-amber-500/40' },
  low: { dot: 'bg-sky-500', text: 'text-sky-600 dark:text-sky-400', ring: 'ring-sky-500/40' },
}

export const KIND_ICON: Record<AdversaryKind, LucideIcon> = { ip: Globe2, host: Server, user: User }

export type ViewId = 'all' | 'high' | 'ip' | 'host' | 'user'

export const VIEWS: { id: ViewId; predicate: (a: Adversary) => boolean }[] = [
  { id: 'all', predicate: () => true },
  { id: 'high', predicate: (a) => a.maxSeverity >= 3 },
  { id: 'ip', predicate: (a) => a.kind === 'ip' },
  { id: 'host', predicate: (a) => a.kind === 'host' },
  { id: 'user', predicate: (a) => a.kind === 'user' },
]

export const LIST_COLS = '32px 1fr 100px 80px 90px 110px'

export function threatFromSeverity(sev: number): ThreatLevel {
  if (sev >= 4) return 'critical'
  if (sev === 3) return 'high'
  if (sev === 2) return 'medium'
  return 'low'
}

export function relativeTime(iso: string | undefined, t: TFunction): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (Number.isNaN(diff)) return '—'
  const m = Math.round(diff / 60_000)
  if (m < 1) return t('adversaries.relative.justNow')
  if (m < 60) return t('adversaries.relative.minutesAgo', { count: m })
  const h = Math.round(m / 60)
  if (h < 24) return t('adversaries.relative.hoursAgo', { count: h })
  return t('adversaries.relative.daysAgo', { count: Math.round(h / 24) })
}

export function flagEmoji(cc?: string): string {
  if (!cc || cc.length !== 2) return ''
  const A = 0x1f1e6
  return String.fromCodePoint(A - 65 + cc.toUpperCase().charCodeAt(0), A - 65 + cc.toUpperCase().charCodeAt(1))
}
