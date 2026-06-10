import { useSyncExternalStore } from 'react'

/**
 * Org-wide timestamp display preference (from GET /date-format). Data is always
 * stored in UTC; this only controls how timestamps are *rendered*. Mirrors the
 * legacy TimezoneFormatService + UtmDatePipe: a single source applied app-wide.
 */
export type DateFormatStyle = 'short' | 'medium' | 'long' | 'full'

export interface DateFormatPrefs {
  /** IANA tz name, 'UTC', or 'system' (browser local). */
  timezone: string
  dateFormat: DateFormatStyle
}

const STYLES: DateFormatStyle[] = ['short', 'medium', 'long', 'full']

let prefs: DateFormatPrefs = { timezone: 'UTC', dateFormat: 'medium' }
const listeners = new Set<() => void>()

/** Update the global prefs (called by DateFormatProvider after fetch/save). */
export function setDateFormatPrefs(p: Partial<DateFormatPrefs>): void {
  const dateFormat = STYLES.includes(p.dateFormat as DateFormatStyle)
    ? (p.dateFormat as DateFormatStyle)
    : 'medium'
  prefs = { timezone: p.timezone || 'UTC', dateFormat }
  listeners.forEach((l) => l())
}

export function getDateFormatPrefs(): DateFormatPrefs {
  return prefs
}

function subscribe(l: () => void): () => void {
  listeners.add(l)
  return () => listeners.delete(l)
}

function timeZone(): string | undefined {
  return prefs.timezone && prefs.timezone !== 'system' ? prefs.timezone : undefined
}

/** Full date + time, in the configured timezone & style. */
export function formatDateTime(value: string | number | Date): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: prefs.dateFormat,
    timeStyle: 'medium',
    timeZone: timeZone(),
  }).format(d)
}

/** Date only, in the configured timezone & style. */
export function formatDate(value: string | number | Date): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: prefs.dateFormat,
    timeZone: timeZone(),
  }).format(d)
}

/** Time only, in the configured timezone. */
export function formatTime(value: string | number | Date): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return new Intl.DateTimeFormat(undefined, {
    timeStyle: 'medium',
    timeZone: timeZone(),
  }).format(d)
}

/**
 * Subscribe a component to the prefs so it re-renders when they load/change, and
 * get the formatters back. Use this in any component that renders timestamps.
 */
export function useDateFormat() {
  useSyncExternalStore(subscribe, getDateFormatPrefs, getDateFormatPrefs)
  return { formatDateTime, formatDate, formatTime, prefs }
}
