import { useTranslation } from 'react-i18next'
import { Check, Trash2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { TYPE_META, timeAgo } from '../lib'
import type { Notification } from '../types/notification.types'

interface NotificationRowProps {
  notification: Notification
  onToggleRead: (n: Notification) => void
  onDelete: (n: Notification) => void
}

export function NotificationRow({ notification: n, onToggleRead, onDelete }: NotificationRowProps) {
  const { t } = useTranslation()
  const meta = TYPE_META[n.type] ?? TYPE_META.INFO
  const Icon = meta.icon

  return (
    <div className="group relative flex gap-2.5 px-3 py-2.5 transition-colors hover:bg-muted/50">
      <span
        className={cn(
          'mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full',
          n.read ? 'bg-transparent' : 'bg-primary',
        )}
        aria-hidden
      />
      <Icon size={16} className={cn('mt-0.5 shrink-0', meta.tone)} />

      <div className="min-w-0 flex-1">
        <p className={cn('text-sm leading-snug', n.read ? 'text-muted-foreground' : 'text-foreground')}>
          {n.message}
        </p>
        <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className="font-medium">{n.source}</span>
          <span>·</span>
          <span>{timeAgo(n.createdAt)}</span>
        </div>
      </div>

      <div className="flex shrink-0 items-start gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
        <button
          type="button"
          title={n.read ? t('notifications.markUnread') : t('notifications.markRead')}
          onClick={() => onToggleRead(n)}
          className={cn(
            'rounded p-1 hover:bg-muted hover:text-foreground',
            n.read ? 'text-muted-foreground/60' : 'text-primary',
          )}
        >
          <Check size={14} />
        </button>
        <button
          type="button"
          title={t('notifications.delete')}
          onClick={() => onDelete(n)}
          className="rounded p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
        >
          <Trash2 size={14} />
        </button>
      </div>
    </div>
  )
}
