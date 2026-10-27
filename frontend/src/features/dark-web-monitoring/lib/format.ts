import type { BreachStatus } from '../types'

/** i18n key suffix for a breach status (darkWebMonitoring.status.<key>). */
export const STATUS_KEYS: Record<BreachStatus, string> = {
  OPEN: 'open',
  IN_MITIGATION: 'mitigation',
  COMPLETED: 'completed',
}

export const EMPTY_VALUE = '—'

/** Renders `null`/empty values as a dash so tables stay readable. */
export function valueOrDash(value: string | null | undefined): string {
  return value || EMPTY_VALUE
}

const dateFormatters: Partial<Record<Exclude<Intl.DateTimeFormatOptions['dateStyle'], undefined>, Intl.DateTimeFormat>> = {}

function dateFormatter(dateStyle: NonNullable<Intl.DateTimeFormatOptions['dateStyle']>): Intl.DateTimeFormat {
  let formatter = dateFormatters[dateStyle]
  if (!formatter) {
    formatter = new Intl.DateTimeFormat(undefined, { dateStyle })
    dateFormatters[dateStyle] = formatter
  }
  return formatter
}

function safeDate(value: string | null | undefined): Date {
  const date = new Date(value ?? '')
  return Number.isNaN(date.getTime()) ? new Date(0) : date
}

/** Short display date (e.g. `8/28/26`), locale-aware. Invalid values render as 1970-01-01. */
export function formatDateShort(value: string | null | undefined): string {
  return dateFormatter('short').format(safeDate(value))
}

/** Medium display date (e.g. `Aug 28, 2026`), locale-aware. Invalid values render as 1970-01-01. */
export function formatDateMedium(value: string | null | undefined): string {
  return dateFormatter('medium').format(safeDate(value))
}

/**
 * Neutralizes CSV injection: cells that would be parsed as formulas by
 * spreadsheet applications (Excel, Google Sheets, Numbers) get a leading
 * single quote. The display value is unchanged; only the stored cell is.
 */
export function neutralizeCsvCell(value: string): string {
  return /^[=+@\-]/.test(value) ? `'${value}` : value
}

/** HTML-escapes a data-derived value before it is interpolated into trusted locale markup. */
export function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (character) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[character] ?? character)
}
