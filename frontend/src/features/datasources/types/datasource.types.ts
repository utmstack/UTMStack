/* Mirrors the backend datasources module + ingestion-stats. */

export type SourceKind = 'agent' | 'puller' | 'direct'

export interface AssetGroupRef {
  id: number
  groupName: string
  groupDescription?: string
}

/** GET /datasources item / GET /datasources/:id */
export interface Datasource {
  id: number
  name: string
  dataType?: string
  ip?: string
  sourceKind?: SourceKind
  metadata?: Record<string, unknown>
  labels?: string // comma-separated
  group?: AssetGroupRef
  discoveredAt?: string
  modifiedAt?: string
  lastPingAt?: string // liveness; status is derived from its staleness
}

/** Common paginated envelope (total in body, not a header). */
export interface ListResponse<T> {
  items: T[]
  page_number: number
  page_size: number
  total_items: number
  total_pages: number
}

/** GET /datasources/usage */
export interface Usage {
  count: number
  limit: number // 0 when unlimited
  unlimited: boolean
  overLimit: boolean
}

/** GET /datasource-groups item */
export interface AssetGroup {
  id: number
  groupName: string
  groupDescription?: string
  createdDate?: string
  assetsCount: number
  metrics?: Record<string, number>
}

/** GET /eventprocessing/ingestion-stats (groupBy=dataSource) */
export interface IngestionBucket {
  key: string // dataSource name
  count: number
  lastSeen?: string
}
export interface IngestionTotals {
  groupBy: string
  status: string
  from: string
  to: string
  total: number
  buckets: IngestionBucket[]
}

/** GET /eventprocessing/ingestion-stats/timeline (no groupBy → single line) */
export interface TimelinePoint {
  timestamp: string
  count: number
}
export interface IngestionTimeline {
  status: string
  interval: string
  from: string
  to: string
  points?: TimelinePoint[]
}
