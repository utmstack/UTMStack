import { DEFAULT_WIDGET_LAYOUT } from '@/features/dashboard/constants'
import type {
  DashboardVisualization,
  GridLayoutItem,
  WidgetLayout,
} from '@/features/dashboard/types'

export function parseLayout(raw: string | undefined | null): WidgetLayout {
  if (!raw) return { ...DEFAULT_WIDGET_LAYOUT }
  try {
    const parsed = JSON.parse(raw) as Partial<WidgetLayout>
    return {
      x: Number.isFinite(parsed.x) ? Number(parsed.x) : DEFAULT_WIDGET_LAYOUT.x,
      y: Number.isFinite(parsed.y) ? Number(parsed.y) : DEFAULT_WIDGET_LAYOUT.y,
      w: Number.isFinite(parsed.w) ? Number(parsed.w) : DEFAULT_WIDGET_LAYOUT.w,
      h: Number.isFinite(parsed.h) ? Number(parsed.h) : DEFAULT_WIDGET_LAYOUT.h,
      order: typeof parsed.order === 'number' ? parsed.order : undefined,
    }
  } catch {
    return { ...DEFAULT_WIDGET_LAYOUT }
  }
}

export function serializeLayout(layout: WidgetLayout): string {
  return JSON.stringify(layout)
}

export function toGridItems(rows: DashboardVisualization[]): GridLayoutItem[] {
  return rows.map((dv) => {
    const { x, y, w, h } = parseLayout(dv.layout)
    return { i: String(dv.id), x, y, w, h }
  })
}

export function fromGridItem(item: GridLayoutItem, order?: number): WidgetLayout {
  return { x: item.x, y: item.y, w: item.w, h: item.h, order }
}
