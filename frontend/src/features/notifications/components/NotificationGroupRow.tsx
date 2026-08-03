import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { TYPE_META, timeAgo } from '../lib'
import type { NotificationGroup } from '../types/notification.types'
import { NotificationGroupItems } from './NotificationGroupItems'

interface NotificationGroupRowProps {
  group: NotificationGroup
  /** Notified when a nested item is toggled/deleted so the parent can refresh counts. */
  onChanged?: () => void
  /** Height cap for the nested infinite-scroll list (Tailwind max-h-* class). */
  itemsMaxHeightClass?: string
}

export function NotificationGroupRow({
  group,
  onChanged,
  itemsMaxHeightClass,
}: NotificationGroupRowProps) {
  const [open, setOpen] = useState(false)
  const meta = TYPE_META[group.type] ?? TYPE_META.INFO
  const Icon = meta.icon
  const expandable = group.count > 1
  const hasUnread = group.unreadCount > 0

  return (
    <div>
      <div
        className={cn(
          'flex gap-2.5 px-3 py-2.5 m-2 transition-colors rounded-md border-border',
          expandable ? 'cursor-pointer hover:bg-muted/50' : 'hover:bg-muted/30',
          open ? 'border-b-0 border-t-0 border-x-0':'border'
        )}
        onClick={expandable ? () => setOpen((v) => !v) : undefined}
        role={expandable ? 'button' : undefined}
        tabIndex={expandable ? 0 : undefined}
        onKeyDown={
          expandable
            ? (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault()
                  setOpen((v) => !v)
                }
              }
            : undefined
        }
      >
        <span
          className={cn(
            'mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full',
            hasUnread ? 'bg-primary' : 'bg-transparent',
          )}
          aria-hidden
        />
        <Icon size={16} className={cn('mt-0.5 shrink-0', meta.tone)} />

        <div className="min-w-0 flex-1">
          <p
            className={cn(
              'text-sm leading-snug',
              hasUnread ? 'text-foreground' : 'text-muted-foreground',
            )}
          >
            {group.message}
          </p>
          <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <span className="font-medium">{group.source}</span>
            <span>·</span>
            <span>{timeAgo(group.lastCreated)}</span>
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          {group.count > 1 && (
            <span className="rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-semibold text-muted-foreground">
              {group.count}
            </span>
          )}
          {hasUnread && (
            <span className="rounded-full bg-primary/10 px-1.5 py-0.5 text-[10px] font-semibold text-primary">
              {group.unreadCount}
            </span>
          )}
          {expandable && (
            <ChevronDown
              size={14}
              className={cn(
                'text-muted-foreground transition-transform',
                open ? 'rotate-180' : 'rotate-0',
              )}
            />
          )}
        </div>
      </div>

      {expandable && open && (
        <NotificationGroupItems
          source={group.source}
          type={group.type}
          message={group.message}
          onItemChanged={onChanged}
          maxHeightClass={itemsMaxHeightClass}
        />
      )}
    </div>
  )
}
