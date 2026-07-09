import { dump, load } from 'js-yaml'
import type { AfterStep, Condition, RuleFormState } from '../components/rule-form'

/**
 * YAML ↔ rule-form conversion for the "Code" view of the rule editor. The YAML
 * mirrors the real on-disk rule format (Event Processor `Rule`): `dataTypes`
 * (string list), nested `impact`, `where` for the CEL — NOT the API field names.
 * That way the Code view shows the rule exactly as it is stored.
 *
 * `active` is intentionally not serialized: it isn't part of the rule file and is
 * managed by the toggle, so callers preserve the form's current ruleActive when
 * applying parsed YAML.
 */

export type RuleParseResult = { ok: true; form: RuleFormState } | { ok: false; error: string }

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}
function str(v: unknown): string {
  return v == null ? '' : String(v)
}
function num(v: unknown): number {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
function strArray(v: unknown): string[] {
  return Array.isArray(v) ? v.map((x) => String(x)) : []
}

function condsToYaml(cs: Condition[]) {
  return cs.filter((c) => c.field.trim()).map((c) => ({ field: c.field, operator: c.operator, value: c.value }))
}

function parseConditions(v: unknown): Condition[] {
  if (!Array.isArray(v)) return []
  return v.map((c) => {
    const o = isRecord(c) ? c : {}
    return { field: str(o.field), operator: str(o.operator) || 'filter_term', value: str(o.value) }
  })
}

function parseSteps(v: unknown): AfterStep[] {
  if (!Array.isArray(v)) return []
  return v.map((s) => {
    const o = isRecord(s) ? s : {}
    return {
      indexPattern: str(o.indexPattern),
      within: str(o.within),
      count: num(o.count),
      with: parseConditions(o.with),
      or: Array.isArray(o.or)
        ? o.or.map((g) => ({ with: parseConditions(isRecord(g) ? g.with : undefined) }))
        : [],
    }
  })
}

/** Serialize the rule form to YAML matching the on-disk rule file. */
export function ruleFormToYaml(f: RuleFormState): string {
  // Build in the on-disk field order; omit empties like the Go struct's omitempty.
  const doc: Record<string, unknown> = {
    dataTypes: f.dataTypes,
    name: f.name,
    impact: {
      confidentiality: f.confidentiality,
      integrity: f.integrity,
      availability: f.availability,
    },
    category: f.category,
    technique: f.technique,
    adversary: f.adversary,
  }
  if (f.references.length) doc.references = f.references
  if (f.description) doc.description = f.description
  doc.where = f.definition

  const steps = f.correlation
    .filter((s) => s.indexPattern.trim())
    .map((s) => {
      const step: Record<string, unknown> = {
        indexPattern: s.indexPattern,
        within: s.within,
        count: s.count,
        with: condsToYaml(s.with),
      }
      const ors = s.or.filter((g) => g.with.some((c) => c.field.trim()))
      if (ors.length) step.or = ors.map((g) => ({ with: condsToYaml(g.with) }))
      return step
    })
  if (steps.length) doc.correlation = steps
  if (f.groupBy.length) doc.groupBy = f.groupBy
  if (f.deduplicateBy.length) doc.deduplicateBy = f.deduplicateBy

  return dump(doc, { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"' })
}

/**
 * Parse a YAML rule document back into the form. Accepts the on-disk shape
 * (`where`, `impact`) and tolerates a single-item list wrapper (real rule files
 * are a YAML list) plus the older API-style names (`definition`, `severity`).
 * ruleActive is left at its default; callers preserve the current value.
 */
export function yamlToRuleForm(content: string): RuleParseResult {
  let doc: unknown
  try {
    doc = load(content)
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : 'invalid YAML' }
  }
  if (Array.isArray(doc)) doc = doc[0] // real rule files wrap the rule in a list
  if (!isRecord(doc)) return { ok: false, error: 'root must be a mapping of rule fields' }

  const impact = isRecord(doc.impact) ? doc.impact : isRecord(doc.severity) ? doc.severity : {}
  const where = doc.where != null ? doc.where : doc.definition

  const form: RuleFormState = {
    name: str(doc.name),
    adversary: str(doc.adversary) || 'origin',
    confidentiality: num(impact.confidentiality),
    integrity: num(impact.integrity),
    availability: num(impact.availability),
    category: str(doc.category),
    technique: str(doc.technique),
    description: str(doc.description),
    ruleActive: doc.active !== false,
    dataTypes: strArray(doc.dataTypes),
    definition: str(where),
    correlation: parseSteps(doc.correlation ?? doc.afterEvents),
    groupBy: strArray(doc.groupBy),
    deduplicateBy: strArray(doc.deduplicateBy),
    references: strArray(doc.references),
  }
  return { ok: true, form }
}
