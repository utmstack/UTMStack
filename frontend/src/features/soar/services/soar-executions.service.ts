import { createApiClient } from '@/shared/lib/api-client'
import type { Execution, ExecutionListQuery } from '../types/soar.types'

const api = createApiClient()

const BASE = '/soar/rule-executions'

function listQuery(q: ExecutionListQuery): string {
  const p = new URLSearchParams()
  if (q.rulePath) p.set('rulePath.equals', q.rulePath)
  if (q.alertId) p.set('alertId.contains', q.alertId)
  if (q.agent) p.set('agent.contains', q.agent)
  if (q.executionStatus) p.set('executionStatus.equals', q.executionStatus)
  if (q.executionDateGte) p.set('executionDate.greaterThanOrEqual', q.executionDateGte)
  if (q.executionDateLte) p.set('executionDate.lessThanOrEqual', q.executionDateLte)
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 20))
  return p.toString()
}

export const soarExecutionsService = {
  // Returns { data: Execution[], total } — total from X-Total-Count.
  list: (q: ExecutionListQuery = {}) => api.getPaged<Execution[]>(`${BASE}?${listQuery(q)}`),
}
