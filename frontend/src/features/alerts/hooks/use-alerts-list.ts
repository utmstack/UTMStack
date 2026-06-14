import { useCallback, useEffect, useState } from 'react'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import type { Alert, FilterType } from '../types/alert.types'

export interface UseAlertsListResult {
  alerts: Alert[]
  total: number
  loading: boolean
  error: boolean
  refresh: () => void
}

/** Paginated alert list driven by `filters` (page is 0-based). */
export function useAlertsList(page: number, pageSize: number, filters: FilterType[]): UseAlertsListResult {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const { data, total } = await svc.list({ page: page + 1, size: pageSize, filters })
      // page is 0-based here; the first page replaces, later pages append.
      setAlerts((prev) => (page === 0 ? (data ?? []) : [...prev, ...(data ?? [])]))
      setTotal(total)
    } catch {
      setError(true)
      if (page === 0) {
        setAlerts([])
        setTotal(0)
      }
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, filters])

  useEffect(() => {
    void load()
  }, [load])

  return { alerts, total, loading, error, refresh: () => void load() }
}
