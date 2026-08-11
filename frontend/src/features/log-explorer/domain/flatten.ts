/**
 * Turning a stored document into the flat field/value pairs the explorer works
 * in. The table renders them, the CSV export writes them, and both have to
 * agree — a download that disagrees with the screen is worse than no download.
 * The flattening itself is shared: a dashboard table widget shows the same
 * records.
 */

export { flattenDoc } from '@/shared/lib/flatten'

/** The first candidate the document actually carries. */
export function pick(flat: Record<string, unknown>, fields: string[]): string | undefined {
  for (const f of fields) {
    const v = flat[f]
    if (v != null && v !== '') return String(v)
  }
  return undefined
}

// Field-name candidates for the compact result columns. A log has no single
// agreed name for "what happened" or "where from", so each is a list and the
// row uses whichever it has.
export const MSG_FIELDS = ['log.message', 'logx.message', 'message', 'event.original', 'rule.name', 'logx.raw']
// "Source" = where the log came from (the host/origin, which varies row to row).
export const SRC_FIELDS = ['dataSource', 'host.name', 'agent.name', 'log.computer', 'source', 'origin']

// Metadata/noise fields excluded from the document preview (same on every row).
const NOISE_KEYS = new Set(['@timestamp', '@version', 'dataType', 'deviceTime', 'id', 'isAnomaly', 'timestamp'])
const NOISE_PREFIXES = [
  'tenant',
  'globalaccount',
  'log.activityid',
  'log.correlation',
  'log.version',
  'log.opcode',
  'log.task',
  'log.keywords',
  'log.processid',
  'log.threadid',
  'log.recordid',
  'log.providerguid',
  'log.level',
]

function isNoise(k: string): boolean {
  if (NOISE_KEYS.has(k)) return true
  const lower = k.toLowerCase()
  return NOISE_PREFIXES.some((p) => lower.startsWith(p))
}

// Rank fields so the event-specific content (log.data.*) leads the preview.
function fieldRank(k: string): number {
  if (k.startsWith('log.data.')) return 0
  if (k.startsWith('event.')) return 1
  if (k.startsWith('log.') || k.startsWith('logx.') || k.startsWith('alert.')) return 2
  return 3
}

/**
 * Ordered, signal-carrying key/value pairs standing in for a message. Most
 * normalized logs have no message field at all — a cloudtrail record is an
 * event name and an ARN, a firewall record is five tuple fields — so "what
 * happened" has to be assembled from what the record does carry.
 */
export function docPreview(flat: Record<string, unknown>): [string, string][] {
  return Object.entries(flat)
    .filter(([k, v]) => v != null && v !== '' && k !== 'dataSource' && !isNoise(k))
    .sort((a, b) => fieldRank(a[0]) - fieldRank(b[0]))
    .slice(0, 8)
    .map(([k, v]) => [k.split('.').pop() ?? k, String(v)])
}

/** The preview as one cell: what the row shows, for somewhere that has no row. */
export function previewText(flat: Record<string, unknown>): string {
  return docPreview(flat)
    .map(([k, v]) => `${k}=${v}`)
    .join(' ')
}
