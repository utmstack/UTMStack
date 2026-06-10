/* Mirrors the backend opensearch + log-analyzer search contract. */

/** Subset of backend FilterType operators the UI actually produces. */
export type FilterOperator =
  | 'IS'
  | 'IS_NOT'
  | 'IS_BETWEEN'
  | 'IS_IN_FIELDS'
  | 'CONTAIN'
  | 'EXIST'

/** A single search filter (POST /opensearch/search body is FilterType[]). */
export interface FilterType {
  field: string
  operator: FilterOperator
  value?: unknown
}

/** GET /opensearch/index-patterns row. */
export interface IndexPattern {
  id: number
  pattern: string
  patternModule: string | null
  patternSystem: boolean | null
  isActive: boolean | null
}

/** GET /opensearch/index/properties row. */
export interface IndexField {
  name: string
  type: string // date | keyword | text | ip | long | integer | boolean | ...
}

/** A raw OpenSearch _source document (arbitrary fields). */
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
