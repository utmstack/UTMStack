import { createApiClient } from '@/shared/lib/api-client'
import type { AuditListQuery, AuditListResponse, AuditLog } from '../types/audit.types'

const api = createApiClient()

function buildQuery(q: AuditListQuery): string {
  const params = new URLSearchParams()
  if (q.user_id != null) params.set('user_id', String(q.user_id))
  if (q.action) params.set('action', q.action)
  if (q.status) params.set('status', q.status)
  if (q.resource_type) params.set('resource_type', q.resource_type)
  if (q.resource_id) params.set('resource_id', q.resource_id)
  if (q.from) params.set('from', q.from)
  if (q.to) params.set('to', q.to)
  if (q.page != null) params.set('page', String(q.page))
  if (q.page_size != null) params.set('page_size', String(q.page_size))
  const s = params.toString()
  return s ? `?${s}` : ''
}

export const auditHttpService = {
  list: (query: AuditListQuery = {}) =>
    api.get<AuditListResponse>(`/audit${buildQuery(query)}`),
  get: (id: number) => api.get<AuditLog>(`/audit/${id}`),
}
