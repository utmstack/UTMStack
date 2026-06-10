/* Mirrors backend appconfig/dto branding shapes. */

export type BrandingAssetSlot = 'logo' | 'logoDark' | 'favicon' | 'reportLogo' | 'reportCover'

/** GET/PUT /branding (admin). */
export interface Branding {
  enabled: boolean
  productName: string
  accentColor: string
  logoUrl: string
  logoDarkUrl: string
  faviconUrl: string
  reportLogoUrl: string
  reportCoverUrl: string
  updatedAt?: string
  updatedBy?: string
}

/** GET /branding/public — unauthenticated, applied app-wide + on pre-login. */
export interface BrandingPublic {
  enabled: boolean
  productName: string
  accentColor: string
  logoUrl: string
  logoDarkUrl: string
  faviconUrl: string
  language: string
}

export interface BrandingRequest {
  enabled: boolean
  productName: string
  accentColor: string
  logoUrl: string
  logoDarkUrl: string
  faviconUrl: string
  reportLogoUrl: string
  reportCoverUrl: string
}
