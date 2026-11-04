/* Mirrors the backend log-analyzer search contract. */

/** Subset of backend FilterType operators the UI actually produces. */
export type FilterOperator =
  | 'IS'
  | 'IS_NOT'
  | 'IS_BETWEEN'
  | 'IS_NOT_BETWEEN'
  | 'IS_IN_FIELDS'
  | 'IS_ONE_OF_TERMS' // used by the "related logs" deep-link (id in [...])
  | 'IS_NOT_ONE_OF_TERMS'
  | 'IS_GREATER_THAN'
  | 'IS_GREATER_THAN_OR_EQUALS'
  | 'IS_LESS_THAN'
  | 'IS_LESS_THAN_OR_EQUALS'
  | 'CONTAIN'
  | 'DOES_NOT_CONTAIN'
  | 'START_WITH'
  | 'DOES_NOT_START_WITH'
  | 'ENDS_WITH'
  | 'DOES_NOT_END_WITH'
  | 'EXIST'
  | 'DOES_NOT_EXIST'

/** A single search filter; a search body carries FilterType[]. */
export interface FilterType {
  field: string
  operator: FilterOperator
  value?: unknown
}

/** GET /log-analyzer/datasets/:dataset/fields row. */
export interface IndexField {
  name: string
  type: string // date | keyword | text | ip | long | integer | boolean | ...
}

/** A stored record, with whatever fields it carries. */
export type LogDocument = Record<string, unknown>

/** POST /log-analyzer/chart-view response. */
export interface ChartView {
  categories: string[]
  values: number[]
}

/** POST /log-analyzer/top-x-values/... response. */
export interface TopValues {
  total: number
  top: { value: string; count: number; percent: number }[]
}

/** Persisted-per-tab Log Explorer configuration. Plain-JSON-serializable. */
export interface LogExplorerTabConfig {
  id: string
  name: string
  dataset: string
  patternStr: string | null
  range: { from: string | null; to: string; interval: string }
  filters: FilterType[]
  columns: string[]
  searchInput: string
  appliedQuery: string
  sqlMode: boolean
  sqlInput: string
  appliedSql: string
  viewMode: 'table' | 'chart'
}
