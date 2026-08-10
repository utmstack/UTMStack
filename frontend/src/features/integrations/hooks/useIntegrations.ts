import { useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createIntegrationsService } from '@/features/integrations/service/integrations.service'
import type {
  IntegrationResponse,
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  DataTypeOption,
  ConfigGroupResponse,
} from '@/features/integrations/types'

const INTEGRATIONS_QUERY_KEYS = {
  all: ['integrations'] as const,
  list: () => [...INTEGRATIONS_QUERY_KEYS.all, 'list'] as const,
  one: (id: string) => [...INTEGRATIONS_QUERY_KEYS.list(), id] as const,
  dataTypes: () => [...INTEGRATIONS_QUERY_KEYS.all, 'data-types'] as const,
  configGroups: (integration: string) =>
    [...INTEGRATIONS_QUERY_KEYS.all, 'config', integration] as const,
}

export interface UseIntegrationsResult {
  // Queries
  integrations: ReturnType<typeof useQuery<IntegrationResponse[]>>
  integration: (id: string) => ReturnType<typeof useQuery<IntegrationResponse>>
  dataTypes: ReturnType<typeof useQuery<DataTypeOption[]>>
  configGroups: (integration: string) => ReturnType<typeof useQuery<ConfigGroupResponse[]>>

  // Mutations
  createIntegration: ReturnType<
    typeof useMutation<IntegrationResponse, Error, CreateIntegrationRequest>
  >
  updateIntegration: ReturnType<
    typeof useMutation<IntegrationResponse, Error, { id: string; data: UpdateIntegrationRequest }>
  >
  deleteIntegration: ReturnType<typeof useMutation<void, Error, string>>
  saveConfigGroup: ReturnType<
    typeof useMutation<void, Error, { integration: string; data: ConfigGroupResponse }>
  >
  deleteConfigGroup: ReturnType<
    typeof useMutation<void, Error, { integration: string; name: string }>
  >

  // Loading state
  isLoading: boolean
}

export function useIntegrations(baseUrl?: string): UseIntegrationsResult {
  const queryClient = useQueryClient()
  const service = createIntegrationsService(baseUrl)

  const integrations = useQuery({
    queryKey: INTEGRATIONS_QUERY_KEYS.list(),
    queryFn: () => service.listIntegrations(),
  })

  const integration = (id: string) =>
    useQuery({
      queryKey: INTEGRATIONS_QUERY_KEYS.one(id),
      queryFn: () => service.getIntegration(id),
      enabled: !!id,
    })

  const dataTypes = useQuery({
    queryKey: INTEGRATIONS_QUERY_KEYS.dataTypes(),
    queryFn: () => service.getDataTypes(),
  })

  const configGroups = (name: string) =>
    useQuery({
      queryKey: INTEGRATIONS_QUERY_KEYS.configGroups(name),
      queryFn: () => service.listConfigGroups(name),
      enabled: !!name,
    })

  const createIntegration = useMutation({
    mutationFn: (data: CreateIntegrationRequest) => service.createIntegration(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.list() })
    },
  })

  const updateIntegration = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateIntegrationRequest }) =>
      service.updateIntegration(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.one(id) })
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.list() })
    },
  })

  const deleteIntegration = useMutation({
    mutationFn: (id: string) => service.deleteIntegration(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.list() })
    },
  })

  // A saved group is what makes an integration count as configured, and the
  // catalog cards read that state, so both caches have to be refreshed.
  const saveConfigGroup = useMutation({
    mutationFn: ({ integration: name, data }: { integration: string; data: ConfigGroupResponse }) =>
      service.saveConfigGroup(name, data),
    onSuccess: (_, { integration: name }) => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.configGroups(name) })
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.list() })
    },
  })

  const deleteConfigGroup = useMutation({
    mutationFn: ({ integration: name, name: groupName }: { integration: string; name: string }) =>
      service.deleteConfigGroup(name, groupName),
    onSuccess: (_, { integration: name }) => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.configGroups(name) })
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.list() })
    },
  })

  const isLoading = useMemo(
    () =>
      integrations.isLoading ||
      dataTypes.isLoading ||
      createIntegration.isPending ||
      updateIntegration.isPending ||
      deleteIntegration.isPending ||
      saveConfigGroup.isPending ||
      deleteConfigGroup.isPending,
    [
      integrations.isLoading,
      dataTypes.isLoading,
      createIntegration.isPending,
      updateIntegration.isPending,
      deleteIntegration.isPending,
      saveConfigGroup.isPending,
      deleteConfigGroup.isPending,
    ]
  )

  return {
    integrations,
    integration,
    dataTypes,
    configGroups,

    createIntegration,
    updateIntegration,
    deleteIntegration,
    saveConfigGroup,
    deleteConfigGroup,

    isLoading,
  }
}
