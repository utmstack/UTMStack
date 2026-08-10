import { createApiClient } from '@/shared/lib/api-client'
import type {
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  IntegrationResponse,
  DataTypeOption,
  ConfigGroupResponse,
} from '@/features/integrations/types'

export interface IntegrationsService {
  // Catalog
  listIntegrations(): Promise<IntegrationResponse[]>
  getIntegration(id: string): Promise<IntegrationResponse>
  createIntegration(data: CreateIntegrationRequest): Promise<IntegrationResponse>
  updateIntegration(id: string, data: UpdateIntegrationRequest): Promise<IntegrationResponse>
  deleteIntegration(id: string): Promise<void>
  getDataTypes(): Promise<DataTypeOption[]>

  // Configuration groups — one connector instance each, scoped to this tenant
  listConfigGroups(integration: string): Promise<ConfigGroupResponse[]>
  saveConfigGroup(integration: string, data: ConfigGroupResponse): Promise<void>
  deleteConfigGroup(integration: string, name: string): Promise<void>
}

const BASE = '/integrations'

export function createIntegrationsService(baseUrl?: string): IntegrationsService {
  const api = createApiClient(baseUrl)

  return {
    // The integrations page is a full catalog (cards/tabs), not a paginated table.
    // GET /integrations defaults to page size 20 (database.Params), so request the
    // backend max (MaxPageSize=200) to bring everything in one call.
    listIntegrations: () => api.get<IntegrationResponse[]>(`${BASE}?size=200`),

    getIntegration: (id) => api.get<IntegrationResponse>(`${BASE}/${id}`),

    createIntegration: (data) => api.post<IntegrationResponse>(BASE, data),

    updateIntegration: (id, data) => api.put<IntegrationResponse>(`${BASE}/${id}`, data),

    deleteIntegration: (id) => api.delete<void>(`${BASE}/${id}`),

    getDataTypes: () => api.get<DataTypeOption[]>(`${BASE}/data-types`),

    listConfigGroups: (integration) =>
      api.get<ConfigGroupResponse[]>(`${BASE}/config/${encodeURIComponent(integration)}`),

    saveConfigGroup: (integration, data) =>
      api.put<void>(`${BASE}/config/${encodeURIComponent(integration)}`, data),

    deleteConfigGroup: (integration, name) =>
      api.delete<void>(
        `${BASE}/config/${encodeURIComponent(integration)}/${encodeURIComponent(name)}`,
      ),
  }
}
