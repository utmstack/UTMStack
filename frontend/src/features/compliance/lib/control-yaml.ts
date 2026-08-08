import { dump, load } from 'js-yaml'
import type { Check, CheckFilter, CheckFilterOperator, CheckRule, Control, Dataset } from '../types/compliance.types'

export const CONTROL_SCOPES = ['data', 'governance']
export const CONTROL_STRATEGIES = ['ALL', 'ANY']
// Two rules, and they are the two directions: at least N, or at most N.
// "No hits allowed" is THRESHOLD_MAX with 0, and matching a field value is a
// filter — fusing selection into the rule is what the old vocabulary did.
export const CHECK_RULES: ('' | CheckRule)[] = ['', 'MIN_HITS_REQUIRED', 'THRESHOLD_MAX']
export const CHECK_DATASETS: Dataset[] = ['logs', 'alerts']
export const CHECK_OPERATORS: CheckFilterOperator[] = [
  'IS',
  'IS_NOT',
  'IS_ONE_OF_TERMS',
  'IS_NOT_ONE_OF',
  'EXIST',
  'DOES_NOT_EXIST',
  'CONTAIN',
  'DOES_NOT_CONTAIN',
  'START_WITH',
  'NOT_START_WITH',
  'ENDS_WITH',
  'NOT_ENDS_WITH',
]

/** Operators whose meaning is the field's presence, so they carry no value. */
export const VALUELESS_OPERATORS: ReadonlySet<CheckFilterOperator> = new Set<CheckFilterOperator>([
  'EXIST',
  'DOES_NOT_EXIST',
])

/** Operators that take a list rather than a single value. */
export const LIST_OPERATORS: ReadonlySet<CheckFilterOperator> = new Set<CheckFilterOperator>([
  'IS_ONE_OF_TERMS',
  'IS_NOT_ONE_OF',
])

export interface ControlFormState {
  id: string
  name: string
  family: string
  familyName: string
  scope: string
  strategy: string
  statement: string
  remediation: string
  checks: Check[]
}

export type ControlParseResult = { ok: true; form: ControlFormState } | { ok: false; error: string }

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}
const str = (v: unknown): string => (v == null ? '' : String(v))

function parseFilters(v: unknown): CheckFilter[] {
  if (!Array.isArray(v)) return []
  return v.map((f) => {
    const o = isRecord(f) ? f : {}
    return {
      field: str(o.field),
      operator: (str(o.operator) || 'IS') as CheckFilterOperator,
      value: o.value,
    }
  })
}

function parseChecks(v: unknown): Check[] {
  if (!Array.isArray(v)) return []
  return v.map((c) => {
    const o = isRecord(c) ? c : {}
    const check: Check = { key: str(o.key), name: str(o.name) }
    if (o.dataset != null) check.dataset = str(o.dataset) as Dataset
    if (o.dataType != null) check.dataType = str(o.dataType)
    const filters = parseFilters(o.filters)
    if (filters.length) check.filters = filters
    if (o.rule != null) check.rule = str(o.rule) as CheckRule
    if (o.ruleValue != null && o.ruleValue !== '') check.ruleValue = Number(o.ruleValue)
    if (o.todo === true) check.todo = true
    return check
  })
}

export function controlToForm(c?: Control): ControlFormState {
  return {
    id: c?.id ?? '',
    name: c?.name ?? '',
    family: c?.family ?? '',
    familyName: c?.familyName ?? '',
    scope: c?.scope ?? 'data',
    strategy: c?.strategy ?? 'ALL',
    statement: c?.statement ?? '',
    remediation: c?.remediation ?? '',
    checks: c?.checks ? c.checks.map((x) => ({ ...x })) : [],
  }
}

function cleanChecks(checks: Check[]): Check[] {
  return checks
    .filter((c) => c.key.trim() || c.name.trim())
    .map((c) => {
      const out: Check = { key: c.key.trim(), name: c.name.trim() }
      // 'logs' is the default, so writing it back would only add noise to the
      // file a person edits.
      if (c.dataset && c.dataset !== 'logs') out.dataset = c.dataset
      if (c.dataType?.trim()) out.dataType = c.dataType.trim()
      const filters = (c.filters ?? []).filter((f) => f.field.trim())
      if (filters.length) out.filters = filters.map((f) => cleanFilter(f))
      if (c.rule) out.rule = c.rule
      if (c.ruleValue != null && !Number.isNaN(c.ruleValue)) out.ruleValue = c.ruleValue
      if (c.todo) out.todo = true
      return out
    })
}

function cleanFilter(f: CheckFilter): CheckFilter {
  const out: CheckFilter = { field: f.field.trim(), operator: f.operator }
  if (!VALUELESS_OPERATORS.has(f.operator)) out.value = f.value
  return out
}

/** Build the Control payload for create/update. */
export function formToControl(f: ControlFormState): Control {
  return {
    id: f.id.trim(),
    name: f.name.trim(),
    family: f.family.trim(),
    familyName: f.familyName.trim(),
    scope: f.scope as Control['scope'],
    strategy: f.strategy as Control['strategy'],
    statement: f.statement.trim(),
    remediation: f.remediation.trim(),
    checks: cleanChecks(f.checks),
  }
}

/** Serialize to the on-disk control YAML (a single mapping; omits empties). */
export function controlFormToYaml(f: ControlFormState): string {
  const doc: Record<string, unknown> = { id: f.id, name: f.name }
  if (f.family) doc.family = f.family
  if (f.familyName) doc.familyName = f.familyName
  if (f.scope) doc.scope = f.scope
  if (f.strategy) doc.strategy = f.strategy
  if (f.statement) doc.statement = f.statement
  if (f.remediation) doc.remediation = f.remediation
  const checks = cleanChecks(f.checks)
  if (checks.length) doc.checks = checks
  return dump(doc, { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"' })
}

export function yamlToControlForm(content: string): ControlParseResult {
  let raw: unknown
  try {
    raw = load(content)
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : 'invalid YAML' }
  }
  if (Array.isArray(raw)) raw = raw[0]
  if (!isRecord(raw)) return { ok: false, error: 'root must be a control mapping' }
  const form: ControlFormState = {
    id: str(raw.id),
    name: str(raw.name),
    family: str(raw.family),
    familyName: str(raw.familyName),
    scope: str(raw.scope) || 'data',
    strategy: str(raw.strategy) || 'ALL',
    statement: str(raw.statement),
    remediation: str(raw.remediation),
    checks: parseChecks(raw.checks),
  }
  return { ok: true, form }
}
