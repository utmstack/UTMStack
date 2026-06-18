import type { IncidentStatus } from '../types/incident.types'

export const STATUSES: IncidentStatus[] = ['OPEN', 'IN_REVIEW', 'COMPLETED', 'MERGED']

export const ST_META: Record<IncidentStatus, { dot: string; pill: string; col: string }> = {
  OPEN: { dot: 'bg-red-500', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', col: 'border-t-red-500' },
  IN_REVIEW: { dot: 'bg-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300', col: 'border-t-sky-500' },
  COMPLETED: { dot: 'bg-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', col: 'border-t-emerald-500' },
  MERGED: { dot: 'bg-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', col: 'border-t-violet-500' },
}

export type SevKey = 'high' | 'medium' | 'low' | 'unknown'

export function sevKey(n?: number): SevKey {
  if (n === 3) return 'high'
  if (n === 2) return 'medium'
  if (n === 1) return 'low'
  return 'unknown'
}

export const SEV_TONE: Record<SevKey, string> = {
  high: 'text-red-500',
  medium: 'text-amber-500',
  low: 'text-sky-500',
  unknown: 'text-muted-foreground',
}

export const SELECT_CLS = 'h-9 rounded-md border border-border bg-background px-2 text-sm'
export const TABLE_COLS = '1fr 120px 90px 150px 70px 110px'
