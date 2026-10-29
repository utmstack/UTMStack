import { createContext, useContext } from 'react'
import type { Alert } from '@/features/alerts/types/alert.types'

/** Time window options; labels are plain (1h / 6h / 24h / 7d). */
export type WindowKey = '1h' | '6h' | '24h' | '7d'
export const WINDOW_OPTIONS: { key: WindowKey; label: string }[] = [
  { key: '1h', label: '1h' },
  { key: '6h', label: '6h' },
  { key: '24h', label: '24h' },
  { key: '7d', label: '7d' },
]
export const WINDOW_MS: Record<WindowKey, number> = {
  '1h': 3_600_000,
  '6h': 6 * 3_600_000,
  '24h': 24 * 3_600_000,
  '7d': 7 * 24 * 3_600_000,
}

export interface LiveOpsData {
  win: WindowKey
  /** Switch window; resets to the first batch. */
  setWin: (w: WindowKey) => void
  /** Cumulative list: every batch from 0 to `page`. */
  alerts: Alert[] | null
  total: number
  /** Farthest batch loaded (0-based). */
  page: number
  pages: number
  /** Batch size, exposed so consumers compute hasMore. */
  limit: number
  /** True while a batch (re)fetch is in flight. */
  loading: boolean
  error: string
  /** Jump to batch p; re-fetches 0..p. */
  goPage: (p: number) => void
  /** Load the next batch (infinite scroll). */
  loadMore: () => void
}

export const LiveOpsDataContext = createContext<LiveOpsData | null>(null)

export function useLiveOpsData(): LiveOpsData {
  const ctx = useContext(LiveOpsDataContext)
  if (!ctx) throw new Error('useLiveOpsData must be used inside LiveOpsDataContext')
  return ctx
}
