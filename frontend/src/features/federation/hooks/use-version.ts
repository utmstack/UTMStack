import { useQuery } from '@tanstack/react-query'
import { createApiClient } from '@/shared/lib/api-client'
import { FS_API_URL, IS_FEDERATION } from '@/shared/config/mode'

const api = createApiClient(FS_API_URL)

/** The Federation Service build version (GET /fs/version). Only fetched in FS mode. */
export function useFederationVersion(): string | undefined {
  const { data } = useQuery({
    queryKey: ['fs', 'version'],
    queryFn: () => api.get<{ version: string }>('/version'),
    staleTime: Infinity,
    retry: false,
    enabled: IS_FEDERATION,
  })
  return data?.version
}
