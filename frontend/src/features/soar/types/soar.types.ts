/* Mirrors backend modules/soar dto (rule.go, execution.go). */

export type SoarOperator = 'IS' | 'IS_ONE_OF' | 'IS_NOT_ONE_OF'
export const SOAR_OPERATORS: SoarOperator[] = ['IS', 'IS_ONE_OF', 'IS_NOT_ONE_OF']

/** A flow condition against an alert field. `value` is a string for IS and a
 *  string[] for IS_ONE_OF / IS_NOT_ONE_OF. */
export interface FlowCondition {
  operator: SoarOperator
  field: string
  value: unknown
}

/** An alert-response flow (file-backed YAML). Identity is `relPath`. */
export interface Flow {
  relPath: string
  name: string
  description: string
  conditions: FlowCondition[]
  commands: string[]
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
  commands: string[]
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
export interface Execution {
  id: number
  rulePath: string
  alertId: string
  command: string
  commandResult?: string
  agent: string
  executionDate: string
  executionStatus: ExecutionStatus
  nonExecutionCause?: NonExecutionCause | null
  executionRetries: number
}

export interface ExecutionListQuery {
  rulePath?: string
  alertId?: string
  agent?: string
  executionStatus?: ExecutionStatus
  executionDateGte?: string
  executionDateLte?: string
  page?: number // 0-based
  size?: number
}
