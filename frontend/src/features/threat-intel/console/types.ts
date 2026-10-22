
export interface IntelEntity {
  id: string
  type: string
  value: string
  reputation: number
  accuracy: number
  tags: string[]
  lastSeen?: string
}

export interface RelationNode {
  id: string
  type: string
  value: string
  reputation: number | null
}

export interface RelationEdge {
  source: string
  target: string
}

export interface RelationsResult {
  nodes: RelationNode[]
  edges: RelationEdge[]
}

export interface EntityDetailAttributes {
  id: string
  type: string
  value: string
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

export interface EntityGeo {
  lat: number
  lng: number
  radius: number
  ip?: string
}

export interface EntityDetails {
  attributes: EntityDetailAttributes
  metadata: Record<string, string> | null
  /** Grouped extended metadata, when the corpus carries it. */
  extendedMetadata?: Record<string, Record<string, string>> | null
  latest_associations: IntelEntity[]
  geolocations: EntityGeo[]
}

export interface EntityHistoryRecord {
  timestamp: string
  reputation: number
  author?: string
}

export type ToolPayload =
  | { kind: 'lookup'; value: string; entity: IntelEntity | null; rejected: string[] }
  | { kind: 'search'; query: string; items: number; results: IntelEntity[] }
  | { kind: 'details'; id: string; details: EntityDetails }
  | { kind: 'relations'; id: string; graph: RelationsResult }
  | { kind: 'geo'; id: string; locations: EntityGeo[] }
  | { kind: 'counts'; buckets: { key: string; count: number }[] }
  | { kind: 'report'; value: string; type: string; reputation: number | undefined; skippedAssociations: string[] }
  | { kind: 'link'; entityID: string; relatedEntityID: string }
  | { kind: 'cases'; cases: import('../lib/casework').Case[] }
  | { kind: 'watchlists'; watchlists: import('../lib/casework').Watchlist[] }
  | { kind: 'alerts'; alerts: import('../lib/casework').Alert[]; unread: number }
  | { kind: 'casework'; action: string; label: string; ref: string }

export interface AgentStep {
  id: string
  toolName: string
  args: Record<string, unknown>
  status: 'running' | 'done' | 'error'
  payload?: ToolPayload
  error?: string
  result?: string
  durationMs?: number
}

export interface LoopTurn {
  id: string
  question: string
  text: string
  steps: AgentStep[]
  status: 'running' | 'done' | 'error'
  error?: string
  stoppedReason?: 'complete' | 'max-turns' | 'aborted'
}

export interface Conversation {
  id: string
  title: string
  updatedAt: number
  turns: LoopTurn[]
  
  history: ChatMessage[]
}

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'tool'
  content: string | null
  tool_calls?: ModelToolCall[]
  tool_call_id?: string
}

export interface ModelToolCall {
  id: string
  type: 'function'
  function: { name: string; arguments: string }
}

export type PanelKind = 'entity' | 'graph' | 'geo' | 'search'

export interface PanelGeometry {
  x: number
  y: number
  width: number
  height: number
}

export interface CanvasPanel extends PanelGeometry {
  id: string
  kind: PanelKind
  title: string
  
  ref: string
  x: number
  y: number
  width: number
  height: number
  z: number
  minimised?: boolean
}

export interface PanelRequest {
  kind: PanelKind
  title: string
  ref: string
}

export const CANVAS_BREAKPOINT = 1024
export const DEFAULT_MODEL = 'silas-1.7-pro'
export const AVAILABLE_MODELS = [
  { id: 'silas-1.7-pro', label: 'Silas 1.7 Pro' },
  { id: 'silas-1.6-pro', label: 'Silas 1.6 Pro' },
  { id: 'silas-1.6', label: 'Silas 1.6' },
]
