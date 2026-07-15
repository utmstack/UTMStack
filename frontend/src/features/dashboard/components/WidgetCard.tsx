import { type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { GripVertical, Pencil, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export function WidgetCard({
  title,
  editing,
  onEdit,
  onRemove,
  children,
}: {
  title: string
  editing: boolean
  onEdit?: () => void
  onRemove?: () => void
  children: ReactNode
}) {
  const { t } = useTranslation()

  return (
    <div className="flex h-full w-full flex-col overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      {/* The header is the drag handle (`.widget-drag-handle`) while editing. */}
      <div
        className={cn(
          'widget-drag-handle flex items-center justify-between gap-2 border-b border-border px-3 py-1.5',
          editing ? 'cursor-move bg-muted/40' : 'bg-card'
        )}
      >
        <div className="flex min-w-0 items-center gap-1.5">
          {editing && <GripVertical size={14} className="shrink-0 text-muted-foreground" />}
          <span className="truncate text-sm font-medium text-foreground" title={title}>
            {title}
          </span>
        </div>
        {editing && (onEdit || onRemove) && (
          <div className="flex shrink-0 items-center gap-1">
            {onEdit && (
              <button
                type="button"
                onClick={onEdit}
                className="no-drag flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                aria-label={t('dashboards.widget.editAriaLabel') ?? 'Edit widget'}
                title={t('dashboards.widget.editAriaLabel') ?? 'Edit widget'}
              >
                <Pencil size={13} />
              </button>
            )}
            {onRemove && (
              <button
                type="button"
                onClick={onRemove}
                className="no-drag flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                aria-label={t('dashboards.widget.removeAriaLabel') ?? 'Remove widget'}
                title={t('dashboards.widget.removeAriaLabel') ?? 'Remove widget'}
              >
                <X size={14} />
              </button>
            )}
          </div>
        )}
      </div>
      <div className="min-h-0 flex-1 overflow-hidden p-2">{children}</div>
    </div>
  )
}
