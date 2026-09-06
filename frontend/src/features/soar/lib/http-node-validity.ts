// Transient per-node body-validity registry for http nodes. Lives outside the
// flow model so the "user is mid-typing invalid JSON" state never leaks into
// YAML or the wire payload. FlowEditor consults it at save time and blocks.
// ponytail: module-scoped Map + clear() on FlowEditor mount — one owner at a
// time, no context provider ceremony.

const errors = new Map<string, string>()

export function setHttpBodyError(nodeId: string, err: string | null): void {
  if (err) errors.set(nodeId, err)
  else errors.delete(nodeId)
}

export function clearHttpBodyErrors(): void {
  errors.clear()
}

export function firstHttpBodyError(): { nodeId: string; err: string } | undefined {
  const it = errors.entries().next()
  if (it.done) return undefined
  return { nodeId: it.value[0], err: it.value[1] }
}

// http URL is considered valid if it parses as http(s):// after substituting
// $(alert.x) / $[variables.x] tokens away — we tolerate templates in path/host.
export function isValidHttpUrl(url: string): boolean {
  const trimmed = url.trim()
  if (!trimmed) return false
  const substituted = trimmed
    .replace(/\$\([^)]*\)/g, 'x')
    .replace(/\$\[[^\]]*\]/g, 'x')
  try {
    const u = new URL(substituted)
    return (u.protocol === 'http:' || u.protocol === 'https:') && Boolean(u.host)
  } catch {
    return false
  }
}
