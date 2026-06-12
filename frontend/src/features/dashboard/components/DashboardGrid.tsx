import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Responsive, WidthProvider, type Layout } from 'react-grid-layout/legacy'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { WidgetCard } from '@/features/dashboard/components/WidgetCard'
import { WidgetRenderer } from '@/features/dashboard/components/WidgetRenderer'
import { GRID_COLS, GRID_MARGIN, GRID_ROW_HEIGHT } from '@/features/dashboard/constants'
import type {
  DashboardVisualization,
  GridLayoutItem,
  Visualization,
} from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

const ResponsiveGridLayout = WidthProvider(Responsive)

const BREAKPOINTS = { lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }
const COLS = { lg: GRID_COLS, md: GRID_COLS, sm: 6, xs: 4, xxs: 2 }

export function DashboardGrid({
  items,
  layouts,
  visualizationsById,
  time,
  editing,
  onLayoutChange,
  onRemoveItem,
}: {
  items: GridLayoutItem[]
  layouts: DashboardVisualization[]
  visualizationsById: Map<number, Visualization>
  time: TimeRange
  editing: boolean
  onLayoutChange?: (next: GridLayoutItem[]) => void
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
    <ResponsiveGridLayout
      className="layout"
      layouts={{ lg: items, md: items, sm: items, xs: items, xxs: items }}
      breakpoints={BREAKPOINTS}
      cols={COLS}
      rowHeight={GRID_ROW_HEIGHT}
      margin={GRID_MARGIN}
      compactType="vertical"
      isDraggable={editing}
      isResizable={editing}
      draggableHandle=".widget-drag-handle"
      onLayoutChange={(next: Layout) => {
        if (!editing || !onLayoutChange) return
        const mapped: GridLayoutItem[] = next.map((n) => ({
          i: String(n.i),
          x: n.x,
          y: n.y,
          w: n.w,
          h: n.h,
        }))
        onLayoutChange(mapped)
      }}
    >
      {items.map((item) => {
        const dv = layoutMap.get(item.i)
        const viz = dv ? visualizationsById.get(dv.idVisualization) : undefined
        const title = viz?.name ?? t('dashboards.grid.unknownVisualization')
        const layoutId = dv?.id
        return (
          <div key={item.i} data-grid={item}>
            <WidgetCard
              title={title}
              editing={editing}
              onRemove={layoutId ? () => onRemoveItem?.(layoutId) : undefined}
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
        )
      })}
    </ResponsiveGridLayout>
  )
}
