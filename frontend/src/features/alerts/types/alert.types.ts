/* Mirrors a row of utmstack.alerts plus the alerts module DTOs. */

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
  // Stored as the label itself — there is no parallel numeric code and no
  // second field holding the text.
  severity?: string // low | medium | high
  status?: string // Automatic review | Open | In review | Completed | Merged
  statusObservation?: string
  isIncident?: boolean
  incidentDetail?: IncidentDetail
  technique?: string
  description?: string
  solution?: string
  references?: string[]
  dataSource?: string
  dataType?: string
  impactScore?: number
  // The CIA triad the rule declared, nought to three. It is not a score: the
  // single number the UI shows is impactScore.
  impact?: { confidentiality?: number; integrity?: number; availability?: number }
  adversary?: Side
  target?: Side
  tags?: string[]
  notes?: string
  assignee?: string
  history?: AlertHistoryEntry[]
  events?: AlertEventItem[]
  echoes?: number
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
  id: string
  tagName: string
  tagColor?: string
  systemOwner?: boolean
}

/* ─── Enum helpers ─────────────────────────────────────────────────────── */

export type SeverityKey = 'high' | 'medium' | 'low'
export type StatusKey = 'auto' | 'open' | 'in_review' | 'completed'

/* Two vocabularies, and mixing them is the bug this shape exists to prevent.
 *
 * SEVERITY_VALUE / STATUS_VALUE are what the store holds, so they are what a
 * filter matches on and what a bucket of an aggregation is keyed by.
 */

export const SEVERITY_VALUE: Record<SeverityKey, string> = { high: 'high', medium: 'medium', low: 'low' }
export const SEVERITY_BY_VALUE: Record<string, SeverityKey> = { high: 'high', medium: 'medium', low: 'low' }

/** The incidents module ranks severity numerically; this is the boundary. */
export const SEVERITY_RANK: Record<string, number> = { low: 1, medium: 2, high: 3 }

export const STATUS_VALUE: Record<StatusKey, string> = {
  auto: 'Automatic review',
  open: 'Open',
  in_review: 'In review',
  completed: 'Completed',
}
export const STATUS_BY_VALUE: Record<string, StatusKey> = {
  'Automatic review': 'auto',
  Open: 'open',
  'In review': 'in_review',
  Completed: 'completed',
}



// "Automatic review" is absent on purpose: nothing writes it. The engine opens
// every alert it raises and the tag-rule pass completes the ones it judges
// false positives, so the tab was a filter that could only ever return nothing.
// The status itself stays in the model — an older alert may still carry it.
export const STATUS_TABS = ['all', 'open', 'in_review', 'completed'] as const
export type StatusTab = (typeof STATUS_TABS)[number]

/** A user-defined filter row in the page filter bar. */
export type { CustomFilter } from '@/shared/components/filters/custom-filter.types'
