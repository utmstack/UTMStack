import { useQuery } from '@tanstack/react-query'
import { connectionKeyHttpService } from '@/features/settings/services/connection-key-http.service'

const CONNECTION_QUERY_KEY = ['integrations', 'connectionkey_query'] as const

export function useConectionKey() {

  const query = useQuery({
    queryKey: CONNECTION_QUERY_KEY,
    queryFn: () => connectionKeyHttpService.get(),
    staleTime: 5 * 60 * 1000,
  })

  return {
    key: query,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
  }
}
