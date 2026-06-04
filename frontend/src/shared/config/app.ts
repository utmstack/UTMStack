// Mock build/license info. Replace with a real fetch (e.g. GET /api/v1/license-info)
// once the backend endpoint exists.
export type LicenseStatus = 'active' | 'trial' | 'expiring' | 'expired' | 'none'
export type Edition = 'community' | 'enterprise'

export interface AppInfo {
  edition: Edition
  tier: string | null
  version: string
  licenseStatus: LicenseStatus
  expiresAt: string | null
  trialDaysLeft: number | null
}

export const APP_INFO: AppInfo = {
  edition: 'community',
  tier: null,
  version: import.meta.env.VITE_APP_VERSION ?? '11.1.4',
  licenseStatus: 'active',
  expiresAt: null,
  trialDaysLeft: null,
}
