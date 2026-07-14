import type { TFunction } from 'i18next'

/** Liveness bucket derived from `lastPingAt` staleness — not a stored flag. */
export type DerivedStatus = 'live' | 'idle' | 'offline'

export const LIVE_MS = 15 * 60 * 1000
export const IDLE_MS = 24 * 60 * 60 * 1000

export function deriveStatus(lastPingAt?: string): DerivedStatus {
  if (!lastPingAt) return 'offline'
  const age = Date.now() - new Date(lastPingAt).getTime()
  if (age < LIVE_MS) return 'live'
  if (age < IDLE_MS) return 'idle'
  return 'offline'
}

/** Icon/colour only — labels are translated separately via `statusLabel`. */
export const STATUS_META: Record<DerivedStatus, { dot: string; tone: string }> = {
  live: { dot: 'bg-emerald-500', tone: 'text-emerald-500' },
  idle: { dot: 'bg-amber-500', tone: 'text-amber-500' },
  offline: { dot: 'bg-red-500', tone: 'text-red-500' },
}

export function statusLabel(t: TFunction, s: DerivedStatus): string {
  return t(`datasources.status.${s}`)
}
