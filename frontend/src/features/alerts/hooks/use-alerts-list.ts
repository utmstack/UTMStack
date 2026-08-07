import { useCallback, useEffect, useMemo, useState } from 'react'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import type { Alert, FilterType } from '../types/alert.types'

export interface UseAlertsListResult {
  alerts: Alert[]
  total: number
  /** Whether another page of the *current* query exists. */
  hasMore: boolean
  loading: boolean
  error: boolean
  refresh: () => void
}

/** What the accumulated list belongs to. Two queries that ask the same thing
 *  share it, so a re-render alone never discards rows. */
function queryKey(filters: FilterType[], pageSize: number): string {
  return `${pageSize}|${JSON.stringify(filters)}`
}

/**
 * Paginated alert list driven by `filters` (page is 0-based, as the search
 * endpoint counts them).
 *
 * The accumulated rows are tagged with the query that produced them. Deciding
 * "replace or append" from the page number alone was not enough: the sentinel
 * that grows the page can fire in the gap between a filter changing and its
 * first page landing, and the next response then appended one query's rows onto
 * another's — a filtered count above a list still showing everything.
 */
export function useAlertsList(page: number, pageSize: number, filters: FilterType[]): UseAlertsListResult {
  const key = useMemo(() => queryKey(filters, pageSize), [filters, pageSize])
  const [acc, setAcc] = useState<{ key: string | null; alerts: Alert[]; total: number }>({
    key: null,
    alerts: [],
    total: 0,
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      // A page bump that belongs to the previous query is not a page of this
      // one, so it starts over rather than fetching an offset into results
      // nobody has seen yet.
      const fresh = page === 0 || acc.key !== key
      const { data, total } = await svc.list({ page: fresh ? 0 : page, size: pageSize, filters })
      setAcc((prev) => ({
        key,
        alerts: fresh ? (data ?? []) : [...prev.alerts, ...(data ?? [])],
        total,
      }))
    } catch {
      setError(true)
      if (page === 0) setAcc({ key, alerts: [], total: 0 })
    } finally {
      setLoading(false)
    }
    // acc.key is read, not tracked: including it would re-run the load its own
    // result caused.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, filters, key])

  useEffect(() => {
    void load()
  }, [load])

  return {
    alerts: acc.alerts,
    total: acc.total,
    // Until this query's own rows have landed there is nothing to append to,
    // and offering more is what let the sentinel race ahead of the filter.
    hasMore: acc.key === key && acc.alerts.length < acc.total,
    loading,
    error,
    refresh: () => void load(),
  }
}
