import type {
  ChartTypeId,
  FilterOperatorId,
  IntervalId,
  WidgetLayout,
} from '@/features/dashboard/types'

export const GRID_COLS = 12
export const GRID_ROW_HEIGHT = 50
export const GRID_MARGIN: [number, number] = [12, 12]

// Default widget size — wide and tall like the legacy dashboard cards (~590×420px
// on a 12-col grid), so charts have room to breathe instead of being cramped.
export const DEFAULT_WIDGET_LAYOUT: WidgetLayout = {
  x: 0,
  y: Number.POSITIVE_INFINITY,
  w: 4,
  h: 8,
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

export type ChartRenderer =
  | 'echarts'
  | 'table'
  | 'metric'
  | 'region_map'
  | 'text'

export interface ChartTypeMeta {
  id: ChartTypeId
  renderer: ChartRenderer
  icon: string
  defaultConfig: Record<string, unknown>
}

export const CHART_TYPES: ChartTypeMeta[] = [
  {
    id: 'bar',
    renderer: 'echarts',
    icon: 'bar',
    defaultConfig: {
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category' },
      yAxis: { type: 'value' },
      series: [{ type: 'bar', encode: { x: 0, y: 1 } }],
    },
  },
  {
    id: 'line',
    renderer: 'echarts',
    icon: 'line',
    defaultConfig: {
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category' },
      yAxis: { type: 'value' },
      series: [{ type: 'line', encode: { x: 0, y: 1 } }],
    },
  },
  {
    id: 'area',
    renderer: 'echarts',
    icon: 'area',
    defaultConfig: {
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category' },
      yAxis: { type: 'value' },
      series: [{ type: 'line', areaStyle: {}, encode: { x: 0, y: 1 } }],
    },
  },
  {
    id: 'pie',
    renderer: 'echarts',
    icon: 'pie',
    defaultConfig: {
      tooltip: { trigger: 'item' },
      series: [
        {
          type: 'pie',
          radius: '60%',
          center: ['50%', '45%'],
          encode: { itemName: 0, value: 1 },
          // No permanent labels/leader lines (they overlap badly with many
          // categories). Show the slice name in white only on hover.
          label: { show: false },
          labelLine: { show: false },
          emphasis: { label: { show: true, color: '#ffffff', fontWeight: 'bold' } },
        },
      ],
    },
  },
  {
    id: 'metric',
    renderer: 'metric',
    icon: 'metric',
    defaultConfig: {},
  },
  {
    id: 'table',
    renderer: 'table',
    icon: 'table',
    defaultConfig: {},
  },
  {
    id: 'list',
    renderer: 'table',
    icon: 'list',
    defaultConfig: {},
  },
  {
    id: 'region_map',
    renderer: 'region_map',
    icon: 'region_map',
    defaultConfig: {},
  },
  {
    id: 'text',
    renderer: 'text',
    icon: 'text',
    defaultConfig: {},
  },
]

export function getChartTypeMeta(id: ChartTypeId): ChartTypeMeta {
  return CHART_TYPES.find((c) => c.id === id) ?? CHART_TYPES[0]
}

// The bucket sizes a time chart offers, in the words the backend maps to a
// store interval. Empty is auto.
export const INTERVALS: IntervalId[] = ['', '1m', '5m', '15m', '1h', '1d', '1w']

export interface OperatorMeta {
  id: FilterOperatorId
  valueShape: 'none' | 'single' | 'pair' | 'list'
}

export const OPERATORS: OperatorMeta[] = [
  { id: 'IS', valueShape: 'single' },
  { id: 'IS_NOT', valueShape: 'single' },
  { id: 'IS_ONE_OF_TERMS', valueShape: 'list' },
  { id: 'IS_GREATER_THAN', valueShape: 'single' },
  { id: 'IS_LESS_THAN_OR_EQUALS', valueShape: 'single' },
  { id: 'IS_BETWEEN', valueShape: 'pair' },
  { id: 'CONTAIN', valueShape: 'single' },
  { id: 'DOES_NOT_CONTAIN', valueShape: 'single' },
  { id: 'START_WITH', valueShape: 'single' },
  { id: 'ENDS_WITH', valueShape: 'single' },
  { id: 'EXIST', valueShape: 'none' },
  { id: 'DOES_NOT_EXIST', valueShape: 'none' },
]

export function getOperatorMeta(id: FilterOperatorId): OperatorMeta {
  return OPERATORS.find((o) => o.id === id) ?? OPERATORS[0]
}

