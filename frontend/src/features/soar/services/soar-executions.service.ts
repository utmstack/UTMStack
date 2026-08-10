import { createApiClient } from '@/shared/lib/api-client'
import type { Execution, ExecutionListQuery } from '../types/soar.types'

const api = createApiClient()

const BASE = '/soar/rule-executions'

function listQuery(q: ExecutionListQuery): string {
  const p = new URLSearchParams()
  if (q.origin) p.set('origin', q.origin)
  if (q.rulePath) p.set('rulePath', q.rulePath)
  if (q.alertId) p.set('alertId', q.alertId)
  if (q.agent) p.set('agent', q.agent)
  if (q.triggeredBy) p.set('triggeredBy', q.triggeredBy)
  if (q.status) p.set('status', q.status)
  if (q.startedAtFrom) p.set('startedAtFrom', q.startedAtFrom)
  if (q.startedAtTo) p.set('startedAtTo', q.startedAtTo)
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 20))
  return p.toString()
}

export const soarExecutionsService = {
  // Returns { data: Execution[], total } — total from X-Total-Count.
  list: (q: ExecutionListQuery = {}) => api.getPaged<Execution[]>(`${BASE}?${listQuery(q)}`),
}
