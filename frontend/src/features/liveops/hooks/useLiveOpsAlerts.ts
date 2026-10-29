import { useEffect, useState } from 'react'
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

  // Batches 0..page in parallel; replaced on every load so a scroll-append and a
  // poll never disagree about what is in the list.
  useEffect(() => {
    const fetchAll = (silent: boolean) =>
      Promise.all(Array.from({ length: page + 1 }, (_, p) => alertsHttpService.list({ page: p, size: PAGE_SIZE, filters })))
        .then((rs) => {
          setAlerts(rs.flatMap((r) => r.data))
          setTotal(rs[0]?.total ?? 0)
          setLoading(false)
          if (!silent) setError('')
        })
        .catch((e: unknown) => {
          setLoading(false)
          if (!silent) setError(e instanceof Error ? e.message : String(e))
        })
    setLoading(true)
    fetchAll(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [win, page])

  return {
    win,
    setWin: (w) => {
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
    goPage: (p) => setPage(Math.min(Math.max(p, 0), pages - 1)),
    loadMore: () => {
      if (loading || page >= pages - 1) return
      setPage((p) => p + 1)
    },
  }
}
