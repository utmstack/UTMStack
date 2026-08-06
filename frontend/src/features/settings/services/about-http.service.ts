import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { DatasourceUsage, LicenseInfo, McpHealth, VersionInfo } from '../types/about.types'

const api = createApiClient()

export { ApiError as AboutHttpError }

/**
 * Each call is independent — the About page loads sections in parallel and tolerates
 * individual failures. All endpoints already exist in the backend; the page adds no
 * new server surface.
 */
export const aboutHttpService = {
  version: () => api.get<VersionInfo>('/billing/version'),
  license: () => api.get<LicenseInfo>('/billing/license'),
  datasourceUsage: () => api.get<DatasourceUsage>('/datasources/count'),
  mcpHealth: () => api.get<McpHealth>('/mcp/health'),
}
