import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createCollectorIntegrationService } from '@/features/integrations/service/collector-integration.service'
import type {
  ForwarderCollector,
  SetDataTypeConfigRequest,
  ConfigKnowledgeResponse,
  SetForwarderCertificatesRequest,
  TLSStatusResponse,
} from '@/features/integrations/types'

const COLLECTOR_QUERY_KEYS = {
  all: ['integrations', 'collectors'] as const,
  forwarders: () => [...COLLECTOR_QUERY_KEYS.all, 'forwarders'] as const,
  tlsStatus: (collectorId: number) => [...COLLECTOR_QUERY_KEYS.all, 'tls-status', collectorId] as const,
}

export function useCollectorIntegration(baseUrl?: string) {
  const queryClient = useQueryClient()
  const service = createCollectorIntegrationService(baseUrl)

  // Online forwarders eligible for remote control (the picker feed). Polled
  // so a forwarder that just went online/offline is picked up without a
  // manual refresh.
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

  // Per-forwarder TLS status. Mirrors the tenants(moduleName)/isActive(name)
  // pattern in useIntegrations.ts — a query factory called with the currently
  // selected id, re-created on every render but keyed/cached by react-query.
  const tlsStatus = (collectorId: number | null) =>
    useQuery<TLSStatusResponse>({
      queryKey: COLLECTOR_QUERY_KEYS.tlsStatus(collectorId ?? -1),
      queryFn: () => service.getTlsStatus(collectorId as number),
      enabled: collectorId != null,
    })

  return {
    forwarders,
    setDataType,
    setCertificates,
    tlsStatus,
  }
}

export type { ForwarderCollector, SetDataTypeConfigRequest, ConfigKnowledgeResponse }
