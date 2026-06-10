import { createApiClient } from '@/shared/lib/api-client'
import type {
  ModuleActivationRequest,
  CreateModuleRequest,
  UpdateModuleRequest,
  ModuleResponse,
  DataTypeOption,
} from '@/features/integrations/types'

export interface IntegrationsService {
  // Module operations
  listModules(): Promise<ModuleResponse[]>
  getModule(id: number): Promise<ModuleResponse>
  createModule(data: CreateModuleRequest): Promise<ModuleResponse>
  updateModule(id: number, data: UpdateModuleRequest): Promise<ModuleResponse>
  deleteModule(id: number): Promise<void>
  activateDeactivateModule(data: ModuleActivationRequest): Promise<ModuleResponse>

  // Module metadata
  getCategories(): Promise<string[]>
  getDataTypes(): Promise<DataTypeOption[]>
  isActive(moduleName: string): Promise<{ isActive: boolean }>

  // Tenant operations
  listTenants(moduleName: string): Promise<unknown>
  saveTenant(moduleName: string, data: unknown): Promise<unknown>
  deleteTenant(moduleName: string, name: string): Promise<void>
}

const BASE_INTEGRATIONS_URL='/integrations'

export function createIntegrationsService(baseUrl?: string): IntegrationsService {
  const api = createApiClient(baseUrl)

  return {
    listModules: () => api.get<ModuleResponse[]>(BASE_INTEGRATIONS_URL),

    getModule: (id: number) => api.get<ModuleResponse>(`${BASE_INTEGRATIONS_URL}${id}`),

    createModule: (data: CreateModuleRequest) =>
      api.post<ModuleResponse>(`${BASE_INTEGRATIONS_URL}`, data),

    updateModule: (id: number, data: UpdateModuleRequest) =>
      api.put<ModuleResponse>(`${BASE_INTEGRATIONS_URL}/${id}`, data),

    deleteModule: (id: number) => api.delete<void>(`${BASE_INTEGRATIONS_URL}/${id}`),

    activateDeactivateModule: (data: ModuleActivationRequest) =>
      api.put<ModuleResponse>(`${BASE_INTEGRATIONS_URL}/activate`, data),

    getCategories: () => api.get<string[]>(`${BASE_INTEGRATIONS_URL}/categories`),

    getDataTypes: () => api.get<DataTypeOption[]>(`${BASE_INTEGRATIONS_URL}/data-types`),

    isActive: (moduleName: string) =>
      api.get<{ isActive: boolean }>(`${BASE_INTEGRATIONS_URL}/is-active?moduleName=${moduleName}`),

    listTenants: (moduleName: string) =>
      api.get(`${BASE_INTEGRATIONS_URL}/tenants/:module`.replace(':module', moduleName)),

    saveTenant: (moduleName: string, data: unknown) =>
      api.put(`${BASE_INTEGRATIONS_URL}/tenants/:module`.replace(':module', moduleName), data),

    deleteTenant: (moduleName: string, name: string) =>
      api.delete<void>(
        `${BASE_INTEGRATIONS_URL}/tenants/:module/:name`
          .replace(':module', moduleName)
          .replace(':name', name)
      ),
  }
}
