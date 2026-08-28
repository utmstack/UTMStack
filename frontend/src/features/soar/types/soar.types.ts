/* Mirrors backend modules/soar dto (rule.go, execution.go). */

export type SoarOperator =
  | 'IS'
  | 'IS_NOT'
  | 'CONTAINS'
  | 'NOT_CONTAINS'
  | 'EXISTS'
  | 'NOT_EXISTS'
  | 'START_WITH'
  | 'NOT_START_WITH'
  | 'ENDS_WITH'
  | 'NOT_ENDS_WITH'
  | 'IS_ONE_OF'
  | 'IS_NOT_ONE_OF'

export const SOAR_OPERATORS: SoarOperator[] = [
  'IS',
  'IS_NOT',
  'CONTAINS',
  'NOT_CONTAINS',
  'EXISTS',
  'NOT_EXISTS',
  'START_WITH',
  'NOT_START_WITH',
  'ENDS_WITH',
  'NOT_ENDS_WITH',
  'IS_ONE_OF',
  'IS_NOT_ONE_OF',
]

export const SOAR_MULTI_VALUE_OPERATORS: SoarOperator[] = ['IS_ONE_OF', 'IS_NOT_ONE_OF']
export const SOAR_NO_VALUE_OPERATORS: SoarOperator[] = ['EXISTS', 'NOT_EXISTS']

export interface FlowCondition {
  operator: SoarOperator
  field: string
  value: unknown
}

export type NodeKind = 'executor' | 'enrichment'
export const NODE_KINDS: NodeKind[] = ['executor', 'enrichment']

/** One DAG node. Matches backend domain.FlowNode / dto.FlowNodeVM. */
export interface FlowNode {
  kind: NodeKind
  executor: string
  command?: string
  shell?: string
  /** OS platform (linux/windows/…) — used to filter the agent picker for
   *  shell nodes; not sent to the runtime. Optional. */
  platform?: string
  /** Hostname of the endpoint agent to run on. When empty, the shell
   *  executor defaults to the alert's dataSource (host that raised the
   *  alert). */
  agent?: string
  /** Hostnames that must not run this node when it's in auto-resolve mode
   *  (no explicit `agent`). Ignored when `agent` is set. */
  excludedAgents?: string[]
  params?: unknown // json.RawMessage on the wire; a plain object here for the editor
  onSuccess?: string[]
  onError?: string[]
}

/** An alert-response flow (file-backed YAML). Identity is `relPath`. */
export interface Flow {
  relPath: string
  name: string
  description: string
  conditions: FlowCondition[]
  roots: string[]
  nodes: Record<string, FlowNode>
  maxDepth?: number
  active: boolean
  systemOwner: boolean
  lastModifiedDate?: string | null
}

/** Create/Update payload (POST /soar/rules, PUT /soar/rules/:relPath). */
export interface SaveFlowInput {
  name: string
  description: string
  conditions: FlowCondition[]
  roots: string[]
  nodes: Record<string, FlowNode>
  maxDepth?: number
  active: boolean
}

export interface FlowListQuery {
  name?: string
  active?: boolean
  systemOwner?: boolean
  page?: number // 0-based
  size?: number
}

/** DAG execution status set. Waiting/Executing/Dead are new in the tree engine. */
export type ExecutionStatus = 'WAITING' | 'PENDING' | 'EXECUTING' | 'EXECUTED' | 'FAILED' | 'DEAD'
export type NonExecutionCause = 'AGENT_OFFLINE' | 'AGENT_NOT_FOUND' | 'MAX_DEPTH_EXCEEDED' | 'UNKNOWN'

/** What raised a command. */
export type ExecutionOrigin = 'FLOW' | 'MANUAL'

export interface Execution {
  id: string // uuid
  origin: ExecutionOrigin
  /** Set when origin is FLOW. */
  rulePath?: string
  alertId?: string
  /** Set when origin is MANUAL: who ran it. */
  triggeredBy?: string
  agent: string
  command: string
  result?: string
  status: ExecutionStatus
  startedAt: string
  finishedAt?: string | null
  nonExecutionCause?: NonExecutionCause | null
  retries: number

  /** DAG node tracking (flow origin only). */
  nodeId?: string
  kind?: NodeKind
  executor?: string
  flowRunId?: string
  depth?: number
}

export interface ExecutionListQuery {
  origin?: ExecutionOrigin
  rulePath?: string
  alertId?: string
  agent?: string
  triggeredBy?: string
  status?: ExecutionStatus
  startedAtFrom?: string
  startedAtTo?: string
  page?: number
  size?: number
}

/** Palette metadata for each backend executor. */
export interface ExecutorMeta {
  type: string
  label: string
  kinds: NodeKind[] // which kinds this executor can back
  paramsPlaceholder?: unknown
}

export const EXECUTOR_CATALOG: ExecutorMeta[] = [
  { type: 'shell', label: 'Shell (endpoint agent)', kinds: ['executor'] },
  { type: 'http', label: 'HTTP call', kinds: ['executor', 'enrichment'], paramsPlaceholder: { method: 'GET', url: '' } },
  { type: 'llm_enrich', label: 'LLM enrichment', kinds: ['enrichment'], paramsPlaceholder: { prompt: '' } },
  { type: 'llm_action', label: 'LLM action', kinds: ['executor'], paramsPlaceholder: { prompt: '' } },
  { type: 'notify', label: 'Send notification', kinds: ['executor'], paramsPlaceholder: { message: '', type: 'INFO' } },
  { type: 'incident', label: 'Open incident', kinds: ['executor'], paramsPlaceholder: { name: '', description: '' } },
  { type: 'conditional', label: 'Conditional (if/else branch)', kinds: ['executor'], paramsPlaceholder: { conditions: [] } },
]

export function executorMeta(type: string): ExecutorMeta | undefined {
  return EXECUTOR_CATALOG.find((e) => e.type === type)
}
