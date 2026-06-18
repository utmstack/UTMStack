import { createApiClient } from '@/shared/lib/api-client'
import { FS_API_URL } from '@/shared/config/mode'
import type { FederationInstance, FederationInstanceInput } from '../types'

const api = createApiClient(FS_API_URL)

export const federationInstancesService = {
  list: () => api.get<FederationInstance[]>('/instances'),
  get: (id: number) => api.get<FederationInstance>(`/instances/${id}`),
  create: (input: FederationInstanceInput) => api.post<FederationInstance>('/instances', input),
  update: (id: number, input: FederationInstanceInput) =>
    api.put<FederationInstance>(`/instances/${id}`, input),
  remove: (id: number) => api.delete<void>(`/instances/${id}`),
}
