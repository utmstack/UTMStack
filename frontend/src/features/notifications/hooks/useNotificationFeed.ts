import { useCallback, useRef, useState } from 'react'
import { notificationsHttpService } from '../services/notifications-http.service'
import type { Notification } from '../types/notification.types'

/**
 * Infinite-scroll feed of ACTIVE notifications, paged `pageSize` at a time.
 * "Has more" is inferred from the last batch size (the list endpoint returns a
 * bare array, no total in the body). Newest first (backend default order).
 */
export function useNotificationFeed(pageSize: number) {
  const [items, setItems] = useState<Notification[]>([])
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [error, setError] = useState(false)

  const pageRef = useRef(0)
  const loadingRef = useRef(false)
  const doneRef = useRef(false)

  const loadMore = useCallback(async () => {
    if (loadingRef.current || doneRef.current) return
    loadingRef.current = true
    setLoading(true)
    setError(false)
    try {
      const batch = await notificationsHttpService.list({
        page: pageRef.current,
        size: pageSize,
        status: 'ACTIVE',
      })
      setItems((prev) => [...prev, ...batch])
      pageRef.current += 1
      if (batch.length < pageSize) {
        doneRef.current = true
        setHasMore(false)
      }
    } catch {
      setError(true)
    } finally {
      loadingRef.current = false
      setLoading(false)
    }
  }, [pageSize])

  const reset = useCallback(() => {
    pageRef.current = 0
    doneRef.current = false
    loadingRef.current = false
    setItems([])
    setHasMore(true)
    setError(false)
  }, [])

  return { items, setItems, loading, hasMore, error, loadMore, reset }
}
