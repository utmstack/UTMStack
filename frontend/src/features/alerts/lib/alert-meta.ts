import {
  type Alert,
  SEVERITY_BY_INT,
  STATUS_BY_INT,
  type SeverityKey,
  type StatusKey,
} from '../types/alert.types'

export const SEV_META: Record<SeverityKey, { label: string; bar: string; pill: string }> = {
  high: { label: 'High', bar: 'bg-red-500', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300' },
  medium: { label: 'Medium', bar: 'bg-amber-500', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300' },
  low: { label: 'Low', bar: 'bg-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300' },
}

export const ST_META: Record<StatusKey, { label: string; pill: string }> = {
  open: { label: 'Open', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300' },
  in_review: { label: 'In review', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300' },
  completed: { label: 'Completed', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300' },
  auto: { label: 'Automatic review', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300' },
}

export function sevKey(a: Alert): SeverityKey {
  return SEVERITY_BY_INT[a.severity ?? 1] ?? 'low'
}
export function statusKey(a: Alert): StatusKey {
  return STATUS_BY_INT[a.status ?? 2] ?? 'open'
}

export function riskOf(a: Alert): string {
  const r = a.impactScore ?? a.impact?.score
  return r != null ? String(Math.round(r)) : '—'
}

// Field paths contain dots; flatten them to a flat i18n key (dots are the
// i18next nesting separator, so they can't appear inside a single key).
export const fieldKey = (field: string) => field.replace(/\./g, '_')

// Filterable alert fields (ported from the legacy ALERT_FILTERS_FIELDS).
export const FILTER_FIELDS: { label: string; field: string }[] = [
  { label: 'Datasource group', field: 'assetGroupName' },
  { label: 'Category', field: 'category' },
  { label: 'Sensor', field: 'dataSource' },
  { label: 'Tags', field: 'tags' },
  { label: 'Incident name', field: 'incidentDetail.incidentName' },
  { label: 'Availability', field: 'impact.availability' },
  { label: 'Confidentiality', field: 'impact.confidentiality' },
  { label: 'Integrity', field: 'impact.integrity' },
  { label: 'Protocol', field: 'protocol' },
  { label: 'Adversary IP', field: 'adversary.ip' },
  { label: 'Adversary domain', field: 'adversary.domain' },
  { label: 'Adversary URL', field: 'adversary.url' },
  { label: 'Adversary ASN', field: 'adversary.geolocation.asn' },
  { label: 'Adversary ASO', field: 'adversary.geolocation.aso' },
  { label: 'Target IP', field: 'target.ip' },
  { label: 'Target domain', field: 'target.domain' },
  { label: 'Target URL', field: 'target.url' },
  { label: 'Target ASN', field: 'target.geolocation.asn' },
  { label: 'Target ASO', field: 'target.geolocation.aso' },
]

export const FILTER_OPS: { id: string; label: string; needsValue: boolean }[] = [
  { id: 'CONTAIN', label: 'contains', needsValue: true },
  { id: 'IS', label: 'is', needsValue: true },
  { id: 'IS_NOT', label: 'is not', needsValue: true },
  { id: 'EXIST', label: 'exists', needsValue: false },
  { id: 'DOES_NOT_EXIST', label: 'does not exist', needsValue: false },
]

// Palette offered when creating a new tag.
export const TAG_COLORS = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4', '#3b82f6', '#8b5cf6', '#ec4899', '#64748b']

export const TS = '@timestamp'
export const PAGE_SIZE_DEFAULT = 20

export const SELECT_CLS =
  'h-9 cursor-pointer rounded-md border border-input bg-background/40 px-2 text-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

export const TABLE_COLS = '6px 24px  1fr 130px 150px 200px 70px 90px 36px'

export function relativeTime(iso?: string) {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (Number.isNaN(diff)) return '—'
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'now'
  if (m < 60) return `${m}m`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h`
  return `${Math.round(h / 24)}d`
}

export function absTime(iso?: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  return Number.isNaN(d.getTime())
    ? iso
    : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

/** ISO 3166-1 alpha-2 → regional-indicator flag emoji. */
export function flagEmoji(cc?: string): string {
  if (!cc || !/^[a-zA-Z]{2}$/.test(cc)) return ''
  const base = 0x1f1e6
  const up = cc.toUpperCase()
  return String.fromCodePoint(base + up.charCodeAt(0) - 65, base + up.charCodeAt(1) - 65)
}
