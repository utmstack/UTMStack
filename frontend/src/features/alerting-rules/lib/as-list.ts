export function asList(v: unknown): string {
  return Array.isArray(v) ? v.join(', ') || '—' : '—'
}
