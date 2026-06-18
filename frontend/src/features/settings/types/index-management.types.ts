/* Mirrors backend opensearch dtos (indices + index patterns / data views). */

/** A real OpenSearch index (GET /opensearch/index/all). */
export interface IndexInfo {
  index: string
  status: string // open / close
  health: string // green / yellow / red
  'docs.count': string
  'store.size': string
}

/** A data view (utm_index_pattern). */
export interface IndexPattern {
  id: number
  pattern: string
  patternModule: string | null
  patternSystem: boolean | null
  isActive: boolean | null
}

export interface IndexPatternUpsert {
  id?: number
  pattern: string
  patternModule?: string | null
  patternSystem?: boolean | null
  isActive?: boolean | null
}
