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
 * Neutralizes CSV injection: cells whose (trimmed) leading character would be
 * parsed as a formula by spreadsheet applications (Excel, Google Sheets,
 * Numbers) get a leading single quote. Whitespace and control characters
 * before the trigger are stripped for the check so ` =SUM(...)` is caught too.
 * The display value is unchanged; only the stored cell is.
 */
export function neutralizeCsvCell(value: string): string {
  const lead = value.replace(/^[\s\u0000-\u001f]+/, '')
  return /^[=+@-]/.test(lead) ? `'${value}` : value
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
 * Minimal allowlist sanitizer for i18n strings that embed markup (e.g.
 * `<strong>` in the breach details copy). Only a fixed set of harmless inline
 * tags survives, always without attributes; every other tag is escaped, so
 * translated content can never inject scripts even if it were ever
 * backend-supplied or tenant-editable.
 */
const ALLOWED_HTML_TAGS = new Set(['strong', 'em', 'b', 'i', 'code'])

export function safeHtml(markup: string): string {
  return markup
    .replace(/<!--[\s\S]*?-->/g, '')
    .replace(/<\/?([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>/g, (tag, name: string) => {
      const lower = name.toLowerCase()
      if (ALLOWED_HTML_TAGS.has(lower)) return tag.startsWith('</') ? `</${lower}>` : `<${lower}>`
      return tag.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    })
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
