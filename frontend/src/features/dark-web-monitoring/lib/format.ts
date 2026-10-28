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

function safeDate(value: string | null | undefined): Date {
  const date = new Date(value ?? '')
  return Number.isNaN(date.getTime()) ? new Date(0) : date
}

/** Short display date (e.g. `8/28/26`) in the given locale. Invalid values render as that locale's epoch date. */
export function formatDateShort(value: string | null | undefined, locale?: string): string {
  return new Intl.DateTimeFormat(locale, { dateStyle: 'short' }).format(safeDate(value))
}

/** Medium display date (e.g. `Aug 28, 2026`) in the given locale. Invalid values render as that locale's epoch date. */
export function formatDateMedium(value: string | null | undefined, locale?: string): string {
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium' }).format(safeDate(value))
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

/**
 * Returns the value only when it is a safe image source (`https:`, `http:`, or
 * a `data:image/` URL). Anything else (e.g. `javascript:`) is dropped so
 * record-supplied `imageUrl` values can never run as code in an `img` tag.
 */
export function safeImageUrl(value: string | null | undefined): string | null {
  if (!value) return null
  const trimmed = value.trim()
  if (trimmed.startsWith('data:image/')) return value
  try {
    const url = new URL(trimmed, window.location.href)
    return url.protocol === 'https:' || url.protocol === 'http:' ? value : null
  } catch {
    return null
  }
}
