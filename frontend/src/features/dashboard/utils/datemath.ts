const TOKEN_RE = /^now(?:([+-])(\d+)([mhdwM]))?$/

const UNIT_MS: Record<string, number> = {
  m: 60_000,
  h: 3_600_000,
  d: 86_400_000,
  w: 604_800_000,
  M: 30 * 86_400_000,
}

const EPOCH_ISO = '1970-01-01T00:00:00.000Z'

export function resolveDateMath(token: string | null, now: Date = new Date()): Date | null {
  if (token == null) return null
  const trimmed = token.trim()
  if (!trimmed) return null
  const m = TOKEN_RE.exec(trimmed)
  if (m) {
    if (!m[1]) return now
    const sign = m[1] === '+' ? 1 : -1
    const amount = parseInt(m[2], 10)
    const unit = m[3]
    const delta = sign * amount * (UNIT_MS[unit] ?? 0)
    return new Date(now.getTime() + delta)
  }
  const d = new Date(trimmed)
  return Number.isNaN(d.getTime()) ? null : d
}

export function toIndexTimestamp(date: Date | null): string {
  return date ? date.toISOString() : EPOCH_ISO
}

export function resolveRangeToISO(
  from: string | null,
  to: string,
  now: Date = new Date()
): { fromISO: string; toISO: string } {
  return {
    fromISO: toIndexTimestamp(resolveDateMath(from, now)),
    toISO: toIndexTimestamp(resolveDateMath(to, now) ?? now),
  }
}
