import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { ADUser, ADUserListQuery, ADUserStats } from '../types/ad-user.types'

const api = createApiClient()

export { ApiError as ADAuditHttpError }

function listQuery(q: ADUserListQuery): string {
  const p = new URLSearchParams()
  if (q.search) p.set('search', q.search)
  if (q.tenantId) p.set('tenantId', q.tenantId)
  if (q.source) p.set('source', q.source)
  if (q.status) p.set('status', q.status)
  if (q.sort) p.set('sort', q.sort)
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 25))
  return `?${p.toString()}`
}

export const adAuditHttpService = {
  // Returns { data: ADUser[], total } — total comes from the X-Total-Count header.
  list: (q: ADUserListQuery = {}) => api.getPaged<ADUser[]>(`/ad-audit/users${listQuery(q)}`),
  stats: (tenantId?: string) =>
    api.get<ADUserStats>(`/ad-audit/stats${tenantId ? `?tenantId=${encodeURIComponent(tenantId)}` : ''}`),
}
