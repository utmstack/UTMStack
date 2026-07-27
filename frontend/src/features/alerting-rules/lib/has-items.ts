export function hasItems(v: unknown): boolean {
  return Array.isArray(v) ? v.length > 0 : v != null && typeof v === 'object' && Object.keys(v).length > 0
}
