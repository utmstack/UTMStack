import { type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronLeft, ChevronRight, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export function WidgetCard({
  title,
  editing,
  canMoveBack,
  canMoveForward,
  width,
  height,
  onMoveBack,
  onMoveForward,
  onResize,
  onRemove,
  children,
}: {
  title: string
  editing: boolean
  canMoveBack?: boolean
  canMoveForward?: boolean
  width?: number
  height?: number
  onMoveBack?: () => void
  onMoveForward?: () => void
  onResize?: (w: number, h: number) => void
  onRemove?: () => void
  children: ReactNode
}) {
  const { t } = useTranslation()
  const w = width ?? 1
  const h = height ?? 1
  const selectCls =
    'h-6 rounded-md border border-border bg-background px-1 text-[11px] text-foreground/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

  return (
    <div className="flex h-full w-full flex-col overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div
        className={cn(
          'flex items-center justify-between gap-2 border-b border-border px-3 py-1.5',
          editing ? 'bg-muted/40' : 'bg-card'
        )}
      >
        <span className="truncate text-sm font-medium text-foreground" title={title}>
          {title}
        </span>
        {editing && (
          <div className="flex items-center gap-0.5">
            <button
              type="button"
              onClick={onMoveBack}
              disabled={!canMoveBack}
              className="flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30 disabled:hover:bg-transparent"
              aria-label={t('dashboards.widget.moveBack') ?? 'Move back'}
              title={t('dashboards.widget.moveBack') ?? 'Move back'}
            >
              <ChevronLeft size={15} />
            </button>
            <button
              type="button"
              onClick={onMoveForward}
              disabled={!canMoveForward}
              className="flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30 disabled:hover:bg-transparent"
              aria-label={t('dashboards.widget.moveForward') ?? 'Move forward'}
              title={t('dashboards.widget.moveForward') ?? 'Move forward'}
            >
              <ChevronRight size={15} />
            </button>

            {/* Size presets: width (columns) and height. */}
            <select
              value={w}
              onChange={(e) => onResize?.(Number(e.target.value), h)}
              className={cn(selectCls, 'ml-1')}
              title={t('dashboards.widget.width') ?? 'Width'}
              aria-label={t('dashboards.widget.width') ?? 'Width'}
            >
              <option value={1}>{t('dashboards.widget.widthSmall') ?? '1×'}</option>
              <option value={2}>{t('dashboards.widget.widthMedium') ?? '2×'}</option>
              <option value={3}>{t('dashboards.widget.widthLarge') ?? '3×'}</option>
            </select>
            <select
              value={h}
              onChange={(e) => onResize?.(w, Number(e.target.value))}
              className={selectCls}
              title={t('dashboards.widget.height') ?? 'Height'}
              aria-label={t('dashboards.widget.height') ?? 'Height'}
            >
              <option value={1}>{t('dashboards.widget.heightNormal') ?? 'Normal'}</option>
              <option value={2}>{t('dashboards.widget.heightTall') ?? 'Tall'}</option>
            </select>

            {onRemove && (
              <button
                type="button"
                onClick={onRemove}
                className="ml-0.5 flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
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
