import { type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export function WidgetCard({
  title,
  editing,
  onRemove,
  children,
}: {
  title: string
  editing: boolean
  onRemove?: () => void
  children: ReactNode
}) {
  const { t } = useTranslation()
  return (
    <div className="flex h-full w-full flex-col overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div
        className={cn(
          'widget-drag-handle flex items-center justify-between gap-2 border-b border-border px-3 py-2',
          editing ? 'cursor-move bg-muted/40' : 'bg-card'
        )}
      >
        <span className="truncate text-sm font-medium text-foreground" title={title}>
          {title}
        </span>
        {editing && onRemove && (
          <button
            type="button"
            onMouseDown={(e) => e.stopPropagation()}
            onClick={(e) => {
              e.stopPropagation()
              onRemove()
            }}
            className="flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            aria-label={t('dashboards.widget.removeAriaLabel') ?? 'Remove widget'}
          >
            <X size={14} />
          </button>
        )}
      </div>
      <div className="min-h-0 flex-1 overflow-hidden p-2">{children}</div>
    </div>
  )
}
