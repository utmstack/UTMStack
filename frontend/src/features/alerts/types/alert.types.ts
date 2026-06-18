/* Mirrors the v11-alert-* OpenSearch alert document + the alerts module DTOs. */

export interface FilterType {
  field: string
  operator: string
  value?: unknown
}

export interface Geolocation {
  country?: string
  countryCode?: string
  city?: string
  latitude?: number
  longitude?: number
}

/** "Side" object — adversary / target. Only the fields the UI uses. */
export interface Side {
  ip?: string
  host?: string
  user?: string
  domain?: string
  geolocation?: Geolocation
}

export interface AlertHistoryEntry {
  user?: string
  action?: string
  message?: string
  newValue?: string
  timestamp?: string
}

export interface IncidentDetail {
  incidentName?: string
  incidentId?: number | string
  creationDate?: string
  createdBy?: string
  source?: string
}

/** An alert document (_source of v11-alert-*). Subset of the ~100-field model. */
export interface Alert {
  id: string
  '@timestamp'?: string
  name?: string
  category?: string
  severity?: number // 1=Low 2=Medium 3=High
  severityLabel?: string
  status?: number // 1=Automatic review 2=Open 3=In review 5=Completed
  statusLabel?: string
  statusObservation?: string
  isIncident?: boolean
  incidentDetail?: IncidentDetail
  technique?: string
  description?: string
  solution?: string
  reference?: string[]
  dataSource?: string
  dataType?: string
  impactScore?: number
  impact?: { score?: number; label?: string }
  adversary?: Side
  target?: Side
  tags?: string[]
  notes?: string
  assignee?: string
  history?: AlertHistoryEntry[]
  events?: AlertEventItem[]
}

/**
 * One related event stored inline on the alert (the engine attaches up to ~11 —
 * a sample). Use the "view all related logs" action for the full set.
 */
export interface AlertEventItem {
  id?: string
  '@timestamp'?: string
  timestamp?: string
  deviceTime?: string
  dataType?: string
  dataSource?: string
  severity?: string
  protocol?: string
  action?: string
  raw?: string
  log?: Record<string, unknown>
  origin?: Side
  target?: Side
}

/**
 * GET /utm-alerts/related-logs — the backend reproduces the Event Processor's
 * correlation search (no 10-hit cap) and returns the matching log ids so the UI
 * can load every related log in the Log Explorer.
 */
export interface RelatedLogsResponse {
  ruleMatched: boolean
  indexPattern: string
  ids: string[]
  total: number
  truncated: boolean
  timeFrom: string
  timeTo: string
}

/** GET /utm-alert-tags item. */
export interface AlertTag {
  id: number
  tagName: string
  tagColor?: string
  systemOwner?: boolean
}

/* ─── Enum helpers ─────────────────────────────────────────────────────── */

export type SeverityKey = 'high' | 'medium' | 'low'
export type StatusKey = 'auto' | 'open' | 'in_review' | 'completed'

export const SEVERITY_BY_INT: Record<number, SeverityKey> = { 3: 'high', 2: 'medium', 1: 'low' }
export const SEVERITY_INT: Record<SeverityKey, number> = { high: 3, medium: 2, low: 1 }

export const STATUS_BY_INT: Record<number, StatusKey> = { 1: 'auto', 2: 'open', 3: 'in_review', 5: 'completed' }
export const STATUS_INT: Record<StatusKey, number> = { auto: 1, open: 2, in_review: 3, completed: 5 }

export const STATUS_TABS = ['all', 'open', 'in_review', 'completed', 'auto'] as const
export type StatusTab = (typeof STATUS_TABS)[number]

/** A user-defined filter row in the page filter bar. */
export interface CustomFilter {
  field: string
  label: string
  operator: string
  value: string
}
