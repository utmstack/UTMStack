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
  page?: number
  size?: number
}

export interface EntitySearchResponse {
  results: EntitySummary[]
  items: number
  pages: number
  aggregations?: unknown | null
}

export type FeedType = 'accumulative' | (string & {})
export type FeedAccuracy = 'level1' | 'level2' | 'level3' | (string & {})

export interface ThreatFeed {
  name: string
  type: FeedType
  accuracy: FeedAccuracy
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

export interface EndpointUsage {
  used: number
  limit: number
  resets_in_seconds: number
}

// GET /threat-intel/usage — keyed by rate-limited CM proxy path prefix
// (e.g. "/api/ai/v1/chat/completions"), see CustomersManager proxy/handlers/usage.go.
export type UsageInfo = Record<string, EndpointUsage>

export type TiResult<T> =
  | { kind: 'ok'; value: T }
  | { kind: 'not-configured' }

export interface AdvancedRangeCondition {
  range: Record<string, { gte?: string; lte?: string }>
}
export interface AdvancedTermCondition {
  term: Record<string, { value: string }>
}
export type AdvancedCondition = AdvancedRangeCondition | AdvancedTermCondition

export interface AdvancedDateHistogram {
  date_histogram: { field: string; interval: string }
}
export type AdvancedAggregation = AdvancedDateHistogram

export interface AdvancedSearchRequest {
  query?: {
    must?: AdvancedCondition[]
    should?: AdvancedCondition[]
    must_not?: AdvancedCondition[]
    filter?: AdvancedCondition[]
  }
  aggs?: Record<string, AdvancedAggregation>
}

export interface AggregationBucket {
  key: number
  key_as_string: string
  doc_count: number
}
export interface AggregationResult {
  buckets: AggregationBucket[]
}

export interface AdvancedSearchResponse {
  items: number
  pages: number
  results: EntitySearchItem[]
  aggregations?: Record<string, AggregationResult> | null
}
