import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { ADUserWithIdentity, ADUserListQuery, ADUserStats, Identity, IdentityDetail, IdentityListItem } from '../types/ad-user.types'

const api = createApiClient()

export { ApiError as ADAuditHttpError }

function listQuery(q: ADUserListQuery): string {
  const p = new URLSearchParams()
  if (q.search) p.set('search', q.search)
  if (q.tenantId) p.set('tenantId', q.tenantId)
  if (q.source) p.set('source', q.source)
  if (q.status) p.set('status', q.status)
  if (q.sort) p.set('sort', q.sort)
  if (q.identity) p.set('identity', q.identity)
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 25))
  return `?${p.toString()}`
}

export const adAuditHttpService = {
  // Returns { data: ADUserWithIdentity[], total } — total comes from the X-Total-Count header.
  list: (q: ADUserListQuery = {}) => api.getPaged<ADUserWithIdentity[]>(`/ad-audit/users${listQuery(q)}`),
  stats: (tenantId?: string) =>
    api.get<ADUserStats>(`/ad-audit/stats${tenantId ? `?tenantId=${encodeURIComponent(tenantId)}` : ''}`),

  // Identity aliases.
  identities: (search?: string) =>
    api.get<IdentityListItem[]>(`/ad-audit/identities${search ? `?search=${encodeURIComponent(search)}` : ''}`),
  identityDetail: (id: string) => api.get<IdentityDetail>(`/ad-audit/identities/${id}`),
  createIdentity: (name: string) => api.post<Identity>(`/ad-audit/identities`, { name }),
  deleteIdentity: (id: string) => api.delete<{ deleted: number }>(`/ad-audit/identities/${id}`),
  // Empty identityId unassigns the user (backend semantics).
  assignIdentity: (userId: string, identityId: string) =>
    api.put<{ assigned: boolean }>(`/ad-audit/users/${userId}/identity`, { identityId }),
}
