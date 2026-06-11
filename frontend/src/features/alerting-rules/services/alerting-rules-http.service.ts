import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as AlertingRulesHttpError }

/* ─── Types (mirror backend modules/eventprocessing/dto correlation_rule) ── */

export interface DataTypeRef {
  id?: number | null
  dataType: string
  dataTypeName?: string
  dataTypeDescription?: string
  included: boolean
}

/** GET /integrations/data-types item. */
export interface DataTypeOption {
  dataType: string
  name: string
  moduleName: string
  active: boolean
  isSystem: boolean
}

/** One uploaded rule YAML file for the bulk import endpoint. */
export interface ImportRuleFile {
  filename: string
  content: string
}

/** Per-file verdict from the import endpoint. */
export interface ImportRuleResult {
  filename: string
  approved: boolean
  name?: string
  relPath?: string
  error?: string
}

export interface ImportRulesResponse {
  results: ImportRuleResult[]
  approved: number
  rejected: number
}

/**
 * A correlation (alerting) rule. Identity is `relPath` (YAML-direct). `definition`
 * is the CEL `where` condition (a JSON-encoded string); `afterEvents`/`references`/
 * `groupBy`/`deduplicateBy` are arbitrary JSON the backend stores verbatim.
 */
export interface CorrelationRule {
  relPath: string
  name: string
  adversary: string // "origin" | "target"
  confidentiality: number // 0-3
  integrity: number // 0-3
  availability: number // 0-3
  category: string
  technique: string
  description: string
  references: unknown
  definition: string
  afterEvents: unknown
  groupBy: unknown
  deduplicateBy: unknown
  ruleLastUpdate?: string
  ruleActive: boolean
  systemOwner: boolean
  dataTypes: DataTypeRef[]
}

export interface SaveRuleInput {
  relPath?: string // present → update; absent → create
  name: string
  adversary: string
  confidentiality: number
  integrity: number
  availability: number
  category: string
  technique: string
  description: string
  references: unknown
  definition: string
  afterEvents: unknown
  groupBy: unknown
  deduplicateBy: unknown
  ruleActive: boolean
  dataTypes: DataTypeRef[]
}

export interface RuleListQuery {
  ruleName?: string
  ruleActive?: boolean
  ruleCategory?: string[]
  ruleAdversary?: string[]
  ruleTechnique?: string[]
  systemOwner?: boolean
  dataTypes?: string[]
  initDate?: string
  endDate?: string
  page?: number // 0-based (Spring Pageable)
  size?: number
  sort?: string
}

function listQuery(q: RuleListQuery): string {
  const p = new URLSearchParams()
  if (q.ruleName) p.set('ruleName', q.ruleName)
  if (q.ruleActive != null) p.set('ruleActive', String(q.ruleActive))
  if (q.systemOwner != null) p.set('systemOwner', String(q.systemOwner))
  for (const c of q.ruleCategory ?? []) p.append('ruleCategory', c)
  for (const a of q.ruleAdversary ?? []) p.append('ruleAdversary', a)
  for (const tq of q.ruleTechnique ?? []) p.append('ruleTechnique', tq)
  for (const d of q.dataTypes ?? []) p.append('dataTypes', d)
  if (q.initDate) p.set('initDate', q.initDate)
  if (q.endDate) p.set('endDate', q.endDate)
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 20))
  if (q.sort) p.set('sort', q.sort)
  return p.toString()
}

// The correlation-rule routes live under the eventprocessing module group.
const BASE = '/eventprocessing/correlation-rule'

export const alertingRulesHttpService = {
  // Returns { data: CorrelationRule[], total } from X-Total-Count.
  list: (q: RuleListQuery = {}) => api.getPaged<CorrelationRule[]>(`${BASE}/search-by-filters?${listQuery(q)}`),
  get: (relPath: string) => api.get<CorrelationRule>(`${BASE}/find?relPath=${encodeURIComponent(relPath)}`),
  create: (input: SaveRuleInput) => api.post<void>(BASE, input),
  importRules: (files: ImportRuleFile[]) =>
    api.post<ImportRulesResponse>(`${BASE}/import`, { files }),
  update: (input: SaveRuleInput) => api.put<void>(BASE, input),
  setActive: (relPath: string, active: boolean) =>
    api.put<void>(`${BASE}/activate-deactivate?relPath=${encodeURIComponent(relPath)}&active=${active}`),
  remove: (relPath: string) => api.delete<void>(`${BASE}?relPath=${encodeURIComponent(relPath)}`),
  // prop ∈ rule_category | rule_technique | rule_adversary | rule_name
  propertyValues: (prop: string, value = '') =>
    api.get<string[]>(`${BASE}/search-property-values?prop=${prop}${value ? `&value=${encodeURIComponent(value)}` : ''}`),
  // Available data types (from the integrations/modules catalog) for the picker + filter.
  dataTypes: () => api.get<DataTypeOption[]>('/integrations/data-types'),
}
