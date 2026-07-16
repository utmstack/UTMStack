import { createApiClient } from '@/shared/lib/api-client'
import type {
  ForwarderCollector,
  SetDataTypeConfigRequest,
  ConfigKnowledgeResponse,
  SetForwarderCertificatesRequest,
  TLSStatusResponse,
} from '@/features/integrations/types'

// Talks to backend/modules/integrations/handler/collector.go — the remote
// collector control surface (list online forwarders, enable/disable a data
// type, push TLS certs, read TLS status). See dto/collector.go for the wire
// shapes mirrored in types/index.ts.
const BASE = '/integrations/collectors'

export interface CollectorIntegrationService {
  listForwarders(): Promise<ForwarderCollector[]>
  setDataType(
    collectorId: number,
    dataType: string,
    data: SetDataTypeConfigRequest,
  ): Promise<ConfigKnowledgeResponse>
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

    setCertificates: (collectorId, data) =>
      api.put<ConfigKnowledgeResponse>(`${BASE}/${collectorId}/certificates`, data),

    getTlsStatus: (collectorId) =>
      api.get<TLSStatusResponse>(`${BASE}/${collectorId}/tls-status`),
  }
}
