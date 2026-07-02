import { dump, load } from 'js-yaml'
import { SOAR_MULTI_VALUE_OPERATORS, SOAR_NO_VALUE_OPERATORS, SOAR_OPERATORS, type Flow, type FlowCondition, type SaveFlowInput, type SoarOperator } from '../types/soar.types'

/** Structured editing state for a flow. `active` is managed by the toggle (not in
 *  the YAML file), so callers preserve it across the code round-trip. */
export interface FlowFormState {
  name: string
  description: string
  conditions: FlowCondition[]
  commands: string[]
  shell: string
  agentPlatform: string
  defaultAgent: string
  excludedAgents: string[]
  active: boolean
}

export type FlowParseResult = { ok: true; form: FlowFormState } | { ok: false; error: string }

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}
const str = (v: unknown): string => (v == null ? '' : String(v))
const strArray = (v: unknown): string[] => (Array.isArray(v) ? v.map(String) : [])

function parseCond(c: unknown): FlowCondition {
  const o = isRecord(c) ? c : {}
  const op = (SOAR_OPERATORS.includes(str(o.operator) as SoarOperator) ? o.operator : 'IS') as SoarOperator
  return { operator: op, field: str(o.field), value: o.value }
}

/** Coerce a condition's value to the right shape for its operator. */
function normalizeCond(c: FlowCondition): FlowCondition {
  const field = c.field.trim()
  if (SOAR_NO_VALUE_OPERATORS.includes(c.operator)) {
    return { operator: c.operator, field, value: '' }
  }
  if (SOAR_MULTI_VALUE_OPERATORS.includes(c.operator)) {
    const arr = Array.isArray(c.value)
      ? c.value.map(String)
      : str(c.value)
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
    return { operator: c.operator, field, value: arr }
  }
  const v = Array.isArray(c.value) ? c.value[0] : c.value
  return { operator: c.operator, field, value: str(v) }
}

export function flowToForm(f?: Flow): FlowFormState {
  return {
    name: f?.name ?? '',
    description: f?.description ?? '',
    conditions: f?.conditions?.length ? f.conditions.map((c) => ({ ...c })) : [{ operator: 'IS', field: '', value: '' }],
    commands: f?.commands?.length ? [...f.commands] : [''],
    shell: f?.shell ?? '',
    agentPlatform: f?.agentPlatform ?? '',
    defaultAgent: f?.defaultAgent ?? '',
    excludedAgents: f?.excludedAgents ? [...f.excludedAgents] : [],
    active: f?.active ?? true,
  }
}

export function formToInput(f: FlowFormState): SaveFlowInput {
  return {
    name: f.name.trim(),
    description: f.description.trim(),
    conditions: f.conditions.filter((c) => c.field.trim()).map(normalizeCond),
    commands: f.commands.map((c) => c.trim()).filter(Boolean),
    active: f.active,
    agentPlatform: f.agentPlatform.trim(),
    defaultAgent: f.defaultAgent.trim(),
    shell: f.shell.trim(),
    excludedAgents: f.excludedAgents.map((a) => a.trim()).filter(Boolean),
  }
}

/** Serialize to YAML matching the on-disk flow file (a single-item list). */
export function flowFormToYaml(f: FlowFormState): string {
  const input = formToInput(f)
  const doc: Record<string, unknown> = { name: input.name }
  if (input.description) doc.description = input.description
  doc.conditions = input.conditions
  doc.commands = input.commands
  if (input.shell) doc.shell = input.shell
  if (input.agentPlatform) doc.agentPlatform = input.agentPlatform
  if (input.defaultAgent) doc.defaultAgent = input.defaultAgent
  if (input.excludedAgents.length) doc.excludedAgents = input.excludedAgents
  return dump([doc], { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"' })
}

/** Parse a flow YAML back into the form. Accepts the on-disk list form and a
 *  single mapping. ruleActive/active is left default; callers preserve it. */
export function yamlToFlowForm(content: string): FlowParseResult {
  let raw: unknown
  try {
    raw = load(content)
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : 'invalid YAML' }
  }
  if (Array.isArray(raw)) raw = raw[0]
  if (!isRecord(raw)) return { ok: false, error: 'root must be a flow mapping' }

  const conditions = Array.isArray(raw.conditions) ? raw.conditions.map(parseCond) : []
  const commands = Array.isArray(raw.commands) ? raw.commands.map(String) : []
  const form: FlowFormState = {
    name: str(raw.name),
    description: str(raw.description),
    conditions: conditions.length ? conditions : [{ operator: 'IS', field: '', value: '' }],
    commands: commands.length ? commands : [''],
    shell: str(raw.shell),
    agentPlatform: str(raw.agentPlatform),
    defaultAgent: str(raw.defaultAgent),
    excludedAgents: strArray(raw.excludedAgents),
    active: true,
  }
  return { ok: true, form }
}
