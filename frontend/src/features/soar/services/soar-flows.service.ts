import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { Flow, FlowListQuery, SaveFlowInput } from '../types/soar.types'

const api = createApiClient()

export { ApiError as SoarHttpError }

const BASE = '/soar/rules' // backend names flows "rules"; identity is relPath

function listQuery(q: FlowListQuery): string {
  const p = new URLSearchParams()
  if (q.name) p.set('name.contains', q.name)
  if (q.active != null) p.set('active.equals', String(q.active))
  if (q.systemOwner != null) p.set('systemOwner.equals', String(q.systemOwner))
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 20))
  return p.toString()
}

export const soarFlowsService = {
  // Returns { data: Flow[], total } — total from X-Total-Count.
  list: (q: FlowListQuery = {}) => api.getPaged<Flow[]>(`${BASE}?${listQuery(q)}`),
  get: (relPath: string) => api.get<Flow>(`${BASE}/${encodeURIComponent(relPath)}`),
  create: (input: SaveFlowInput) => api.post<Flow>(BASE, input),
  update: (relPath: string, input: SaveFlowInput) =>
    api.put<Flow>(`${BASE}/${encodeURIComponent(relPath)}`, input),
  remove: (relPath: string) => api.delete<{ message: string }>(`${BASE}/${encodeURIComponent(relPath)}`),
  setEnabled: (relPath: string, enabled: boolean) =>
    api.put<{ message: string }>(`${BASE}/${encodeURIComponent(relPath)}/enabled`, { enabled }),
}
