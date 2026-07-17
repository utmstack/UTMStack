import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createCollectorIntegrationService } from '@/features/integrations/service/collector-integration.service'
import type {
  ForwarderCollector,
  SetDataTypeConfigRequest,
  ConfigKnowledgeResponse,
  GetDataTypeConfigResponse,
  SetForwarderCertificatesRequest,
  TLSStatusResponse,
} from '@/features/integrations/types'

const COLLECTOR_QUERY_KEYS = {
  all: ['integrations', 'collectors'] as const,
  forwarders: () => [...COLLECTOR_QUERY_KEYS.all, 'forwarders'] as const,
  tlsStatus: (collectorId: number) => [...COLLECTOR_QUERY_KEYS.all, 'tls-status', collectorId] as const,
  dataTypeConfig: (collectorId: number, dataType: string) =>
    [...COLLECTOR_QUERY_KEYS.all, 'data-type-config', collectorId, dataType] as const,
}

export function useCollectorIntegration(baseUrl?: string) {
  const queryClient = useQueryClient()
  const service = createCollectorIntegrationService(baseUrl)

  const forwarders = useQuery({
    queryKey: COLLECTOR_QUERY_KEYS.forwarders(),
    queryFn: () => service.listForwarders(),
    refetchInterval: 15_000,
  })

  const setDataType = useMutation({
    mutationFn: ({
      collectorId,
      dataType,
      data,
    }: {
      collectorId: number
      dataType: string
      data: SetDataTypeConfigRequest
    }) => service.setDataType(collectorId, dataType, data),
    onSuccess: (_, { collectorId, dataType }) => {
      queryClient.invalidateQueries({
        queryKey: COLLECTOR_QUERY_KEYS.dataTypeConfig(collectorId, dataType),
      })
    },
  })

  const setCertificates = useMutation({
    mutationFn: ({
      collectorId,
      data,
    }: {
      collectorId: number
      data: SetForwarderCertificatesRequest
    }) => service.setCertificates(collectorId, data),
    onSuccess: (_, { collectorId }) => {
      queryClient.invalidateQueries({ queryKey: COLLECTOR_QUERY_KEYS.tlsStatus(collectorId) })
    },
  })

  const tlsStatus = (collectorId: number | null) =>
    useQuery<TLSStatusResponse>({
      queryKey: COLLECTOR_QUERY_KEYS.tlsStatus(collectorId ?? -1),
      queryFn: () => service.getTlsStatus(collectorId as number),
      enabled: collectorId != null,
    })

  const dataTypeConfig = (collectorId: number | null, dataType: string) =>
    useQuery<GetDataTypeConfigResponse>({
      queryKey: COLLECTOR_QUERY_KEYS.dataTypeConfig(collectorId ?? -1, dataType),
      queryFn: () => service.getDataTypeConfig(collectorId as number, dataType),
      enabled: collectorId != null && !!dataType,
    })

  return {
    forwarders,
    setDataType,
    setCertificates,
    tlsStatus,
    dataTypeConfig,
  }
}

export type { ForwarderCollector, SetDataTypeConfigRequest, ConfigKnowledgeResponse, GetDataTypeConfigResponse }
