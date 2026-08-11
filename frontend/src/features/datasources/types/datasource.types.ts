/* Mirrors the backend datasources module + ingestion-stats. */

export type SourceKind = 'agent' | 'puller' | 'direct'

/** GET /datasources item / GET /datasources/:id */
export interface Datasource {
  id: string // uuid
  name: string
  dataType?: string
  ip?: string
  sourceKind?: SourceKind
  metadata?: Record<string, unknown>
  labels?: string // comma-separated
  /** Asset sensitivity (CIA, 0–3) — weights alert impact in the correlation engine. */
  assetConfidentiality?: number
  assetIntegrity?: number
  assetAvailability?: number
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

/** GET /datasources/count */
export interface DatasourceCount {
  count: number
}

/** GET /eventprocessing/ingestion-stats (groupBy=dataSource) */
export interface IngestionBucket {
  key: string // dataSource name
  count: number
  /** Ingested volume for this source, in bytes. */
  bytes?: number
  lastSeen?: string
}
export interface IngestionTotals {
  groupBy: string
  status: string
  from: string
  to: string
  total: number
  totalBytes?: number
  buckets: IngestionBucket[]
}

/** GET /eventprocessing/ingestion-stats/timeline (no groupBy → single line) */
export interface TimelinePoint {
  timestamp: string
  count: number
  bytes?: number
}
export interface IngestionTimeline {
  status: string
  interval: string
  from: string
  to: string
  points?: TimelinePoint[]
}
