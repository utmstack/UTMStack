import { resolveTimeBounds, type TimeFilter } from '../components/TimeRangeFilter'
import type { BreachStatus, DarkWebRecord, LeakSurface, OrganizationElementType } from '../types'

/**
 * Client-side record query, kept out of the page component. The dark web data
 * is served by an in-memory mock for now, so filtering/pagination live here
 * instead of in a backend endpoint; when the API lands, move this logic into
 * the service (query parameters) and delete this module.
 */

export interface RecordQuery {
  surface?: LeakSurface
  status?: BreachStatus | 'all'
  employeesOnly?: boolean
  /** Case-insensitive substring across the searchable record fields. */
  search?: string
  /** Page-level time window (applies to both breached and ingested times). */
  time?: TimeFilter
  /** Extra window on the breached `timestamp` only (custom range export). */
  timestampBetween?: { from: number; to: number }
}

export const SURFACE_TYPES: Record<LeakSurface, OrganizationElementType> = {
  domains: 'DOMAIN',
  emails: 'EMAIL',
  ips: 'IP',
  usernames: 'USER',
  phones: 'PHONE',
}

export const SURFACE_OF: Record<OrganizationElementType, LeakSurface> = {
  DOMAIN: 'domains',
  EMAIL: 'emails',
  IP: 'ips',
  USER: 'usernames',
  PHONE: 'phones',
}

function inBounds(at: number, from: number, to: number): boolean {
  return at >= from && at <= to
}

export function matchesRecordQuery(record: DarkWebRecord, query: RecordQuery): boolean {
  if (query.surface && record.organizationElementType !== SURFACE_TYPES[query.surface]) return false
  if (query.status && query.status !== 'all' && record.breachStatus !== query.status) return false
  if (query.employeesOnly && !record.employee) return false
  if (query.time) {
    const bounds = resolveTimeBounds(query.time)
    // Both the breached and the ingested time must sit in the window, so the
    // page, the table and the exports agree on what "in range" means.
    if (!inBounds(new Date(record.timestamp).getTime(), bounds.from, bounds.to)) return false
    if (!inBounds(new Date(record.createdAt).getTime(), bounds.from, bounds.to)) return false
  }
  if (query.timestampBetween && !inBounds(new Date(record.timestamp).getTime(), query.timestampBetween.from, query.timestampBetween.to)) return false
  const needle = query.search?.trim().toLowerCase()
  if (needle) {
    const searchable = [
      record.domain,
      record.email,
      record.ipAddress,
      record.username,
      record.name,
      record.phone,
      record.organizationElement,
      record.databaseName,
      record.leakName,
      record.foundIn,
      record.source,
      record.compromisedData,
    ]
    if (!searchable.some((value) => value?.toLowerCase().includes(needle))) return false
  }
  return true
}

export function filterRecords(records: DarkWebRecord[], query: RecordQuery): DarkWebRecord[] {
  return records.filter((record) => matchesRecordQuery(record, query))
}

/** First `count` records, for client-side "load more" pagination. */
export function paginateRecords(records: DarkWebRecord[], count: number): DarkWebRecord[] {
  return records.slice(0, count)
}
