import { useEffect, useState, useCallback } from 'react'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import { WINDOW_MS, type LiveOpsData, type WindowKey } from '../liveops-data'

/** Batch size shared by the list sidebar and the timeline. */
const PAGE_SIZE = 300

/**
 * The single fetch point for Live Operations: parent alerts inside the window,
 * batched, cumulative, polled. Everything list-shaped (sidebar, timeline) reads
 * this through the context instead of fetching on its own.
 */
export function useLiveOpsAlerts(): LiveOpsData {
  const [win, setWinState] = useState<WindowKey>('1h')
  const [alerts, setAlerts] = useState<LiveOpsData['alerts']>(null)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const filters = [
    { field: 'parentId', operator: 'IS', value: '' },
    { field: '@timestamp', operator: 'IS_BETWEEN', value: [new Date(Date.now() - WINDOW_MS[win]).toISOString(), new Date(Date.now()).toISOString()] },
  ]


  const fetchAll = useCallback((silent: boolean) =>
    alertsHttpService.list({ page, size: PAGE_SIZE, filters })
      .then((r) => {
        setAlerts(r.data)
        setTotal(r.total)
        setLoading(false)
        if (!silent) setError('')
      })
      .catch((e: unknown) => {
        setLoading(false)
        if (!silent) setError(e instanceof Error ? e.message : String(e))
      }), [win, page])

  // One request: the single page in view. A poll or page change re-fetches it
  // in place, so what is on screen is always exactly one batch.
  useEffect(() => {
    setLoading(true)
    fetchAll(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [win, page])

  return {
    win,
    setWin: (w:WindowKey) => {
      setWinState(w)
      setPage(0)
    },
    alerts,
    total,
    page,
    pages,
    limit: PAGE_SIZE,
    loading,
    error,
    goPage: (p,forceRefresh=false) => {
      if(forceRefresh && p == page){
        fetchAll(false)
      }
      setPage(p)
    },
    loadMore: () => {
      if (loading || page >= pages - 1) return
      setPage((p) => p + 1)
    },
  }
}
