import { useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createIntegrationsService } from '@/features/integrations/service/integrations.service'
import type {
  ModuleResponse,
  CreateModuleRequest,
  UpdateModuleRequest,
  ModuleActivationRequest,
  DataTypeOption,
} from '@/features/integrations/types'

const INTEGRATIONS_QUERY_KEYS = {
  all: ['integrations'] as const,
  modules: () => [...INTEGRATIONS_QUERY_KEYS.all, 'modules'] as const,
  module: (id: number) => [...INTEGRATIONS_QUERY_KEYS.modules(), id] as const,
  categories: () => [...INTEGRATIONS_QUERY_KEYS.all, 'categories'] as const,
  dataTypes: () => [...INTEGRATIONS_QUERY_KEYS.all, 'data-types'] as const,
  isActive: (moduleName: string) => [...INTEGRATIONS_QUERY_KEYS.all, 'is-active', moduleName] as const,
  tenants: (moduleName: string) => [...INTEGRATIONS_QUERY_KEYS.all, 'tenants', moduleName] as const,
}

export interface UseIntegrationsResult {
  // Queries
  modules: ReturnType<typeof useQuery<ModuleResponse[]>>
  module: (id: number) => ReturnType<typeof useQuery<ModuleResponse>>
  categories: ReturnType<typeof useQuery<string[]>>
  dataTypes: ReturnType<typeof useQuery<DataTypeOption[]>>
  isActive: (moduleName: string) => ReturnType<typeof useQuery<{ isActive: boolean }>>
  tenants: (moduleName: string) => ReturnType<typeof useQuery>

  // Mutations
  createModule: ReturnType<typeof useMutation<ModuleResponse, Error, CreateModuleRequest>>
  updateModule: ReturnType<
    typeof useMutation<ModuleResponse, Error, { id: number; data: UpdateModuleRequest }>
  >
  deleteModule: ReturnType<typeof useMutation<void, Error, number>>
  activateDeactivateModule: ReturnType<
    typeof useMutation<ModuleResponse, Error, ModuleActivationRequest>
  >
  saveTenant: ReturnType<typeof useMutation<unknown, Error, { moduleName: string; data: unknown }>>
  deleteTenant: ReturnType<
    typeof useMutation<void, Error, { moduleName: string; name: string }>
  >

  // Loading state
  isLoading: boolean
}

export function useIntegrations(baseUrl?: string): UseIntegrationsResult {
  const queryClient = useQueryClient()
  const service = createIntegrationsService(baseUrl)

  // Queries
  const modules = useQuery({
    queryKey: INTEGRATIONS_QUERY_KEYS.modules(),
    queryFn: () => service.listModules(),
  })

  const module = (id: number) =>
    useQuery({
      queryKey: INTEGRATIONS_QUERY_KEYS.module(id),
      queryFn: () => service.getModule(id),
      enabled: id > 0,
    })

  const categories = useQuery({
    queryKey: INTEGRATIONS_QUERY_KEYS.categories(),
    queryFn: () => service.getCategories(),
  })

  const dataTypes = useQuery({
    queryKey: INTEGRATIONS_QUERY_KEYS.dataTypes(),
    queryFn: () => service.getDataTypes(),
  })

  const isActive = (moduleName: string) =>
    useQuery({
      queryKey: INTEGRATIONS_QUERY_KEYS.isActive(moduleName),
      queryFn: () => service.isActive(moduleName),
      enabled: !!moduleName,
    })

  const tenants = (moduleName: string) =>
    useQuery({
      queryKey: INTEGRATIONS_QUERY_KEYS.tenants(moduleName),
      queryFn: () => service.listTenants(moduleName),
      enabled: !!moduleName,
    })

  // Mutations
  const createModule = useMutation({
    mutationFn: (data: CreateModuleRequest) => service.createModule(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.modules() })
    },
  })

  const updateModule = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateModuleRequest }) =>
      service.updateModule(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.module(id) })
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.modules() })
    },
  })

  const deleteModule = useMutation({
    mutationFn: (id: number) => service.deleteModule(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.modules() })
    },
  })

  const activateDeactivateModule = useMutation({
    mutationFn: (data: ModuleActivationRequest) => service.activateDeactivateModule(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.modules() })
    },
  })

  const saveTenant = useMutation({
    mutationFn: ({ moduleName, data }: { moduleName: string; data: unknown }) =>
      service.saveTenant(moduleName, data),
    onSuccess: (_, { moduleName }) => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.tenants(moduleName) })
    },
  })

  const deleteTenant = useMutation({
    mutationFn: ({ moduleName, name }: { moduleName: string; name: string }) =>
      service.deleteTenant(moduleName, name),
    onSuccess: (_, { moduleName }) => {
      queryClient.invalidateQueries({ queryKey: INTEGRATIONS_QUERY_KEYS.tenants(moduleName) })
    },
  })

  const isLoading = useMemo(
    () =>
      modules.isLoading ||
      categories.isLoading ||
      dataTypes.isLoading ||
      createModule.isPending ||
      updateModule.isPending ||
      deleteModule.isPending ||
      activateDeactivateModule.isPending ||
      saveTenant.isPending ||
      deleteTenant.isPending,
    [
      modules.isLoading,
      categories.isLoading,
      dataTypes.isLoading,
      createModule.isPending,
      updateModule.isPending,
      deleteModule.isPending,
      activateDeactivateModule.isPending,
      saveTenant.isPending,
      deleteTenant.isPending,
    ]
  )

  return {
    // Queries
    modules,
    module,
    categories,
    dataTypes,
    isActive,
    tenants,

    // Mutations
    createModule,
    updateModule,
    deleteModule,
    activateDeactivateModule,
    saveTenant,
    deleteTenant,

    // Loading state
    isLoading,
  }
}
