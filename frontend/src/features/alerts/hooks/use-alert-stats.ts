import { useCallback, useEffect, useState } from 'react'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import type { FilterType } from '../types/alert.types'

export interface UseAlertStatsResult {
  sevCounts: Record<string, number>
  statusCounts: Record<string, number>
  timeline: number[]
  openCount: number | null
  refresh: () => void
}

/**
 * Severity counts, status counts, the alerts-over-time histogram and the
 * total-open badge in one hook. Each query is independent — partial failures
 * just keep the prior value.
 */
export function useAlertStats(scopeFilters: FilterType[], interval: string): UseAlertStatsResult {
  const [sevCounts, setSevCounts] = useState<Record<string, number>>({})
  const [statusCounts, setStatusCounts] = useState<Record<string, number>>({})
  const [timeline, setTimeline] = useState<number[]>([])
  const [openCount, setOpenCount] = useState<number | null>(null)

  const load = useCallback(async () => {
    const [sev, st, tl, open] = await Promise.allSettled([
      svc.counts('severity', scopeFilters),
      svc.counts('status', scopeFilters),
      svc.timeline(scopeFilters, interval),
      svc.countOpen(),
    ])
    if (sev.status === 'fulfilled') setSevCounts(Object.fromEntries(sev.value.top.map((b) => [b.value, b.count])))
    if (st.status === 'fulfilled') setStatusCounts(Object.fromEntries(st.value.top.map((b) => [b.value, b.count])))
    if (tl.status === 'fulfilled') setTimeline(tl.value.values ?? [])
    if (open.status === 'fulfilled') setOpenCount(open.value)
  }, [scopeFilters, interval])

  useEffect(() => {
    void load()
  }, [load])

  return { sevCounts, statusCounts, timeline, openCount, refresh: () => void load() }
}
