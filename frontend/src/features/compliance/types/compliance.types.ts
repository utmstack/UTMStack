/* Mirrors backend modules/compliance domain (framework.go, control.go, report.go). */

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

/** A compliance framework (file-backed YAML). Identity = key. */
export interface Framework {
  key: string
  name: string
  description?: string
  source?: string
  sections?: FrameworkSection[]
  relPath?: string
  system: boolean
  enabled: boolean
  /** Gated by the license edition (community can't enable/use it). */
  locked?: boolean
}

export interface Check {
  key: string
  name: string
  indexPattern?: string
  sql?: string
  rule?: string
  ruleValue?: number
  field?: string
  expected?: string
  todo?: boolean
}

/** A reusable control (file-backed YAML, shared library). Identity = id. */
export interface Control {
  id: string
  family?: string
  familyName?: string
  name: string
  scope?: string // "data" | "governance"
  statement?: string
  remediation?: string
  strategy?: string // "ALL" | "ANY"
  checks?: Check[]
  source?: string
  relPath?: string
  system: boolean
  enabled: boolean
  /** Gated by the license edition (community can't enable/use it). */
  locked?: boolean
}

export type ControlStatus =
  | 'COMPLIANT'
  | 'NON_COMPLIANT'
  | 'AT_RISK'
  | 'NOT_COVERED'
  | 'OUT_OF_SCOPE'
  | 'PENDING'

export const CONTROL_STATUSES: ControlStatus[] = [
  'COMPLIANT',
  'NON_COMPLIANT',
  'AT_RISK',
  'NOT_COVERED',
  'OUT_OF_SCOPE',
  'PENDING',
]

export interface ReportControlRow {
  controlId: string
  name: string
  status: ControlStatus
  evidence: string
  coverage: number // enabled correlation rules covering this control
  activity: number // alerts from those rules in the window
  overridden?: boolean // true when status came from a manual override
}

export interface ReportSection {
  name: string
  controls: ReportControlRow[]
}

export interface ReportSummary {
  compliantPct: number
  total: number
  compliant: number
  nonCompliant: number
  atRisk: number
  notCovered: number
  pending: number
  outOfScope: number
}

export interface Report {
  frameworkKey: string
  frameworkName: string
  generatedAt: string
  summary: ReportSummary
  sections: ReportSection[]
}

export interface ReportSnapshotMeta {
  id: string
  frameworkKey: string
  frameworkName: string
  '@timestamp': string
  score: number
}

export interface ReportSnapshot extends ReportSnapshotMeta {
  report: Report
}

export interface ComplianceSchedule {
  id: number
  userId: number
  frameworkKey: string
  scheduleString: string // 5-field cron
  recipients: string // comma-separated emails
  lastExecutionDate: string
}

export interface SaveScheduleInput {
  id?: number // present on update
  frameworkKey: string
  scheduleString: string
  recipients: string
}
