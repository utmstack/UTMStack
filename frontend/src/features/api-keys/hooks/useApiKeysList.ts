import { useQuery } from '@tanstack/react-query'
import { apiKeysHttpService } from '../services/api-keys-http.service'

export const API_KEYS_LIST_QUERY_KEY = ['api-keys', 'list'] as const

export function useApiKeysList() {
  return useQuery({
    queryKey: API_KEYS_LIST_QUERY_KEY,
    queryFn: () => apiKeysHttpService.list(1, 200),
    staleTime: 30_000,
  })
}
