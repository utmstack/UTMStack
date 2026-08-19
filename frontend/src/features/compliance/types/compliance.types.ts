/* Mirrors backend modules/compliance (domain + dto). */

/* ── Catalogue ───────────────────────────────────────────────────────────── */

export interface Requirement {
  id: string
  name: string
  satisfiedBy?: string[] // control IDs
}

export interface FrameworkSection {
  key: string
  name: string
  requirements?: Requirement[]
}

/**
 * A compliance framework (file-backed YAML). Identity = key.
 *
 * There is no `enabled`: a framework is in play for a tenant because it has a
 * report or a schedule. There is no `system` either — provenance is the store's
 * to know, and it reaches the UI only as `locked`.
 */
export interface Framework {
  key: string
  name: string
  description?: string
  source?: string
  sections?: FrameworkSection[]
  /**
   * Shipped with the product, and read-only. The tenant layer is additive: you
   * build your own beside the catalogue rather than forking it. Computed per
   * response from where the file lives, never part of the definition.
   */
  system?: boolean
  /** Withheld by the licence edition. Also computed per response. */
  locked?: boolean
}

export type CheckRule = 'MIN_HITS_REQUIRED' | 'THRESHOLD_MAX'

export type Dataset = 'logs' | 'alerts'

/** Operators a check may use; a subset of the backend FilterType vocabulary. */
export type CheckFilterOperator =
  | 'IS'
  | 'IS_NOT'
  | 'IS_ONE_OF_TERMS'
  | 'IS_NOT_ONE_OF'
  | 'EXIST'
  | 'DOES_NOT_EXIST'
  | 'CONTAIN'
  | 'DOES_NOT_CONTAIN'
  | 'START_WITH'
  | 'NOT_START_WITH'
  | 'ENDS_WITH'
  | 'NOT_ENDS_WITH'

export interface CheckFilter {
  field: string
  operator: CheckFilterOperator
  value?: unknown
}

/**
 * One measurement: a slice of the event store, and a rule over how many records
 * it holds. Filters say which records count, the rule says how many make a
 * pass — the two used to be fused in a SQL string.
 *
 * `dataset` + `dataType` double as the applicability declaration: a check
 * against a data type the tenant does not receive cannot fail, only go
 * unevaluated.
 */
export interface Check {
  key: string
  name: string
  dataset?: Dataset // default 'logs'
  dataType?: string // empty = every type in the dataset
  filters?: CheckFilter[]
  rule?: CheckRule
  ruleValue?: number
  todo?: boolean
}

/** A reusable control (file-backed YAML, shared library). Identity = id. */
export interface Control {
  id: string
  family?: string
  familyName?: string
  name: string
  scope?: 'data' | 'governance'
  statement?: string
  remediation?: string
  strategy?: 'ALL' | 'ANY'
  checks?: Check[]
  source?: string
  /** Shipped with the product, and read-only. */
  system?: boolean
  locked?: boolean
}

/* ── Report ──────────────────────────────────────────────────────────────── */

export type ComplianceStatus =
  | 'COMPLIANT'
  | 'NON_COMPLIANT'
  | 'AT_RISK'
  | 'NOT_COVERED'
  | 'NOT_EVALUATED'
  | 'PENDING'
  | 'OUT_OF_SCOPE'

export const COMPLIANCE_STATUSES: ComplianceStatus[] = [
  'COMPLIANT',
  'NON_COMPLIANT',
  'AT_RISK',
  'NOT_COVERED',
  'NOT_EVALUATED',
  'PENDING',
  'OUT_OF_SCOPE',
]

/**
 * Statuses excluded from the score's denominator, all for the same reason:
 * nobody judged them. Counting them as failure would report a gap that was
 * never measured, and make a tenant look worse for every log source they have
 * not connected.
 */
const UNSCORED: ReadonlySet<ComplianceStatus> = new Set<ComplianceStatus>([
  'OUT_OF_SCOPE',
  'PENDING',
  'NOT_EVALUATED',
])

export const isScored = (s: ComplianceStatus): boolean => !UNSCORED.has(s)

export type CheckOutcome = 'PASSED' | 'FAILED' | 'NOT_APPLICABLE' | 'ERROR'

export interface CheckResult {
  key: string
  name: string
  dataset?: Dataset
  /** Why a NOT_APPLICABLE check could not run. */
  dataType?: string
  rule?: CheckRule
  /** The rule's threshold, so "0 of 1 required" reads on its own. */
  required?: number
  outcome: CheckOutcome
  hits: number
  error?: string
}

/**
 * The only editable level: the lowest one at which a judgement exists. "Did
 * this query return enough rows" is a fact; "is this control satisfied" is
 * something a person can answer.
 */
export interface ControlRow {
  controlId: string
  name: string
  /** What the report states — the human's verdict if one was given. */
  status: ComplianceStatus
  /** What the engine says now, kept so an edit never hides the measurement. */
  engineStatus: ComplianceStatus
  evidence: string
  coverage: number // enabled correlation rules covering this control
  activity: number // alerts from those rules in the window
  checks?: CheckResult[]
  /** What the engine said when the edit was made. */
  originalStatus?: ComplianceStatus
  note?: string
  editedBy?: string
  editedAt?: string
}

/** Someone wrote on this row — a note, with or without a verdict. */
export const isAnnotated = (row: ControlRow): boolean => Boolean(row.editedAt)

/**
 * Someone overrode the engine here. A plain annotation never sets
 * originalStatus, so it is what tells the two apart.
 */
export const isOverridden = (row: ControlRow): boolean => Boolean(row.originalStatus)

/**
 * An override written against a verdict the engine has since changed its mind
 * about. It still stands — only a person should withdraw what a person
 * accepted — but the report should say so.
 */
export const isEditStale = (row: ControlRow): boolean =>
  isOverridden(row) && row.originalStatus !== row.engineStatus

export interface ReportSummary {
  /** Compliant over Total — un-evaluated requirements weigh against the score. */
  compliantPct: number
  total: number
  evaluated: number
  compliant: number
  nonCompliant: number
  atRisk: number
  notCovered: number
  notEvaluated: number
  pending: number
  outOfScope: number
}

/** Derived from the controls that satisfy it; never edited directly. */
export interface ReportRequirement {
  id: string
  name: string
  status: ComplianceStatus
  /** Indexes into Report.controls. */
  controlIds: string[]
}

export interface ReportSection {
  key?: string
  name: string
  summary: ReportSummary
  requirements: ReportRequirement[]
}

/**
 * The standing report: one per framework per tenant, replaced on every run.
 *
 * Controls are flat rather than nested under requirements because the
 * crosswalk aims many requirements at the same control — nesting would copy a
 * row per reference and need editing once per copy.
 */
export interface Report {
  id: string
  frameworkKey: string
  frameworkName: string
  frameworkSource?: string
  generatedAt: string
  /** The period the activity figures were counted over. */
  windowFrom: string
  windowTo: string
  summary: ReportSummary
  sections: ReportSection[]
  controls: ControlRow[]
}

export interface ReportMeta {
  id: string
  frameworkKey: string
  frameworkName: string
  generatedAt: string
  score: number
}

/** One point on the score-over-time chart. */
export interface ScorePoint {
  day: string
  generatedAt: string
  score: number
  /** Total and evaluated ride along: the percentage alone cannot tell a fix
   *  from a log source going quiet — both move it the same way. */
  total: number
  evaluated: number
  compliant: number
  /** False once the stored document has aged past retention. */
  hasDocument: boolean
}

/**
 * Annotates a control, and optionally overrides the engine on it.
 *
 * Status is optional because annotating and overriding are different acts: a
 * remediation note claims nothing about compliance, and needing a status to
 * record one is how a row ends up with a verdict nobody meant to give.
 */
export interface EditControlInput {
  status?: ComplianceStatus
  /** Always required: an entry no one can explain is worth nothing later. */
  note: string
}

/* ── Schedules ───────────────────────────────────────────────────────────── */

export interface ComplianceSchedule {
  id: string
  userId: string
  frameworkKey: string
  scheduleString: string // 5-field cron
  /** How much the report covers, which is not how often it runs. */
  windowDays: number
  to: string // comma-separated
  cc: string // comma-separated
  lastExecutionDate: string
}

export interface SaveScheduleInput {
  id?: string // present on update
  frameworkKey: string
  scheduleString: string
  windowDays?: number
  to: string
  cc?: string
}
