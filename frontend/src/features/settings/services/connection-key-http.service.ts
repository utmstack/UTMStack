import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as ConnectionKeyHttpError }

export interface ConnectionKeyResponse {
  connectionKey: string
}

/**
 * The agent-installation connection key. It is owned by the agent-manager (the
 * backend proxies it via the datasources module — "agents are datasources") and
 * is ONLY used by agents when they register with this instance. Admin only.
 */
export const connectionKeyHttpService = {
  get: () => api.get<ConnectionKeyResponse>('/datasources/connection-key'),
  rotate: () => api.post<ConnectionKeyResponse>('/datasources/connection-key/rotate'),
}
