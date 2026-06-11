import type { WidgetLayout } from '@/features/dashboard/types'

export const GRID_COLS = 12
export const GRID_ROW_HEIGHT = 50
export const GRID_MARGIN: [number, number] = [12, 12]

export const DEFAULT_WIDGET_LAYOUT: WidgetLayout = {
  x: 0,
  y: Number.POSITIVE_INFINITY,
  w: 4,
  h: 4,
}

export const CHART_TYPE_META: Record<string, { label: string }> = {
  bar: { label: 'Bar' },
  line: { label: 'Line' },
  pie: { label: 'Pie' },
  area: { label: 'Area' },
  scatter: { label: 'Scatter' },
  gauge: { label: 'Gauge' },
  table: { label: 'Table' },
  metric: { label: 'Metric' },
}

export const DEFAULT_PAGE_SIZE = 20
