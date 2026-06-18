import { createApiClient } from '@/shared/lib/api-client'
import type { AuditListQuery, AuditListResponse, AuditLog } from '../types/audit.types'

const api = createApiClient()

/**
 * Translate the UI filters into the backend search_query DSL — "field.op.value"
 * terms joined by "&" (e.g. "status.eq.failure&timestamp.gte.2026-06-01"). Only
 * allowlisted fields survive on the backend (user_id, user_login, action, status,
 * resource_type, resource_id, event_type, ip, timestamp).
 */
function buildSearchQuery(q: AuditListQuery): string {
  const terms: string[] = []
  if (q.action) terms.push(`action.like.${q.action}`)
  if (q.status) terms.push(`status.eq.${q.status}`)
  if (q.resource_type) terms.push(`resource_type.like.${q.resource_type}`)
  if (q.resource_id) terms.push(`resource_id.eq.${q.resource_id}`)
  if (q.user_login) terms.push(`user_login.like.${q.user_login}`)
  if (q.user_id != null) terms.push(`user_id.eq.${q.user_id}`)
  if (q.from) terms.push(`timestamp.gte.${q.from}`)
  if (q.to) terms.push(`timestamp.lte.${q.to}`)
  return terms.join('&')
}

function buildQuery(q: AuditListQuery): string {
  const params = new URLSearchParams()
  params.set('page_number', String(q.page ?? 1))
  if (q.page_size != null) params.set('page_size', String(q.page_size))
  const sq = buildSearchQuery(q)
  if (sq) params.set('search_query', sq)
  return `?${params.toString()}`
}

export const auditHttpService = {
  list: (query: AuditListQuery = {}) =>
    api.get<AuditListResponse>(`/audit${buildQuery(query)}`),
  get: (id: number) => api.get<AuditLog>(`/audit/${id}`),
}
