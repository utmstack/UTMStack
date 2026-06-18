import { useQuery } from '@tanstack/react-query'
import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

/** Subset of GET /soc-ai/config we need to decide whether the assistant is usable. */
interface SocAiConfigStatus {
  configured: boolean
}

/**
 * Whether SOC-AI has a usable provider configured (Settings → SOC-AI). The
 * floating composer and chat panel only make sense once a model is wired up, so
 * the UI gates on this. Cached for a few minutes; a 401/403/network error
 * resolves to "not configured" rather than throwing.
 */
export function useSocAiConfigured(): boolean {
  const { data } = useQuery({
    queryKey: ['soc-ai', 'config', 'configured'],
    queryFn: () => api.get<SocAiConfigStatus>('/soc-ai/config'),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
  return data?.configured ?? false
}
