import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { SupportAccess } from '@/features/tenants/types/tenant.types'

const api = createApiClient()

export { ApiError as SupportAccessHttpError }

interface SupportAccessPayload {
  supportAccess: SupportAccess
}

/**
 * The tenant's own view of the grant. Both calls are refused for anyone but an
 * administrator of that same tenant — the platform operator cannot raise its
 * own access, which is the whole point of keeping this outside their reach.
 */
export const supportAccessHttpService = {
  get: (tenantId: string) =>
    api.get<SupportAccessPayload>(`/tenants/${tenantId}/support-access`),
  set: (tenantId: string, supportAccess: SupportAccess) =>
    api.put<unknown>(`/tenants/${tenantId}/support-access`, { supportAccess }),
}
