import { dump, load } from 'js-yaml'
import {
  SOAR_MULTI_VALUE_OPERATORS,
  SOAR_NO_VALUE_OPERATORS,
  SOAR_OPERATORS,
  NODE_KINDS,
  type Flow,
  type FlowCondition,
  type FlowNode,
  type NodeKind,
  type SaveFlowInput,
  type SoarOperator,
} from '../types/soar.types'

/** Structured editing state for a flow. `active` is managed by the toggle
 *  outside the YAML file, so callers preserve it across the code round-trip. */
export interface FlowFormState {
  name: string
  description: string
  conditions: FlowCondition[]
  roots: string[]
  nodes: Record<string, FlowNode>
  maxDepth: number
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

function parseNode(raw: unknown): FlowNode {
  const o = isRecord(raw) ? raw : {}
  const kind = (NODE_KINDS.includes(str(o.kind) as NodeKind) ? o.kind : 'executor') as NodeKind
  const excluded = strArray(o.excludedAgents)
  return {
    kind,
    executor: str(o.executor),
    command: o.command ? str(o.command) : undefined,
    shell: o.shell ? str(o.shell) : undefined,
    platform: o.platform ? str(o.platform) : undefined,
    agent: o.agent ? str(o.agent) : undefined,
    excludedAgents: excluded.length ? excluded : undefined,
    params: o.params,
    onSuccess: strArray(o.onSuccess),
    onError: strArray(o.onError),
  }
}

function parseNodes(raw: unknown): Record<string, FlowNode> {
  if (!isRecord(raw)) return {}
  const out: Record<string, FlowNode> = {}
  for (const [id, v] of Object.entries(raw)) {
    out[id] = parseNode(v)
  }
  return out
}

/** Legacy compat: if the file still uses `commands:` chain shape, upgrade it
 *  to a linear DAG the same way the backend does — one shell node per step.
 *  Flow-level platform/agent get stamped onto every shell node so behavior
 *  survives the upgrade. */
function upgradeLegacyChain(raw: Record<string, unknown>): { roots: string[]; nodes: Record<string, FlowNode> } {
  const commands = Array.isArray(raw.commands) ? raw.commands : []
  const shell = str(raw.shell)
  const platform = str(raw.agentPlatform)
  const agent = str(raw.defaultAgent)
  if (!commands.length) return { roots: [], nodes: {} }
  const nodes: Record<string, FlowNode> = {}
  const roots: string[] = ['step_0']
  commands.forEach((c, i) => {
    const id = `step_${i}`
    const command = typeof c === 'string' ? c : str((c as Record<string, unknown>)?.command)
    const cond = typeof c === 'string' ? '' : str((c as Record<string, unknown>)?.condition)
    nodes[id] = {
      kind: 'executor',
      executor: 'shell',
      command,
      shell: shell || undefined,
      platform: platform || undefined,
      agent: agent || undefined,
      onSuccess: [],
      onError: [],
    }
    if (i === 0) return
    const prevId = `step_${i - 1}`
    const prev = nodes[prevId]
    if (cond === 'OnSuccess') prev.onSuccess = [...(prev.onSuccess ?? []), id]
    else if (cond === 'OnFailure') prev.onError = [...(prev.onError ?? []), id]
    else {
      prev.onSuccess = [...(prev.onSuccess ?? []), id]
      prev.onError = [...(prev.onError ?? []), id]
    }
  })
  return { roots, nodes }
}

export function emptyForm(): FlowFormState {
  return {
    name: '',
    description: '',
    conditions: [{ operator: 'IS', field: '', value: '' }],
    roots: [],
    nodes: {},
    maxDepth: 50,
    active: true,
  }
}

export function flowToForm(f?: Flow): FlowFormState {
  if (!f) return emptyForm()
  return {
    name: f.name,
    description: f.description ?? '',
    conditions: f.conditions?.length ? f.conditions.map((c) => ({ ...c })) : [{ operator: 'IS', field: '', value: '' }],
    roots: f.roots ? [...f.roots] : [],
    nodes: f.nodes ? Object.fromEntries(Object.entries(f.nodes).map(([id, n]) => [id, { ...n }])) : {},
    maxDepth: f.maxDepth ?? 50,
    active: f.active ?? true,
  }
}

export function formToInput(f: FlowFormState): SaveFlowInput {
  return {
    name: f.name.trim(),
    description: f.description.trim(),
    conditions: f.conditions.filter((c) => c.field.trim()).map(normalizeCond),
    roots: [...f.roots],
    nodes: f.nodes,
    maxDepth: f.maxDepth || undefined,
    active: f.active,
  }
}

/** Serialize to YAML matching the on-disk flow file (a single-item list). */
export function flowFormToYaml(f: FlowFormState): string {
  const input = formToInput(f)
  const doc: Record<string, unknown> = { name: input.name }
  if (input.description) doc.description = input.description
  doc.conditions = input.conditions
  if (input.maxDepth) doc.maxDepth = input.maxDepth
  doc.roots = input.roots
  doc.nodes = input.nodes
  return dump([doc], { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"' })
}

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
  let roots = strArray(raw.roots)
  let nodes = parseNodes(raw.nodes)
  if (!roots.length && !Object.keys(nodes).length) {
    // Legacy chain YAML — upgrade in place so old flows keep opening.
    const legacy = upgradeLegacyChain(raw)
    roots = legacy.roots
    nodes = legacy.nodes
  }
  const form: FlowFormState = {
    name: str(raw.name),
    description: str(raw.description),
    conditions: conditions.length ? conditions : [{ operator: 'IS', field: '', value: '' }],
    roots,
    nodes,
    maxDepth: typeof raw.maxDepth === 'number' ? raw.maxDepth : 50,
    active: true,
  }
  return { ok: true, form }
}
