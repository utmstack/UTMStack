import type { TFunction } from 'i18next'
import { CLOUD_SOURCES } from '../types/ad-user.types'
import type { ADUser, ADUserSource } from '../types/ad-user.types'

export const STALE_MS = 30 * 86_400_000

export type ViewId = 'all' | ADUser['source']
export const VIEW_IDS: ViewId[] = ['all', 'windows', 'linux', ...CLOUD_SOURCES]

export function isCloudSource(source: ADUser['source']): boolean {
  return (CLOUD_SOURCES as string[]).includes(source)
}

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

/* Windows accounts key off SID + samAccountName; Linux accounts off username + uid@host; cloud accounts off username only. */
export function accountName(u: ADUser): string {
  if (u.source === 'linux' || isCloudSource(u.source)) return u.username ?? '—'
  return u.samAccountName ?? '—'
}
export function accountIdentity(u: ADUser): string {
  if (u.source === 'linux') {
    const uid = u.uidNumber ? `uid=${u.uidNumber}` : ''
    const host = u.hostname ?? ''
    return [uid, host].filter(Boolean).join(' · ') || '—'
  }
  if (isCloudSource(u.source)) return u.username ?? '—'
  return u.sid ?? '—'
}
export function accountScope(u: ADUser): string {
  if (isCloudSource(u.source)) return '—'
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

const CLOUD_DATA_TYPE: Record<ADUserSource, string> = {
  windows: '',
  linux: '',
  o365: 'o365,o365-graph',
  azure: 'azure',
  gcp: 'google',
  aws: 'aws',
}

export function logExplorerHref(user: ADUser): string | null {
  const p: Record<string, string> = { tenantId: user.tenantId, '@timestamp': '30d' }
  if (isCloudSource(user.source)) {
    p.dataType = CLOUD_DATA_TYPE[user.source]
    if (user.username) p['origin.user'] = user.username
  } else if (user.source === 'linux') {
    p.dataType = 'linux'
    if (user.username) p['origin.user'] = user.username
    if (user.hostname) p['origin.host'] = user.hostname
  } else {
    p.dataType = 'wineventlog'
    p['log.eventCode'] = '4720,4726,4624'
    if (user.sid) p['log.eventDataTargetSid'] = user.sid
  }
  return `/log-explorer?${new URLSearchParams(p).toString()}`
}
