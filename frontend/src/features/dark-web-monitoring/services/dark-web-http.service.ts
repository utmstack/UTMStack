import { createApiClient } from '@/shared/lib/api-client'
import type { BreachStatus, DarkWebRecord, OrganizationBreachStats } from '../types'

/**
 * HTTP client for the dark web monitoring API. The backend endpoints do not
 * exist yet, so the page fetches through `darkWebMonitoringService` (see
 * `services/index.ts`), which serves fake data until `USE_MOCK_DARK_WEB` is
 * unset. Once the API lands, flip the flag and these endpoints become live.
 */
const api = createApiClient()
const BASE = '/dark-web'

export const darkWebHttpService = {
  /** Breach records for the organization, newest first. */
  listRecords: () => api.get<DarkWebRecord[]>(`${BASE}/records`),
  /** Aggregated breach counts per organization element type. */
  getOrganizationStats: () => api.get<OrganizationBreachStats>(`${BASE}/organization-stats`),
  /** Update a breach's handling status (open → in mitigation → completed). */
  updateRecordStatus: (id: string, breachStatus: BreachStatus) =>
    api.put<DarkWebRecord>(`${BASE}/records/${encodeURIComponent(id)}/status`, { breachStatus }),
}
