import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useAuth } from '@/features/auth'
import { IS_FEDERATION } from '@/shared/config/mode'
import { useCurrentInstanceId } from '@/shared/lib/current-instance'
import { notificationsHttpService } from './notifications-http.service'

interface NotificationsContextValue {
  unreadCount: number
  refreshUnread: () => Promise<void>
  markRead: (id: number, read: boolean) => Promise<void>
  markAllRead: () => Promise<void>
  remove: (id: number) => Promise<void>
}

const UNREAD_POLL_MS = 60_000

const NotificationsContext = createContext<NotificationsContextValue | undefined>(undefined)

/**
 * Owns the unread badge count (polled while authenticated) and the write actions
 * shared by the bell dropdown and the Notifications page. List data itself is
 * fetched per-view (see useNotificationFeed) — only the count lives here.
 */
export function NotificationsProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const instanceId = useCurrentInstanceId()
  // In federation mode notifications belong to the selected instance — don't poll
  // until one is chosen (the proxy 400s without an instance, e.g. on the picker).
  const ready = isAuthenticated && (!IS_FEDERATION || instanceId != null)
  const [unreadCount, setUnreadCount] = useState(0)

  const refreshUnread = useCallback(async () => {
    try {
      const n = await notificationsHttpService.unreadCount()
      setUnreadCount(typeof n === 'number' && n >= 0 ? n : 0)
    } catch {
      // non-critical; leave the previous count
    }
  }, [])

  const markRead = useCallback(
    async (id: number, read: boolean) => {
      await notificationsHttpService.markRead(id, read)
      void refreshUnread()
    },
    [refreshUnread]
  )

  const markAllRead = useCallback(async () => {
    await notificationsHttpService.markAllRead()
    setUnreadCount(0)
  }, [])

  const remove = useCallback(
    async (id: number) => {
      await notificationsHttpService.remove(id)
      void refreshUnread()
    },
    [refreshUnread]
  )

  useEffect(() => {
    if (!ready) {
      setUnreadCount(0)
      return
    }
    void refreshUnread()
    const t = setInterval(() => void refreshUnread(), UNREAD_POLL_MS)
    return () => clearInterval(t)
  }, [ready, refreshUnread])

  const value = useMemo<NotificationsContextValue>(
    () => ({ unreadCount, refreshUnread, markRead, markAllRead, remove }),
    [unreadCount, refreshUnread, markRead, markAllRead, remove]
  )

  return <NotificationsContext.Provider value={value}>{children}</NotificationsContext.Provider>
}

export function useNotifications(): NotificationsContextValue {
  const ctx = useContext(NotificationsContext)
  if (!ctx) throw new Error('useNotifications must be used within NotificationsProvider')
  return ctx
}
