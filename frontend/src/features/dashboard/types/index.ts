export interface Dashboard {
  id: string
  name: string
  description?: string
  config?: string
  filters?: string
  systemOwner?: boolean
  createdDate?: string
  modifiedDate?: string
}

export interface Visualization {
  id: string
  // The one dashboard this visualization belongs to — not reusable elsewhere.
  dashboardId: string
  /** The question the widget asks, as a JSON {@link VizSpec}. */
  spec: string
  config: string
  // Grid position/size (JSON) on its dashboard.
  layout: string
  systemOwner?: boolean
  createdDate?: string
  modifiedDate?: string
}

export interface DashboardCreateInput {
  name: string
  description?: string
  config?: string
  filters?: string
}

export interface DashboardUpdateInput {
  id: string
  name: string
  description?: string
  config?: string
  filters?: string
}

export interface VisualizationCreateInput {
  dashboardId: string
  spec: string
  config: string
  layout: string
}

export interface VisualizationUpdateInput {
  id: string
  dashboardId: string
  spec: string
  config: string
  layout: string
}

export interface DashboardListParams {
  name?: string
  page?: number
  size?: number
}

export interface VisualizationListParams {
  dashboardId?: string
  page?: number
  size?: number
}

export interface GridLayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
}

export interface WidgetLayout {
  x: number
  y: number
  w: number
  h: number
  order?: number
}

export type ChartTypeId =
  | 'bar'
  | 'line'
  | 'area'
  | 'pie'
  | 'metric'
  | 'table'
  | 'list'
  | 'region_map'
  | 'text'

export interface IndexProperty {
  name: string
  type: string
}

export type FilterOperatorId =
  | 'IS'
  | 'IS_NOT'
  | 'IS_ONE_OF_TERMS'
  | 'IS_GREATER_THAN'
  | 'IS_LESS_THAN_OR_EQUALS'
  | 'IS_BETWEEN'
  | 'CONTAIN'
  | 'DOES_NOT_CONTAIN'
  | 'START_WITH'
  | 'ENDS_WITH'
  | 'EXIST'
  | 'DOES_NOT_EXIST'

export interface FilterRow {
  id: string
  field: string
  operator: FilterOperatorId
  value: string | string[] | [string, string] | null
}

// Runtime shape sent to the backend with a query. Mirrors the Go
// common_models.FilterType — field + operator + value.
export interface FilterType {
  field: string
  operator: FilterOperatorId
  value: unknown
}

// Persistent chip config stored on dashboard.filters (JSON blob). Describes a
// single dropdown in the dashboard filter bar; the *value* the user picks is
// session-only (not persisted).
export interface DashboardFilterChip {
  id: string
  field: string
  label: string
  placeholder?: string
  dataset: string
  multiple: boolean
  searchable: boolean
}

/**
 * What the editor holds while a widget is being built. It is the spec plus the
 * things only the editor cares about — which chart draws it, and whether the
 * ECharts config has been hand-edited — and it is stored alongside the chart
 * config so reopening the editor starts where it left off.
 */
export interface BuilderState {
  chartType: ChartTypeId
  dataset: string
  /** What the x axis is: buckets of time, or the top values of a field. */
  breakdown: BreakdownMode
  /** Category charts: the field broken down by. Time charts: the optional split. */
  dimension: string | null
  /** Time charts: bucket size. Empty means auto. */
  interval: IntervalId
  /** How many buckets, series or rows to ask for. */
  limit: number | null
  filters: FilterRow[]
  /** Table charts: the columns shown. Empty → whatever the records carry. */
  columns: string[]
  configTouched: boolean
}

export type BreakdownMode = 'time' | 'field'

export type IntervalId = '' | '1m' | '5m' | '15m' | '1h' | '1d' | '1w'

/** One row of an answer, as the renderers and the ECharts dataset read it. */
export type Row = Record<string, unknown>

/**
 * The question a widget asks, sent to POST /visualizations/query and stored on
 * the visualization. Mirrors the Go dashboards domain.Spec — the tenant is
 * never part of it, it comes from the session.
 */
export interface VizSpec {
  dataset: string
  dataType?: string
  chart: SpecChart
  metric: SpecMetric
  dimension?: string
  columns?: string[]
  filters?: SpecFilter[]
  interval?: string
  limit?: number
  from?: string
  to?: string
}

export type SpecChart = 'metric' | 'category' | 'time' | 'table'

/**
 * The event store counts records and does nothing else — no sum, average or
 * cardinality — so count is the only measure there is. The backend refuses any
 * other rather than answering it with a count.
 */
export interface SpecMetric {
  agg: 'count'
}

export type SpecOp =
  | 'eq'
  | 'not_eq'
  | 'in'
  | 'not_in'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'
  | 'between'
  | 'not_between'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'not_starts_with'
  | 'ends_with'
  | 'not_ends_with'
  | 'exists'
  | 'not_exists'

export interface SpecFilter {
  field: string
  op: SpecOp
  value?: unknown
}

/** What POST /visualizations/query answers, one field per chart shape. */
export interface QueryResult {
  total?: number
  buckets?: { key: string; count: number }[]
  points?: { at: string; count: number }[]
  series?: { key: string; points: { at: string; count: number }[] }[]
  rows?: Record<string, unknown>[]
}

