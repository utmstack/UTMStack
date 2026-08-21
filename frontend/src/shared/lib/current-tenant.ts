import { useSyncExternalStore } from 'react'

/**
 * The tenant a platform operator is currently working inside, if any.
 *
 * Same shape as the federation instance selection and for the same reason: the
 * api-client request interceptor runs outside React and has to read it to stamp
 * `X-Tenant-Id`. A hook mirrors it for components.
 *
 * Only ever set for the instance operator, and only for a tenant that granted
 * support access — the backend answers 403 to the header otherwise, so this is
 * a convenience, never a permission.
 */
const STORAGE_KEY = 'utmstack_support_tenant'

export interface SupportTenant {
  id: string
  name: string
  /** The level the tenant granted when we entered, for the banner to name it. */
  access: 'READ' | 'FULL'
  /** The tenant's routable domain. Used to build install/ingest URLs in MSSP mode. */
  domain?: string
}

function read(): SupportTenant | null {
  if (typeof localStorage === 'undefined') return null
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as SupportTenant
    return parsed?.id ? parsed : null
  } catch {
    return null
  }
}

let current: SupportTenant | null = read()
const listeners = new Set<() => void>()

export function getSupportTenant(): SupportTenant | null {
  return current
}

export function getSupportTenantId(): string | null {
  return current?.id ?? null
}

export function setSupportTenant(tenant: SupportTenant | null): void {
  current = tenant
  if (typeof localStorage !== 'undefined') {
    if (tenant == null) localStorage.removeItem(STORAGE_KEY)
    else localStorage.setItem(STORAGE_KEY, JSON.stringify(tenant))
  }
  listeners.forEach((l) => l())
}

function subscribe(fn: () => void): () => void {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

/** React binding for the tenant being supported (re-renders on enter/leave). */
export function useSupportTenant(): SupportTenant | null {
  return useSyncExternalStore(subscribe, getSupportTenant, () => null)
}

/**
 * Paths that must never carry the header.
 *
 * `/auth/*` resolves the caller's own account, which lives in the platform
 * tenant: scoped to somebody else's it simply is not found, the call 401s, and
 * the client's refresh-then-logout path throws the operator out of their own
 * session. `/tenants` is worse than useless — entering a tenant makes the actor
 * that tenant, so the platform check behind the tenant list no longer passes,
 * and the way back would 403.
 */
export function carriesSupportTenant(url: string): boolean {
  const path = url.split('?')[0]
  return !path.startsWith('/auth/') && !path.startsWith('/tenants')
}
