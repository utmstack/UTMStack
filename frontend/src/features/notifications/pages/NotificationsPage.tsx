import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bell, CheckCheck, ChevronLeft, ChevronRight, Loader2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { useNotifications } from '../services/notifications.context'
import { notificationsHttpService } from '../services/notifications-http.service'
import { NotificationRow } from '../components/NotificationRow'
import type { Notification } from '../types/notification.types'

const PAGE_SIZE_OPTIONS = [10, 20, 50, 100]

export function NotificationsPage() {
  const { t } = useTranslation()
  const { markRead, remove, markAllRead, unreadCount, refreshUnread } = useNotifications()
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [items, setItems] = useState<Notification[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const load = useCallback(
    async (p: number) => {
      setLoading(true)
      setError(false)
      try {
        const { data, total } = await notificationsHttpService.listPaged({
          page: p,
          size: pageSize,
          status: 'ACTIVE',
        })
        setItems(data)
        setTotal(total)
      } catch {
        setError(true)
      } finally {
        setLoading(false)
      }
    },
    [pageSize],
  )

  useEffect(() => {
    void load(page)
  }, [page, load])

  const onPageSizeChange = (size: number) => {
    setPageSize(size)
    setPage(0)
  }

  const onToggleRead = async (n: Notification) => {
    setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: !n.read } : x)))
    try {
      await markRead(n.id, !n.read)
    } catch {
      setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: n.read } : x)))
    }
  }

  const onDelete = async (n: Notification) => {
    // Optimistic: drop the row immediately so there's no dead time / reload flash.
    const remaining = items.length - 1
    setItems((prev) => prev.filter((x) => x.id !== n.id))
    setTotal((t) => Math.max(0, t - 1))
    try {
      await remove(n.id)
      // If that emptied a non-first page, step back one (this reloads).
      if (remaining === 0 && page > 0) setPage((p) => p - 1)
    } catch {
      // Failed — re-sync the page from the server.
      void load(page)
    }
  }

  const onMarkAll = async () => {
    setItems((prev) => prev.map((x) => ({ ...x, read: true })))
    try {
      await markAllRead()
    } catch {
      void refreshUnread()
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pb-6 pt-3">
      <header className="flex items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-base font-semibold">
            <Bell size={16} strokeWidth={1.75} />
            {t('notifications.title')}
          </h1>
        </div>
        <Button variant="outline" size="sm" onClick={() => void onMarkAll()} disabled={unreadCount === 0}>
          <CheckCheck size={14} className="mr-1.5" />
          {t('notifications.markAllRead')}
        </Button>
      </header>

      <div className="relative mt-6 min-h-[180px] divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
        {loading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-card/60">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </div>
        )}

        {!loading && error && (
          <div className="px-4 py-12 text-center text-sm">
            <p className="text-muted-foreground">{t('notifications.loadFailed')}</p>
            <Button variant="outline" size="sm" className="mt-3" onClick={() => void load(page)}>
              {t('common.actions.retry')}
            </Button>
          </div>
        )}

        {!loading && !error && items.length === 0 && (
          <div className="px-4 py-16 text-center text-sm text-muted-foreground">
            {t('notifications.empty')}
          </div>
        )}

        {items.map((n) => (
          <NotificationRow
            key={n.id}
            notification={n}
            onToggleRead={(x) => void onToggleRead(x)}
            onDelete={(x) => void onDelete(x)}
          />
        ))}
      </div>

      {total > 0 && (
        <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span>{t('pagination.perPage')}</span>
            <select
              value={pageSize}
              onChange={(e) => onPageSizeChange(Number(e.target.value))}
              className="cursor-pointer rounded-md border border-border bg-card px-1.5 py-1 text-xs text-foreground outline-none focus:ring-1 focus:ring-ring"
            >
              {PAGE_SIZE_OPTIONS.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
            <span className="ml-1">{t('pagination.pageOf', { page: page + 1, total: totalPages })}</span>
          </div>
          {totalPages > 1 && (
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page === 0 || loading}
                onClick={() => setPage((p) => Math.max(0, p - 1))}
              >
                <ChevronLeft size={14} className="mr-1" />
                {t('pagination.previous')}
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages - 1 || loading}
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              >
                {t('pagination.next')}
                <ChevronRight size={14} className="ml-1" />
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
