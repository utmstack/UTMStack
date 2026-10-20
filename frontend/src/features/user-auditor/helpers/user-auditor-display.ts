import type { TFunction } from 'i18next'
import type { ADUser } from '../types/ad-user.types'

export const STALE_MS = 30 * 86_400_000

export type ViewId = 'all' | ADUser['source']
export const VIEW_IDS: ViewId[] = ['all', 'windows', 'linux']

export type Status = 'active' | 'disabled' | 'deleted'

export const STATUS_DOT: Record<Status, string> = {
  active: 'bg-emerald-500',
  disabled: 'bg-amber-500',
  deleted: 'bg-red-500',
}
export const STATUS_TEXT: Record<Status, string> = {
  active: 'text-emerald-600 dark:text-emerald-400',
  disabled: 'text-amber-600 dark:text-amber-400',
  deleted: 'text-red-600 dark:text-red-400',
}

export function statusOf(u: ADUser): Status {
  if (u.accountDeletedAt) return 'deleted'
  if (!u.active) return 'disabled'
  return 'active'
}
export function isStale(u: ADUser): boolean {
  return statusOf(u) === 'active' && !!u.lastLogon && Date.now() - new Date(u.lastLogon).getTime() > STALE_MS
}
export function isService(u: ADUser): boolean {
  return u.source === 'windows' && /^svc/i.test(u.samAccountName ?? '')
}

/* Windows accounts key off SID + samAccountName; Linux accounts off username + uid@host. */
export function accountName(u: ADUser): string {
  return (u.source === 'linux' ? u.username : u.samAccountName) ?? '—'
}
export function accountIdentity(u: ADUser): string {
  if (u.source === 'linux') {
    const uid = u.uidNumber ? `uid=${u.uidNumber}` : ''
    const host = u.hostname ?? ''
    return [uid, host].filter(Boolean).join(' · ') || '—'
  }
  return u.sid ?? '—'
}
export function accountScope(u: ADUser): string {
  return (u.source === 'linux' ? u.hostname : u.domain) ?? '—'
}

export function relativeTime(iso: string | undefined, t: TFunction): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return t('userAuditor.relative.justNow')
  if (m < 60) return t('userAuditor.relative.minutesAgo', { count: m })
  const h = Math.round(m / 60)
  if (h < 24) return t('userAuditor.relative.hoursAgo', { count: h })
  return t('userAuditor.relative.daysAgo', { count: Math.round(h / 24) })
}

export function absTimestamp(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function logExplorerHref(user: ADUser): string {
  const p: Record<string, string> = { tenantId: user.tenantId, '@timestamp': '30d' }
  if (user.source === 'linux') {
    p.dataType = 'linux'
    if (user.username) p['origin.user'] = user.username
    if (user.hostname) p['origin.host'] = user.hostname
  } else {
    p.dataType = 'wineventlog'
    p.eventCode = '4720,4726,4624'
    if (user.sid) p.eventDataTargetSid = user.sid
  }
  return `/log-explorer?${new URLSearchParams(p).toString()}`
}
