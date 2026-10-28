import { createApiClient } from '@/shared/lib/api-client'
import type { BreachStatus, DarkWebRecord, OrganizationBreachStats } from '../types'

/**
 * HTTP client for the dark web monitoring API.
 *
 * The backend endpoints do not exist yet, so the page fetches through
 * `darkWebMonitoringService` (see `services/index.ts`), which serves the mock
 * until `USE_MOCK` is unset. The paths and shapes below are a *proposed*
 * contract mirroring `types.ts`; before flipping the flag to live, align them
 * with the actual backend implementation (route prefixes, verbs, payloads) —
 * do not treat them as the final API.
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
