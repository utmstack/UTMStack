import { dump, load } from 'js-yaml'
import type { Check, Control } from '../types/compliance.types'

export const CONTROL_SCOPES = ['data', 'governance']
export const CONTROL_STRATEGIES = ['ALL', 'ANY']
export const CHECK_RULES = ['', 'NO_HITS_ALLOWED', 'MIN_HITS_REQUIRED', 'THRESHOLD_MAX', 'MATCH_FIELD_VALUE']

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
  enabled: boolean
}

export type ControlParseResult = { ok: true; form: ControlFormState } | { ok: false; error: string }

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}
const str = (v: unknown): string => (v == null ? '' : String(v))

function parseChecks(v: unknown): Check[] {
  if (!Array.isArray(v)) return []
  return v.map((c) => {
    const o = isRecord(c) ? c : {}
    const check: Check = { key: str(o.key), name: str(o.name) }
    if (o.indexPattern != null) check.indexPattern = str(o.indexPattern)
    if (o.sql != null) check.sql = str(o.sql)
    if (o.rule != null) check.rule = str(o.rule)
    if (o.ruleValue != null && o.ruleValue !== '') check.ruleValue = Number(o.ruleValue)
    if (o.field != null) check.field = str(o.field)
    if (o.expected != null) check.expected = str(o.expected)
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
    enabled: c?.enabled ?? true,
  }
}

function cleanChecks(checks: Check[]): Check[] {
  return checks
    .filter((c) => c.key.trim() || c.name.trim())
    .map((c) => {
      const out: Check = { key: c.key.trim(), name: c.name.trim() }
      if (c.indexPattern?.trim()) out.indexPattern = c.indexPattern.trim()
      if (c.sql?.trim()) out.sql = c.sql.trim()
      if (c.rule?.trim()) out.rule = c.rule.trim()
      if (c.ruleValue != null && !Number.isNaN(c.ruleValue)) out.ruleValue = c.ruleValue
      if (c.field?.trim()) out.field = c.field.trim()
      if (c.expected?.trim()) out.expected = c.expected.trim()
      if (c.todo) out.todo = true
      return out
    })
}

/** Build the Control payload for create/update. */
export function formToControl(f: ControlFormState): Control {
  return {
    id: f.id.trim(),
    name: f.name.trim(),
    family: f.family.trim(),
    familyName: f.familyName.trim(),
    scope: f.scope,
    strategy: f.strategy,
    statement: f.statement.trim(),
    remediation: f.remediation.trim(),
    checks: cleanChecks(f.checks),
    system: false,
    enabled: f.enabled,
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
    enabled: true,
  }
  return { ok: true, form }
}
