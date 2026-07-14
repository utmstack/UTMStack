import {
  BarChart3,
  LineChart,
  AreaChart,
  PieChart,
  Hash,
  List as ListIcon,
  Map,
  Type,
  Table as TableIcon,
  type LucideIcon,
} from 'lucide-react'
import { getChartTypeMeta } from '@/features/dashboard/constants'
import type { ChartTypeId } from '@/features/dashboard/types'

// Maps a chart type's `icon` slug (see CHART_TYPES) to a lucide icon component.
export const CHART_ICONS: Record<string, LucideIcon> = {
  bar: BarChart3,
  line: LineChart,
  area: AreaChart,
  pie: PieChart,
  metric: Hash,
  table: TableIcon,
  list: ListIcon,
  region_map: Map,
  text: Type,
}

// Resolves the icon for a chart type id (falls back to the bar icon).
export function getChartIcon(chartType: ChartTypeId | string | null | undefined): LucideIcon {
  if (!chartType) return BarChart3
  return CHART_ICONS[getChartTypeMeta(chartType as ChartTypeId).icon] ?? BarChart3
}
