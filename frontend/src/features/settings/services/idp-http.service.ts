import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { GroupMapping, IdentityProvider, IdentityProviderRequest } from '../types/idp.types'

const api = createApiClient()

export { ApiError as IdpHttpError }

export const idpHttpService = {
  // size=200 (backend max) — there are only ever a handful of IdPs, so no paging UI.
  list: () => api.get<IdentityProvider[]>('/identity-providers?page=0&size=200'),
  create: (input: IdentityProviderRequest) => api.post<IdentityProvider>('/identity-providers', input),
  update: (input: IdentityProviderRequest) => api.put<IdentityProvider>('/identity-providers', input),
  remove: (id: string) => api.delete<void>(`/identity-providers/${id}`),
  mappings: (id: string) => api.get<GroupMapping[]>(`/identity-providers/${id}/group-mappings`),
}
