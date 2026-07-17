import { createApiClient } from '@/shared/lib/api-client'
import type {
  ForwarderCollector,
  SetDataTypeConfigRequest,
  ConfigKnowledgeResponse,
  GetDataTypeConfigResponse,
  SetForwarderCertificatesRequest,
  TLSStatusResponse,
} from '@/features/integrations/types'

const BASE = '/integrations/collectors'

export interface CollectorIntegrationService {
  listForwarders(): Promise<ForwarderCollector[]>
  setDataType(
    collectorId: number,
    dataType: string,
    data: SetDataTypeConfigRequest,
  ): Promise<ConfigKnowledgeResponse>
  getDataTypeConfig(collectorId: number, dataType: string): Promise<GetDataTypeConfigResponse>
  setCertificates(
    collectorId: number,
    data: SetForwarderCertificatesRequest,
  ): Promise<ConfigKnowledgeResponse>
  getTlsStatus(collectorId: number): Promise<TLSStatusResponse>
}

export function createCollectorIntegrationService(baseUrl?: string): CollectorIntegrationService {
  const api = createApiClient(baseUrl)

  return {
    listForwarders: () => api.get<ForwarderCollector[]>(BASE),

    setDataType: (collectorId, dataType, data) =>
      api.put<ConfigKnowledgeResponse>(
        `${BASE}/${collectorId}/data-types/${encodeURIComponent(dataType)}`,
        data,
      ),

    getDataTypeConfig: (collectorId, dataType) =>
      api.get<GetDataTypeConfigResponse>(
        `${BASE}/${collectorId}/data-types/${encodeURIComponent(dataType)}`,
      ),

    setCertificates: (collectorId, data) =>
      api.put<ConfigKnowledgeResponse>(`${BASE}/${collectorId}/certificates`, data),

    getTlsStatus: (collectorId) =>
      api.get<TLSStatusResponse>(`${BASE}/${collectorId}/tls-status`),
  }
}
