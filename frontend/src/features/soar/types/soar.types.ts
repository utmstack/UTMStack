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

/** A flow condition against an alert field. `value` is string[] for
 *  IS_ONE_OF/IS_NOT_ONE_OF, unused for EXISTS/NOT_EXISTS, and a string
 *  otherwise. */
export interface FlowCondition {
  operator: SoarOperator
  field: string
  value: unknown
}

export type SoarCondition = 'OnSuccess' | 'OnFailure' | 'Always'
export const SOAR_CONDITIONS: SoarCondition[] = ['OnSuccess', 'OnFailure', 'Always']

/** Join semantic for a command relative to the previous one:
 *  OnSuccess → `&&`, OnFailure → `||`, Always → `;`. Absent on the first
 *  command (nothing to chain against). */
export interface FlowCommand {
  command: string
  condition?: SoarCondition
}

/** An alert-response flow (file-backed YAML). Identity is `relPath`. */
export interface Flow {
  relPath: string
  name: string
  description: string
  conditions: FlowCondition[]
  commands: FlowCommand[]
  active: boolean
  agentPlatform: string
  defaultAgent: string
  shell: string
  excludedAgents: string[]
  systemOwner: boolean
  lastModifiedDate?: string | null
}

/** Create/Update payload (POST /soar/rules, PUT /soar/rules/:relPath). */
export interface SaveFlowInput {
  name: string
  description: string
  conditions: FlowCondition[]
  commands: FlowCommand[]
  active: boolean
  agentPlatform: string
  defaultAgent: string
  shell: string
  excludedAgents: string[]
}

export interface FlowListQuery {
  name?: string
  active?: boolean
  agentPlatform?: string
  systemOwner?: boolean
  page?: number // 0-based
  size?: number
}

export type ExecutionStatus = 'EXECUTED' | 'PENDING' | 'FAILED'
export type NonExecutionCause = 'AGENT_OFFLINE' | 'AGENT_NOT_FOUND' | 'UNKNOWN'

/** One recorded flow execution (Postgres-backed). */
/** What raised a command: a flow that matched an alert, or a person at the
 *  interactive console. It decides which half of the row carries anything. */
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
  page?: number // 0-based
  size?: number
}
