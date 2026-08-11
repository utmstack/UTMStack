/* Mirrors the backend storage module: retention, what the store holds, and
 * where records go when they stop costing local disk. */

export type Dataset = 'logs' | 'alerts' | 'statistics'

export interface Retention {
  dataset: Dataset
  /** Whole lifetime: a record is deleted this many days after it happened. */
  keepDays: number
  /** When it moves to object storage. 0 means it never leaves local disk. */
  coldDays: number
  tiered: boolean
}

export interface DatasetUsage {
  dataset: Dataset
  documents: number
  bytes: number
  oldest?: string
  newest?: string
}

export interface StoreHealth {
  status: 'ok' | 'degraded' | 'unavailable'
  diskUsedPct: number
  message?: string
}

/** Configured means a bucket is written; ready means the store has it. Only
 *  ready allows a dataset to move records there. */
export interface Tiering {
  configured: boolean
  ready: boolean
  endpoint?: string
  policy?: string
}

export interface ObjectStoreInput {
  endpoint: string
  accessKey: string
  secretKey: string
  cacheBytes?: number
}
