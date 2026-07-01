export interface Dashboard {
  id: number
  name: string
  description?: string
  config?: string
  filters?: string
  systemOwner?: boolean
  createdDate?: string
  modifiedDate?: string
}

export interface Visualization {
  id: number
  name: string
  description?: string
  sqlQuery: string
  config: string
  systemOwner?: boolean
  createdDate?: string
  modifiedDate?: string
}

export interface DashboardVisualization {
  id: number
  idDashboard: number
  idVisualization: number
  layout: string
}

export interface DashboardCreateInput {
  name: string
  description?: string
  config?: string
  filters?: string
}

export interface DashboardUpdateInput {
  id: number
  name: string
  description?: string
  config?: string
  filters?: string
}

export interface VisualizationCreateInput {
  name: string
  description?: string
  sqlQuery: string
  config: string
}

export interface VisualizationUpdateInput {
  id: number
  name: string
  description?: string
  sqlQuery: string
  config: string
}

export interface DashboardLayoutCreateInput {
  idDashboard: number
  idVisualization: number
  layout: string
}

export interface DashboardLayoutUpdateInput {
  id: number
  idDashboard: number
  idVisualization: number
  layout: string
}

export interface DashboardListParams {
  name?: string
  page?: number
  size?: number
}

export interface VisualizationListParams {
  name?: string
  page?: number
  size?: number
}

export interface DashboardLayoutListParams {
  idDashboard?: number
  idVisualization?: number
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
  | 'horizontal_bar'
  | 'line'
  | 'area'
  | 'pie'
  | 'gauge'
  | 'goal'
  | 'heatmap'
  | 'tag_cloud'
  | 'metric'
  | 'table'
  | 'list'
  | 'region_map'
  | 'text'

export interface IndexProperty {
  name: string
  type: string
}

export type AggregationId = 'count' | 'count_distinct' | 'sum' | 'avg' | 'min' | 'max'

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

// Runtime shape sent to the backend on /opensearch/search/sql. Mirrors the Go
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
  indexPattern: string
  multiple: boolean
  searchable: boolean
}

export interface BuilderMetric {
  agg: AggregationId
  field: string | null
}

export interface BuilderState {
  chartType: ChartTypeId
  indexPattern: string
  rawMode: boolean
  filters: FilterRow[]
  metric: BuilderMetric
  dimension: string | null
  /** Table chart: explicit columns to SELECT. Empty → all columns (SELECT *). */
  columns: string[]
  advancedSelect: string | null
  rawSql: string | null
  configTouched: boolean
}

export interface IndexPattern {
  id: number
  pattern: string
  patternModule?: string | null
  patternSystem?: boolean | null
  isActive?: boolean | null
}

export interface IndexPatternListParams {
  isActive?: boolean
  page?: number
  size?: number
  sort?: string
}
