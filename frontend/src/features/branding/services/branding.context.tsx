import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { brandingHttpService } from './branding-http.service'
import { getSupportTenantId, useSupportTenant } from '@/shared/lib/current-tenant'
import type { BrandingPublic } from '../types/branding.types'

interface BrandingContextValue {
  branding: BrandingPublic | null
  /** Re-fetch the public branding and re-apply it (call after an admin save). */
  refresh: () => Promise<void>
}

const BrandingContext = createContext<BrandingContextValue | null>(null)

const DEFAULT_PRODUCT_NAME = 'UTMStack'

/** Apply branding to the live document: accent color (--primary), title, favicon. */
function applyBranding(b: BrandingPublic | null) {
  const root = document.documentElement
  const active = !!b?.enabled

  // Accent → --primary (inline style wins over the stylesheet's light/dark vars,
  // so a single brand color applies across both modes). Cleared when inactive.
  if (active && b?.accentColor) {
    root.style.setProperty('--primary', b.accentColor)
  } else {
    root.style.removeProperty('--primary')
  }

  // Document title.
  document.title = (active && b?.productName) || DEFAULT_PRODUCT_NAME

  // Favicon.
  if (active && b?.faviconUrl) {
    let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
    if (!link) {
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.appendChild(link)
    }
    link.href = b.faviconUrl
  }
}

export function BrandingProvider({ children }: { children: ReactNode }) {
  const [branding, setBranding] = useState<BrandingPublic | null>(null)
  const supportTenantId = useSupportTenant()?.id ?? null

  const refresh = useCallback(async () => {
    try {
      // Inside a support session the chrome has to be the tenant's, and only the
      // authenticated route can tell us whose it is: /branding/public carries no
      // credentials, so the support header is never looked at and it can only
      // ever answer for the host it was reached on — the platform's. Everyone
      // else keeps the public route, which works before anyone has signed in.
      const raw = getSupportTenantId()
        ? await brandingHttpService.get()
        : await brandingHttpService.public()
      // The admin shape has no language; it is only used pre-login, which a
      // support session is not.
      const b: BrandingPublic = { ...raw, language: 'language' in raw ? raw.language : '' }
      setBranding(b)
      applyBranding(b)
    } catch {
      // Public branding is best-effort; fall back to the default UTMStack look.
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh, supportTenantId])

  const value = useMemo(() => ({ branding, refresh }), [branding, refresh])
  return <BrandingContext.Provider value={value}>{children}</BrandingContext.Provider>
}

export function useBranding(): BrandingContextValue {
  const ctx = useContext(BrandingContext)
  if (!ctx) throw new Error('useBranding must be used within BrandingProvider')
  return ctx
}
