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
  // Stored as the label itself; there is no parallel numeric code.
  severity?: string // low | medium | high
  description?: string
  technique?: string
  status?: string
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
