import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  AIUsage,
  CreateTenantRequest,
  Tenant,
  TenantFilter,
  TenantStats,
  UpdateTenantRequest,
} from '../types/tenant.types'

const api = createApiClient()

export { ApiError as TenantsHttpError }

function listQuery(f: TenantFilter): string {
  const p = new URLSearchParams()
  if (f.name) p.set('name', f.name)
  if (f.domain) p.set('domain', f.domain)
  if (f.status) p.set('status', f.status)
  if (f.page != null) p.set('page', String(f.page))
  if (f.size != null) p.set('size', String(f.size))
  const s = p.toString()
  return s ? `?${s}` : ''
}

export const tenantsHttpService = {
  list: (f: TenantFilter = {}) => api.get<Tenant[]>(`/tenants${listQuery(f)}`),
  get: (id: string) => api.get<Tenant>(`/tenants/${id}`),
  create: (input: CreateTenantRequest) => api.post<Tenant>('/tenants', input),
  update: (id: string, input: UpdateTenantRequest) => api.put<Tenant>(`/tenants/${id}`, input),
  terminate: (id: string) => api.delete<void>(`/tenants/${id}`),
}

/**
 * Counts for one tenant, read through the support-access header.
 *
 * Nothing here is a tenant-specific endpoint: these are the ordinary module
 * routes, executed inside the target tenant because `X-Tenant-Id` puts the
 * request there. The backend only honours the header when that tenant granted
 * the platform READ or FULL, so a card that comes back empty is the tenant's
 * decision and not a display rule invented here. Sending it for our own tenant
 * is a no-op: the middleware returns early when the target is where we already
 * are.
 *
 * Each count is settled on its own — an OpenSearch that is down must not blank
 * the user count sitting next to it in Postgres.
 */
export async function fetchTenantStats(tenantId: string): Promise<TenantStats> {
  const headers = { 'X-Tenant-Id': tenantId }

  const [users, datasources, openAlerts, ai] = await Promise.allSettled([
    api.get<{ page_info: { total_items: number } }>('/users?page=1&page_size=1', { headers }),
    api.get<{ count: number }>('/datasources/count', { headers }),
    api.get<number>('/utm-alerts/count-open-alerts', { headers }),
    api.get<AIUsage>('/soc-ai/usage', { headers }),
  ])

  return {
    users: users.status === 'fulfilled' ? (users.value.page_info?.total_items ?? null) : null,
    datasources: datasources.status === 'fulfilled' ? (datasources.value?.count ?? null) : null,
    openAlerts:
      openAlerts.status === 'fulfilled' && typeof openAlerts.value === 'number'
        ? openAlerts.value
        : null,
    ai:
      ai.status === 'fulfilled'
        ? { used: ai.value.used ?? 0, limit: ai.value.limit ?? 0 }
        : null,
  }
}

/**
 * Whether the platform may read this tenant's data.
 *
 * This only decides what to ask for. Access itself is decided by the backend,
 * which answers 403 to the header regardless of what is drawn here.
 */
export function canReadTenant(tenant: Tenant): boolean {
  return tenant.supportAccess !== 'NONE'
}
