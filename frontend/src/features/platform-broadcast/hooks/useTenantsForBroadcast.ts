import { useEffect, useState } from 'react'
import { tenantsHttpService } from '@/features/tenants/services/tenants-http.service'
import type { Tenant } from '@/features/tenants/types/tenant.types'

/**
 * The platform-plane tenant. Broadcasts to "all tenants" deliberately skip it
 * — see backend/bulk.md notes for SMTP and branding. Kept as the module-level
 * constant so callers do not each import the same UUID.
 */
export const DEFAULT_TENANT_ID = 'ce66672c-e36d-4761-a8c8-90058fee1a24'

export interface UseTenantsForBroadcastState {
  tenants: Tenant[]
  loading: boolean
  error: string | null
}

/**
 * Fetch every ACTIVE tenant, once per mount of the modal that uses it.
 *
 * A modal that only appears on click does not need caching — the network call
 * is one round trip against a small list. When it hurts, add SWR here without
 * touching callers.
 */
export function useTenantsForBroadcast(open: boolean): UseTenantsForBroadcastState {
  const [state, setState] = useState<UseTenantsForBroadcastState>({
    tenants: [],
    loading: false,
    error: null,
  })

  useEffect(() => {
    if (!open) return
    let cancelled = false
    setState({ tenants: [], loading: true, error: null })
    tenantsHttpService
      .list({ status: 'ACTIVE' })
      .then((list) => {
        if (cancelled) return
        setState({ tenants: list, loading: false, error: null })
      })
      .catch((err: unknown) => {
        if (cancelled) return
        const msg = err instanceof Error ? err.message : 'Failed to load tenants'
        setState({ tenants: [], loading: false, error: msg })
      })
    return () => {
      cancelled = true
    }
  }, [open])

  return state
}

/**
 * Tenants a "select all" would actually target. Some endpoints (SMTP, branding)
 * exclude the platform-plane tenant even when `allTenants=true`; mirror that in
 * the picker so the count the operator sees matches what the backend will do.
 */
export function filterForAllTenants(tenants: Tenant[], excludeDefault: boolean): Tenant[] {
  if (!excludeDefault) return tenants
  return tenants.filter((t) => t.id !== DEFAULT_TENANT_ID)
}
