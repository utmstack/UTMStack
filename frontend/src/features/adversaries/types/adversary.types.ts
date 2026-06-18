/* Backend: POST /adversary/alerts → AdversaryResponse[] (alerts grouped by adversary). */

export interface Geolocation {
  country?: string
  city?: string
  countryCode?: string
  asn?: number
  aso?: string
  latitude?: number
  longitude?: number
}

export interface Side {
  ip?: string
  host?: string
  user?: string
  port?: number
  domain?: string
  mac?: string
  geolocation?: Geolocation
}

/** Subset of domain.UtmAlert we consume. */
export interface AlertRaw {
  id?: string
  '@timestamp'?: string
  name?: string
  category?: string
  severity?: number
  severityLabel?: string
  description?: string
  technique?: string
  status?: number
  statusLabel?: string
  tags?: string[]
  target?: Side
}

export interface AlertWithChildren {
  alert: AlertRaw
  children?: AlertRaw[]
}

export interface AdversaryResponse {
  adversary: Side | null
  alerts: AlertWithChildren[]
}

/* ── Derived view model (computed client-side from AdversaryResponse) ── */

export type ThreatLevel = 'critical' | 'high' | 'medium' | 'low'
export type AdversaryKind = 'ip' | 'host' | 'user'

export interface TargetAgg {
  label: string
  kind: AdversaryKind
  hits: number
}

export interface AlertLite {
  id: string
  timestamp?: string
  name: string
  category?: string
  severity: number
  severityLabel?: string
  technique?: string
  statusLabel?: string
  echoes: number // number of correlated child alerts
  target?: { label: string; kind: AdversaryKind }
}

export interface Adversary {
  id: string
  identifier: string
  kind: AdversaryKind
  geo?: Geolocation
  threatLevel: ThreatLevel
  maxSeverity: number
  alertsCount: number
  firstSeen?: string
  lastSeen?: string
  techniques: { id: string; count: number }[]
  targets: TargetAgg[]
  activity: number[] // 24 hourly buckets
  alerts: AlertLite[]
  tags: string[]
}
