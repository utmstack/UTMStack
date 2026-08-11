/**
 * A stored document as flat field/value pairs. Records are nested — a log's
 * origin has a geolocation, an alert carries its events — and everything that
 * shows one as a row (the explorer's table, a dashboard table widget, a CSV)
 * has to agree on how a nested field is named, or the same record reads
 * differently in each place.
 */

// An array of tags reads as "a, b"; an array of objects — an alert's events or
// its history — has to be rendered as what it is. join() would call String() on
// each element and print [object Object].
function joinArray(v: unknown[]): string {
  return v
    .map((e) => (e !== null && typeof e === 'object' ? JSON.stringify(e) : String(e)))
    .join(', ')
}

export function flattenDoc(
  obj: unknown,
  prefix = '',
  out: Record<string, unknown> = {}
): Record<string, unknown> {
  if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
    for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
      const key = prefix ? `${prefix}.${k}` : k
      if (v && typeof v === 'object' && !Array.isArray(v)) flattenDoc(v, key, out)
      else out[key] = Array.isArray(v) ? joinArray(v) : v
    }
  }
  return out
}
