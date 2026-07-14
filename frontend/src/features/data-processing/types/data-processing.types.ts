/* Mirrors backend eventprocessing DTOs (modules/eventprocessing/dto). */

/** A YAML event-processing filter file. Identity = relPath. */
export interface Filter {
  relPath: string
  content: string
  system: boolean
  active: boolean
  /** The dataTypes this filter's (single) pipeline entry applies to — reusable across several. */
  dataTypes: string[]
  /** Position in the global pipeline order (system band no longer enforced). */
  order: number
}

export interface FilterListQuery {
  relPathContains?: string
  isActive?: boolean
  system?: boolean
  dataType?: string
  page?: number // 1-based
  size?: number
}

/** Catalog item from GET /integrations/data-types — known dataTypes to pick from. */
export interface DataTypeOption {
  dataType: string
  name: string
  moduleName: string
  active: boolean
  isSystem: boolean
}

export interface SaveFilterRequest {
  relPath: string
  content: string
}

/* ── Ingestion stats ── */

export type IngestionStatus = 'received' | 'parsing_dropped' | 'analysis_dropped' | 'correlation_dropped' | 'all'
export type GroupBy = 'dataSource' | 'dataType'

export interface IngestionBucket {
  key: string
  count: number
  lastSeen?: string
}

export interface IngestionStats {
  groupBy: string
  status: string
  from: string
  to: string
  total: number
  buckets: IngestionBucket[]
}

export interface TimelinePoint {
  timestamp: string
  count: number
}

export interface TimelineSeries {
  key: string
  points: TimelinePoint[]
}

export interface IngestionTimeline {
  status: string
  groupBy?: string
  interval: string
  from: string
  to: string
  points?: TimelinePoint[]
  series?: TimelineSeries[]
}

export interface IngestionQuery {
  groupBy?: GroupBy
  status?: IngestionStatus
  from?: string
  to?: string
  interval?: string
  top?: number
}
