import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { useNotificationFeed } from '../hooks/useNotificationFeed'
import { useNotifications } from '../services/notifications.context'
import type { Notification, NotificationType } from '../types/notification.types'
import { NotificationRow } from './NotificationRow'

interface NotificationGroupItemsProps {
  source: string
  type: NotificationType
  message: string
  pageSize?: number
  maxHeightClass?: string
  onItemChanged?: () => void
}

export function NotificationGroupItems({
  source,
  type,
  message,
  pageSize = 10,
  maxHeightClass = 'max-h-72',
  onItemChanged,
}: NotificationGroupItemsProps) {
  const { t } = useTranslation()
  const { markRead, remove, refreshUnread } = useNotifications()
  const { items, setItems, loading, hasMore, error, loadMore } = useNotificationFeed(pageSize, {
    source,
    type,
    message,
  })

  useEffect(() => {
    void loadMore()
  }, [loadMore])

  const onScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget
    if (el.scrollHeight - el.scrollTop - el.clientHeight < 60 && hasMore && !loading) {
      void loadMore()
    }
  }

  const toggleRead = async (n: Notification) => {
    setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: !n.read } : x)))
    try {
      await markRead(n.id, !n.read)
      onItemChanged?.()
    } catch {
      setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: n.read } : x)))
    }
  }

  const deleteItem = async (n: Notification) => {
    setItems((prev) => prev.filter((x) => x.id !== n.id))
    try {
      await remove(n.id)
      onItemChanged?.()
    } catch {
      void refreshUnread()
    }
  }

  return (
    <div onScroll={onScroll} className={`${maxHeightClass} pl-2 divide-y divide-border overflow-y-auto bg-muted/20  `}>
      {items.map((n) => (
        <NotificationRow
          key={n.id}
          notification={n}
          onToggleRead={(x) => void toggleRead(x)}
          onDelete={(x) => void deleteItem(x)}
        />
      ))}
      {loading && (
        <div className="flex items-center justify-center py-3 text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      )}
      {error && !loading && items.length === 0 && (
        <div className="px-3 py-4 text-center text-xs text-destructive">
          {t('notifications.loadFailed')}
        </div>
      )}
    </div>
  )
}
