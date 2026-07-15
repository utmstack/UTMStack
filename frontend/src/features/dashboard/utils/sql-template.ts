import { filterRowToWhere } from '@/features/dashboard/utils/sql-builder'
import type { FilterRow, FilterType } from '@/features/dashboard/types'

const FROM_PLACEHOLDER = /\{\{\s*from\s*\}\}/g
const TO_PLACEHOLDER = /\{\{\s*to\s*\}\}/g
const TIME_FILTER_PLACEHOLDER = /\{\{\s*timeFilter\s*\}\}/g
const DASHBOARD_FILTERS_PLACEHOLDER = /\{\{\s*dashboardFilters\s*\}\}/g

const DEFAULT_TIMESTAMP_FIELD = '@timestamp'
const EMPTY_FILTERS: FilterType[] = []

export interface SqlTemplateInput {
  sql: string
  fromISO: string
  toISO: string
  timestampField?: string
  filters?: FilterType[]
}

// `{{dashboardFilters}}` sits directly against whatever follows it in the
// stored template (e.g. `{{dashboardFilters}}{{timeFilter}}`, no separator),
// so the fragment must end in " AND " when non-empty and be "" when there
// are no active dashboard filters.
function dashboardFiltersToWhere(filters: FilterType[]): string {
  const parts: string[] = []
  for (const f of filters) {
    const fragment = filterRowToWhere({ id: '', ...f } as FilterRow)
    if (fragment) parts.push(fragment)
  }
  return parts.length > 0 ? `${parts.join(' AND ')} AND ` : ''
}

// A quoted string literal compared against a date/timestamp field is parsed by
// OpenSearch SQL as an ODBC datetime literal, which expects `yyyy-MM-dd
// HH:mm:ss.SSS` — not full ISO-8601 (no `T` separator, no `Z`/offset suffix).
// `resolveRangeToISO` stays real ISO-8601 (it also feeds query-cache keys);
// convert only here, right before the value lands in SQL text.
function toOdbcTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  const pad = (n: number, len = 2) => String(n).padStart(len, '0')
  return (
    `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())} ` +
    `${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}.${pad(d.getUTCMilliseconds(), 3)}`
  )
}

export function applySqlTemplate({
  sql,
  fromISO,
  toISO,
  timestampField = DEFAULT_TIMESTAMP_FIELD,
  filters = EMPTY_FILTERS,
}: SqlTemplateInput): string {
  const from = toOdbcTimestamp(fromISO)
  const to = toOdbcTimestamp(toISO)
  const tf = `${timestampField} >= '${from}' AND ${timestampField} <= '${to}'`
  const dashboardWhere = dashboardFiltersToWhere(filters)
  return sql
    .replace(DASHBOARD_FILTERS_PLACEHOLDER, () => dashboardWhere)
    .replace(TIME_FILTER_PLACEHOLDER, tf)
    .replace(FROM_PLACEHOLDER, from)
    .replace(TO_PLACEHOLDER, to)
}

export function hasTimePlaceholder(sql: string): boolean {
  return (
    FROM_PLACEHOLDER.test(sql) ||
    TO_PLACEHOLDER.test(sql) ||
    TIME_FILTER_PLACEHOLDER.test(sql)
  )
}
