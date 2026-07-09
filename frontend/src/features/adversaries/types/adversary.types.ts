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
