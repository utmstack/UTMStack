import { useCallback, useRef, useState } from 'react'
import { notificationsHttpService } from '../services/notifications-http.service'
import type { NotificationGroup } from '../types/notification.types'

/**
 * Infinite-scroll feed of notification groups (source/type/message + counts),
 * paged `pageSize` at a time. Uses X-Total-Count from /notifications/grouped
 * to know when to stop. Groups are ordered by most-recent lastCreated first.
 */
export function useNotificationGroupFeed(pageSize: number) {
  const [groups, setGroups] = useState<NotificationGroup[]>([])
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [error, setError] = useState(false)

  const pageRef = useRef(0)
  const loadedRef = useRef(0)
  const totalRef = useRef<number | null>(null)
  const loadingRef = useRef(false)
  const doneRef = useRef(false)

  const loadMore = useCallback(async () => {
    if (loadingRef.current || doneRef.current) return
    loadingRef.current = true
    setLoading(true)
    setError(false)
    try {
      const { data, total } = await notificationsHttpService.listGrouped({
        page: pageRef.current,
        size: pageSize,
        status: 'ACTIVE',
      })
      totalRef.current = total
      loadedRef.current += data.length
      setGroups((prev) => [...prev, ...data])
      pageRef.current += 1
      if (data.length < pageSize || loadedRef.current >= total) {
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
    loadedRef.current = 0
    totalRef.current = null
    doneRef.current = false
    loadingRef.current = false
    setGroups([])
    setHasMore(true)
    setError(false)
  }, [])

  return { groups, setGroups, loading, hasMore, error, loadMore, reset }
}
