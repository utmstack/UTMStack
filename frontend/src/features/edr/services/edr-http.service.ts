import { createApiClient } from '@/shared/lib/api-client'
import type { EdrActionResult, EdrEndpoint, EdrOverview, EdrQuarantineItem } from '../types/edr.types'

const api = createApiClient()

// Paths are relative to /api/v1; they map 1:1 to backend modules/edr routes.
// Responses are already the backend DTO shapes — no field aliasing/normalization.
export const edrHttpService = {
  overview: () => api.get<EdrOverview>('/edr/overview'),
  endpoints: () => api.get<EdrEndpoint[]>('/edr/endpoints'),
  endpoint: (agentId: number) => api.get<EdrEndpoint>(`/edr/endpoints/${agentId}`),
  quarantine: (agentId?: number) => api.get<EdrQuarantineItem[]>(`/edr/quarantine${agentId ? `?agentId=${agentId}` : ''}`),

  // Single-target dispatch. Returns the cmdId to poll with command().
  action: (agentId: number, action: string, payload?: Record<string, unknown>, reason?: string) =>
    api.post<EdrActionResult>(`/edr/endpoints/${agentId}/actions`, { action, payload, reason }),

  // Poll a dispatch by its cmdId: pending = working, error = endpoint refused,
  // 404 = never received (offline at dispatch time).
  command: (agentId: number, cmdId: string) =>
    api.get<{ cmdId: string; action: string; status: string; result?: string; executedBy: string }>(
      `/edr/endpoints/${agentId}/actions/${cmdId}`,
    ),

  // Bulk actions the UI offers: each becomes one dispatch per selected endpoint.
  bulkAction: (action: string, endpointIds: number[], payload?: Record<string, unknown>, reason?: string) =>
    Promise.all(endpointIds.map((id) => edrHttpService.action(id, action, payload, reason))),
}
