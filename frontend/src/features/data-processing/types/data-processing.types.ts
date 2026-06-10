/* Mirrors backend eventprocessing DTOs (modules/eventprocessing/dto). */

/** A YAML event-processing filter file. Identity = relPath. */
export interface Filter {
  relPath: string
  content: string
  system: boolean
  active: boolean
}

export interface FilterListQuery {
  relPathContains?: string
  isActive?: boolean
  system?: boolean
  page?: number // 1-based
  size?: number
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
