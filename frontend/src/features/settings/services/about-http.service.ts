import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ClusterHealth,
  DatasourceUsage,
  LicenseInfo,
  McpHealth,
  VersionInfo,
} from '../types/about.types'

const api = createApiClient()

export { ApiError as AboutHttpError }

/**
 * Each call is independent — the About page loads sections in parallel and tolerates
 * individual failures (e.g. OpenSearch returns 204 when no cluster info is available).
 * All endpoints already exist in the backend; the page adds no new server surface.
 */
export const aboutHttpService = {
  version: () => api.get<VersionInfo>('/billing/version'),
  license: () => api.get<LicenseInfo>('/billing/license'),
  datasourceUsage: () => api.get<DatasourceUsage>('/datasources/usage'),
  // 204 No Content → no cluster info; api.get yields undefined which we normalize to null.
  clusterHealth: async (): Promise<ClusterHealth | null> => {
    const resp = await api.get<{ health: ClusterHealth } | ''>('/opensearch/cluster/status')
    return resp && typeof resp === 'object' ? (resp.health ?? null) : null
  },
  mcpHealth: () => api.get<McpHealth>('/mcp/health'),
}
