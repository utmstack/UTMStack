import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import GridLayout, { useContainerWidth, type Layout } from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import { WidgetCard } from '@/features/dashboard/components/WidgetCard'
import { WidgetRenderer } from '@/features/dashboard/components/WidgetRenderer'
import { GRID_COLS, GRID_MARGIN, GRID_ROW_HEIGHT } from '@/features/dashboard/constants'
import type {
  DashboardVisualization,
  FilterType,
  GridLayoutItem,
  Visualization,
} from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export function DashboardGrid({
  items,
  layouts,
  visualizationsById,
  time,
  filters,
  editing,
  onLayoutChange,
  onRemoveItem,
}: {
  items: GridLayoutItem[]
  layouts: DashboardVisualization[]
  visualizationsById: Map<number, Visualization>
  time: TimeRange
  filters?: FilterType[]
  editing: boolean
  onLayoutChange?: (items: GridLayoutItem[]) => void
  onRemoveItem?: (id: number) => void
}) {
  const { t } = useTranslation()
  // v2 has no WidthProvider — it measures the container via this hook.
  const { width, containerRef } = useContainerWidth()

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
    <div ref={containerRef} className="w-full">
      {width > 0 && (
        <GridLayout
          width={width}
          layout={items}
          gridConfig={{ cols: GRID_COLS, rowHeight: GRID_ROW_HEIGHT, margin: GRID_MARGIN }}
          // Drag only by the card header; never start a drag from a control.
          dragConfig={{
            enabled: editing,
            handle: '.widget-drag-handle',
            cancel: 'button, select, a, input, textarea, .no-drag',
          }}
          resizeConfig={{ enabled: editing, handles: ['se'] }}
          onLayoutChange={
            editing
              ? (l: Layout) =>
                  onLayoutChange?.(
                    l.map((it) => ({ i: it.i, x: it.x, y: it.y, w: it.w, h: it.h }))
                  )
              : undefined
          }
        >
          {items.map((item) => {
            const dv = layoutMap.get(item.i)
            const viz = dv ? visualizationsById.get(dv.idVisualization) : undefined
            const title = viz?.name ?? t('dashboards.grid.unknownVisualization')
            return (
              <div key={item.i}>
                <WidgetCard
                  title={title}
                  editing={editing}
                  onRemove={dv ? () => onRemoveItem?.(dv.id) : undefined}
                >
                  {viz ? (
                    <WidgetRenderer visualization={viz} time={time} filters={filters} />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
                      {t('dashboards.grid.missingVisualization')}
                    </div>
                  )}
                </WidgetCard>
              </div>
            )
          })}
        </GridLayout>
      )}
    </div>
  )
}
