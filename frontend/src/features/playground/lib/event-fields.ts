/**
 * Field-flattening and severity-styling helpers shared by the playground's
 * result views (`ParsedEventView` for the parsed event, `AlertsListView` for
 * triggered alerts). Extracted from `ParsedEventView` so both views render
 * fields identically — intentionally not shared with log-explorer's own
 * flattening so the playground can evolve independently for testing purposes.
 */

/** Flattens a nested object into dotted-path key/value pairs, e.g.
 * `{ origin: { ip: '1.2.3.4' } }` → `{ 'origin.ip': '1.2.3.4' }`. */
export function flatten(obj: unknown, prefix = '', out: Record<string, unknown> = {}): Record<string, unknown> {
  if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
    for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
      const key = prefix ? `${prefix}.${k}` : k
      if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, out)
      else out[key] = Array.isArray(v) ? v.join(', ') : v
    }
  }
  return out
}

// Fields that duplicate what's already visible in the "Raw log" input on the
// left — showing the full raw string again inside the field grid wrecks its
// rhythm (one giant multi-line cell among dozens of short ones) for zero gain.
export const HIDDEN_FIELDS = new Set(['raw'])

// Same severity/level palette already used across the product (see
// alerts/lib/alert-meta.ts SEV_META and log-explorer's LEVEL_TONE) — kept
// local here since the playground is intentionally self-contained.
export const SEVERITY_TONE: Record<string, string> = {
  critical: 'bg-red-500/15 text-red-600 ring-1 ring-inset ring-red-500/30 dark:text-red-300',
  high: 'bg-red-500/15 text-red-600 ring-1 ring-inset ring-red-500/30 dark:text-red-300',
  error: 'bg-orange-500/15 text-orange-600 ring-1 ring-inset ring-orange-500/30 dark:text-orange-300',
  medium: 'bg-amber-500/15 text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300',
  warning: 'bg-amber-500/15 text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300',
  warn: 'bg-amber-500/15 text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300',
  low: 'bg-sky-500/15 text-sky-600 ring-1 ring-inset ring-sky-500/30 dark:text-sky-300',
  info: 'bg-sky-500/15 text-sky-600 ring-1 ring-inset ring-sky-500/30 dark:text-sky-300',
  debug: 'bg-muted text-muted-foreground ring-1 ring-inset ring-border',
}

export function isSeverityField(key: string): boolean {
  const last = key.split('.').pop() ?? key
  return /(severity|level)$/i.test(last)
}
