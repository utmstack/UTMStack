import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ApiKey,
  ApiKeyGenerateResponse,
  ApiKeyListResponse,
  ApiKeyUpsertRequest,
} from '../types/api-key.types'

const api = createApiClient()

export { ApiError as ApiKeyHttpError }

export const apiKeysHttpService = {
  list: (page = 1, pageSize = 50) =>
    api.get<ApiKeyListResponse>(`/api-keys?page=${page}&page_size=${pageSize}`),
  create: (input: ApiKeyUpsertRequest) => api.post<ApiKey>('/api-keys', input),
  update: (id: number, input: ApiKeyUpsertRequest) => api.put<ApiKey>(`/api-keys/${id}`, input),
  remove: (id: number) => api.delete<void>(`/api-keys/${id}`),
  // Rotate: issues a fresh secret and returns it once. The previous value stops
  // working immediately. This is also how a brand-new key's secret is revealed,
  // since POST /api-keys does not return the secret.
  generate: (id: number) => api.post<ApiKeyGenerateResponse>(`/api-keys/${id}/generate`),
}
