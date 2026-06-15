import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { WidgetCard } from '@/features/dashboard/components/WidgetCard'
import { WidgetRenderer } from '@/features/dashboard/components/WidgetRenderer'
import type {
  DashboardVisualization,
  GridLayoutItem,
  Visualization,
} from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

// Column span per width preset. Grid is responsive (1 col mobile, 2 tablet, 3
// desktop), so spans are clamped at each breakpoint. Literal classes so Tailwind
// keeps them.
const WIDTH_CLASS: Record<number, string> = {
  1: '',
  2: 'md:col-span-2 xl:col-span-2',
  3: 'md:col-span-2 xl:col-span-3',
}
const HEIGHT_CLASS: Record<number, string> = {
  1: 'h-[340px]',
  2: 'h-[560px]',
}

export function DashboardGrid({
  items,
  layouts,
  visualizationsById,
  time,
  editing,
  onMove,
  onResize,
  onRemoveItem,
}: {
  items: GridLayoutItem[]
  layouts: DashboardVisualization[]
  visualizationsById: Map<number, Visualization>
  time: TimeRange
  editing: boolean
  onMove?: (id: string, dir: -1 | 1) => void
  onResize?: (id: string, w: number, h: number) => void
  onRemoveItem?: (id: number) => void
}) {
  const { t } = useTranslation()

  const layoutMap = useMemo(() => {
    const map = new Map<string, DashboardVisualization>()
    for (const dv of layouts) map.set(String(dv.id), dv)
    return map
  }, [layouts])

  if (items.length === 0) {
    return (
      <div className="flex h-full min-h-[300px] w-full items-center justify-center rounded-lg border border-dashed border-border bg-card/50 px-6 py-12 text-center text-sm text-muted-foreground">
        {editing ? t('dashboards.grid.emptyEditing') : t('dashboards.grid.empty')}
      </div>
    )
  }

  return (
    // Responsive columns — adapts widgets-per-row to the screen. Each widget
    // spans 1–3 columns (its width preset) and is short or tall (height preset).
    <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-2 xl:grid-cols-3">
      {items.map((item, idx) => {
        const dv = layoutMap.get(item.i)
        const viz = dv ? visualizationsById.get(dv.idVisualization) : undefined
        const title = viz?.name ?? t('dashboards.grid.unknownVisualization')
        const w = item.w || 1
        const h = item.h || 1
        return (
          <div key={item.i} className={cn('min-w-0', WIDTH_CLASS[w] ?? '')}>
            <div className={cn('w-full', HEIGHT_CLASS[h] ?? HEIGHT_CLASS[1])}>
              <WidgetCard
                title={title}
                editing={editing}
                width={w}
                height={h}
                canMoveBack={idx > 0}
                canMoveForward={idx < items.length - 1}
                onMoveBack={() => onMove?.(item.i, -1)}
                onMoveForward={() => onMove?.(item.i, 1)}
                onResize={(nw, nh) => onResize?.(item.i, nw, nh)}
                onRemove={dv ? () => onRemoveItem?.(dv.id) : undefined}
              >
                {viz ? (
                  <WidgetRenderer visualization={viz} time={time} />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
                    {t('dashboards.grid.missingVisualization')}
                  </div>
                )}
              </WidgetCard>
            </div>
          </div>
        )
      })}
    </div>
  )
}
