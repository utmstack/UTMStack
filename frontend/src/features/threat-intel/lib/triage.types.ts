import type { EntitySearchItem } from '@/features/threat-intel/domain/threat-intel.types'

export type Verdict = 'malicious' | 'suspicious' | 'benign' | 'unknown' | 'notFound' | 'invalid'

export interface TriageRow {
  /** Stable key for React; the raw input may repeat across a paste. */
  key: string
  input: string
  type: string
  verdict: Verdict
  entity: EntitySearchItem | null
  error?: string
}

/** Upper bound on one triage run. */
export const MAX_TRIAGE_ITEMS = 500

/**
 * Each row is a separate API call, so unbounded fan-out over a 500-line paste
 * bursts the rate limit. Six keeps a large batch moving.
 */
export const TRIAGE_CONCURRENCY = 6

/**
 * Splits pasted text into candidate indicators. Real pastes are messy — CSV
 * columns, quoted values, bracketed defanging, bullet points — so this is
 * deliberately liberal about separators.
 */
export function parseIndicators(raw: string): string[] {
  const seen = new Set<string>()
  const out: string[] = []

  for (const token of raw.split(/[\s,;|]+/)) {
    const cleaned = refang(token)
    if (!cleaned) continue
    // Indicators are not case-sensitive; a report often repeats the same host
    // in different casing.
    const dedupeKey = cleaned.toLowerCase()
    if (seen.has(dedupeKey)) continue
    seen.add(dedupeKey)
    out.push(cleaned)
  }

  return out
}

/**
 * Undoes the common "defanging" conventions used to make indicators
 * unclickable in reports: `1.2.3[.]4`, `hxxp://`, wrapping brackets and quotes.
 */
export function refang(token: string): string {
  return token
    .trim()
    .replace(/^[<("'[]+|[>)"'\].,;]+$/g, '')
    .replace(/\[\.\]|\(\.\)|\{\.\}/g, '.')
    .replace(/\[:\]/g, ':')
    .replace(/\[at\]/gi, '@')
    .replace(/^hxxp(s?):\/\//i, 'http$1://')
    .trim()
}

/**
 * Maps a looked-up entity to a verdict. Reputation runs -3..+3. Anything
 * negative is actionable; -3/-2 are malicious, -1 suspicious.
 */
export function verdictFor(entity: EntitySearchItem | null): Verdict {
  if (!entity) return 'notFound'
  if (entity.reputation <= -2) return 'malicious'
  if (entity.reputation < 0) return 'suspicious'
  if (entity.reputation > 0) return 'benign'
  return 'unknown'
}

/** Sort order: worst first, so the analyst reads top-down. */
const VERDICT_RANK: Record<Verdict, number> = {
  malicious: 0,
  suspicious: 1,
  unknown: 2,
  benign: 3,
  notFound: 4,
  invalid: 5,
}

export function sortRows(rows: TriageRow[]): TriageRow[] {
  return [...rows].sort((a, b) => {
    const byVerdict = VERDICT_RANK[a.verdict] - VERDICT_RANK[b.verdict]
    if (byVerdict !== 0) return byVerdict
    // Within a verdict, the more negative reputation is more urgent.
    const repA = a.entity?.reputation ?? 0
    const repB = b.entity?.reputation ?? 0
    if (repA !== repB) return repA - repB
    return a.input.localeCompare(b.input)
  })
}

export interface TriageOptions {
  signal?: AbortSignal
  onRow?: (row: TriageRow) => void
  /** Resolves a single indicator; injected by the hook (API-bound). */
  lookup: (input: string, index: number, signal?: AbortSignal) => Promise<TriageRow>
}

/**
 * Resolves every indicator (bounded concurrency), reporting progress as
 * results land. Rows are emitted through `onRow` so a 500-item batch fills
 * the table progressively instead of showing a spinner for a minute.
 */
export async function triageIndicators(inputs: string[], options: TriageOptions): Promise<TriageRow[]> {
  const { signal, onRow, lookup } = options
  const capped = inputs.slice(0, MAX_TRIAGE_ITEMS)
  const rows: (TriageRow | null)[] = []
  rows.length = capped.length
  let cursor = 0

  async function worker() {
    while (cursor < capped.length) {
      if (signal?.aborted) return
      const index = cursor++
      rows[index] = await lookup(capped[index], index, signal)
      onRow?.(rows[index]!)
    }
  }

  await Promise.all(
    Array.from({ length: Math.min(TRIAGE_CONCURRENCY, capped.length) }, () => worker()),
  )

  return rows.flatMap(row => (row ? [row] : []))
}

/** STIX 2.1 pattern for an indicator, or null for types with no mapping. */
function stixPattern(type: string, value: string): string | null {
  const escaped = value.replace(/\\/g, '\\\\').replace(/'/g, "\\'")
  switch (type) {
    case 'ip':
      return `[ipv4-addr:value = '${escaped}']`
    case 'domain':
    case 'hostname':
      return `[domain-name:value = '${escaped}']`
    case 'url':
      return `[url:value = '${escaped}']`
    case 'email':
    case 'email-address':
      return `[email-addr:value = '${escaped}']`
    case 'md5':
    case 'sha1':
    case 'sha256':
      return `[file:hashes.'${type.toUpperCase()}' = '${escaped}']`
    default:
      return null
  }
}

/**
 * STIX 2.1 bundle of the actionable rows. Only malicious and suspicious rows
 * are exported: a STIX indicator asserts that something is of interest, so
 * emitting benign or not-found values would poison whatever consumes it.
 */
export function toStix(rows: TriageRow[]): string {
  const objects = rows
    .filter(row => row.verdict === 'malicious' || row.verdict === 'suspicious')
    .map(row => {
      const pattern = stixPattern(row.type, row.input)
      if (!pattern) return null
      const timestamp = row.entity?.lastSeen || new Date().toISOString()
      return {
        type: 'indicator',
        spec_version: '2.1',
        id: `indicator--${stableUuid(row.input)}`,
        created: timestamp,
        modified: timestamp,
        name: row.input,
        pattern,
        pattern_type: 'stix',
        valid_from: timestamp,
        labels: row.entity?.tags?.length ? row.entity.tags : ['malicious-activity'],
        confidence: confidenceFrom(row.entity?.accuracy ?? 0),
      }
    })
    .filter(Boolean)

  return JSON.stringify(
    { type: 'bundle', id: `bundle--${stableUuid(String(rows.length) + Date.now())}`, objects },
    null,
    2,
  )
}

/**
 * Maps the 0..3 accuracy score onto STIX's 0-100 confidence scale, using the
 * midpoint of each band so no value claims certainty the corpus does not have.
 */
function confidenceFrom(accuracy: number): number {
  switch (accuracy) {
    case 3:
      return 85
    case 2:
      return 65
    case 1:
      return 40
    default:
      return 15
  }
}

/**
 * Deterministic UUIDv4-shaped identifier derived from the indicator, so
 * re-exporting the same batch produces stable IDs. Not cryptographic — STIX
 * only requires uniqueness, not secrecy.
 */
function stableUuid(seed: string): string {
  let h1 = 0x811c9dc5
  let h2 = 0x01000193
  for (let i = 0; i < seed.length; i++) {
    h1 = Math.imul(h1 ^ seed.charCodeAt(i), 0x01000193) >>> 0
    h2 = Math.imul(h2 ^ seed.charCodeAt(seed.length - 1 - i), 0x85ebca6b) >>> 0
  }
  const hex = (n: number) => n.toString(16).padStart(8, '0')
  const raw = `${hex(h1)}${hex(h2)}${hex(h1 ^ h2)}${hex((h1 + h2) >>> 0)}`
  return [
    raw.slice(0, 8),
    raw.slice(8, 12),
    `4${raw.slice(13, 16)}`,
    `8${raw.slice(17, 20)}`,
    raw.slice(20, 32),
  ].join('-')
}
