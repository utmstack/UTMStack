import { useCallback, useEffect, useState } from 'react'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import type { Alert } from '../types/alert.types'

export interface UseAlertEchoesResult {
  echoes: Alert[]
  total: number
  page: number
  pageSize: number
  setPage: (p: number) => void
  setPageSize: (s: number) => void
  loading: boolean
  error: boolean
}

/** Paginated child-echo list for a parent alert. Page is 0-based; the service
 *  expects 1-based, so we translate at the boundary. Skips the fetch when
 *  parentId is null (collapsed row). */
export function useAlertEchoes(parentId: string | null): UseAlertEchoesResult {
  const [echoes, setEchoes] = useState<Alert[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  const setPageSizeReset = useCallback((s: number) => {
    setPageSize(s)
    setPage(0)
  }, [])

  useEffect(() => {
    if (!parentId) return
    let cancelled = false
    setLoading(true)
    setError(false)
    svc
      .echoes(parentId, page + 1, pageSize)
      .then(({ data, total }) => {
        if (cancelled) return
        setEchoes(data ?? [])
        setTotal(total)
      })
      .catch(() => {
        if (cancelled) return
        setError(true)
        setEchoes([])
        setTotal(0)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [parentId, page, pageSize])

  return { echoes, total, page, pageSize, setPage, setPageSize: setPageSizeReset, loading, error }
}
