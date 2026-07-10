// CM/ThreatWinds analytics response shapes. Wire the proxy responses through as-is.

export type EntityType =
  | 'ip'
  | 'hostname'
  | 'domain'
  | 'url'
  | 'hash'
  | 'cve'
  | 'threat'
  | (string & {})

export interface EntityAttributes {
  id: string
  type: EntityType
  value: string
  label: string
  description?: string
  tags: string[]
  first_seen: string
  last_seen: string
  reputation_score: number
  reputation: string
  best_reputation_score: number
  best_reputation: string
  worst_reputation_score: number
  worst_reputation: string
  accuracy_score: number
  accuracy: string
}

export interface EntityMetadata {
  asn: number
  aso: string
  city: string
  country: string
  latitude: number
  longitude: number
}

export interface EntityGeolocation {
  object: string
  country: string
  city: string
  latitude: number
  longitude: number
  accuracy_radius: number
  asn: number
  aso: string
}

export interface EntityDetail {
  attributes: EntityAttributes
  metadata: EntityMetadata
  extended_metadata: unknown[]
  latest_associations: EntityAttributes[]
  geolocations: EntityGeolocation[]
}

// Search results are a different CM shape than entity detail: camelCase, no label /
// description, value nested under `attributes[<type-key>]`, reputation is a number.
export interface EntitySearchItem {
  '@timestamp': string
  id: string
  type: EntityType
  attributes: Record<string, string>
  tags: string[]
  lastSeen: string
  reputation: number
  bestReputation: number
  worstReputation: number
  accuracy: number
  score: number
  version: number
  visibleBy: string[]
  wellKnown: boolean
  fields: unknown | null
  sort: unknown[]
}

export type EntitySummary = EntitySearchItem
// ponytail: relations endpoint shape unconfirmed — assumed EntityAttributes[] like
// latest_associations from detail. Narrow when a real response is observed.
export type EntityRelation = EntityAttributes

// Search items nest the display value inside `attributes[<type-key>]` (e.g.
// { type: 'hostname', attributes: { hostname: 'asas-asso.com' } }). Pick the
// first string value if the expected key isn't found.
export function searchItemValue(item: EntitySearchItem): string {
  return item.attributes[item.type] ?? Object.values(item.attributes)[0] ?? ''
}

export interface EntitySearchRequest {
  query: string
  types?: EntityType[]
  limit?: number
  offset?: number
}

export interface EntitySearchResponse {
  results: EntitySummary[]
  total: number
}

export type FeedKind = 'commercial' | 'open' | 'internal'
export type FeedStatus = 'healthy' | 'stale' | 'error' | 'paused'

export interface ThreatFeed {
  id: string
  name: string
  kind: FeedKind
  status: FeedStatus
  itemsTotal: number
  itemsAdded24h: number
  lastSync: string
  syncIntervalMin: number
}

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
}

export interface ChatRequest {
  messages: ChatMessage[]
  model?: string
}

export interface ChatResponse {
  message: ChatMessage
  usage?: { totalTokens: number }
}

export interface UsageInfo {
  used: number
  quota: number
  resetsAt?: string
}

export type TiResult<T> =
  | { kind: 'ok'; value: T }
  | { kind: 'not-configured' }
