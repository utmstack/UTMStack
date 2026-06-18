export type Edition = 'community' | 'enterprise'

/** GET /billing/license — the evaluated license (mirrors backend domain.License). */
export interface License {
  edition: Edition
  mssp: boolean
  /** Datasource cap; 0 = unlimited. */
  datasources: number
  /** 'online' | 'offline' | '' (empty when community). */
  type: string
  /** ISO timestamp. Absent/zero when community or perpetual. */
  expiresAt?: string
}

/** GET /billing/version — deployment + instance info from version.json / instance-config.yml. */
export interface VersionInfo {
  version: string
  edition: string
  changelog?: string
  server?: string
  instanceId?: string
}

/** Datasource cap for the community edition (backend: entitlements plugin communityCap). */
export const COMMUNITY_DATASOURCE_CAP = 25

export type LicenseStatus = 'community' | 'active' | 'expiring' | 'expired'

const EXPIRY_WARN_DAYS = 30

/**
 * The backend serializes a Go zero time as year 1 for community/perpetual licenses
 * (time.Time + omitempty doesn't drop the field). Treat anything before the epoch
 * as "no expiry".
 */
export function hasExpiry(license: License): boolean {
  if (!license.expiresAt) return false
  const d = new Date(license.expiresAt)
  return !Number.isNaN(d.getTime()) && d.getUTCFullYear() > 1970
}

export function daysUntilExpiry(license: License): number | null {
  if (!hasExpiry(license)) return null
  const ms = new Date(license.expiresAt as string).getTime() - Date.now()
  return Math.ceil(ms / 86_400_000)
}

export function licenseStatus(license: License): LicenseStatus {
  if (license.edition !== 'enterprise') return 'community'
  const days = daysUntilExpiry(license)
  if (days == null) return 'active' // perpetual enterprise
  if (days < 0) return 'expired'
  if (days <= EXPIRY_WARN_DAYS) return 'expiring'
  return 'active'
}
