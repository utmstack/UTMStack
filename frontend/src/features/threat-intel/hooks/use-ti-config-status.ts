import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'

/**
 * Whether Threat Intelligence is set up on this instance.
 *
 * Asks the dedicated status endpoint rather than the usage counters: those are
 * the instance's own consumption across every tenant and are refused to anyone
 * but the operator, so a tenant probing them could not tell "not set up" from
 * "not yours to see".
 */
export function useTiConfigStatus(): { isConfigured: boolean | undefined; isLoading: boolean } {
  const q = useQuery({
    queryKey: ['ti', 'status'],
    queryFn: () => threatIntelHttpService.status(),
    staleTime: 60_000,
  })

  if (q.isLoading) return { isConfigured: undefined, isLoading: true }
  // Unreachable is not the same as unconfigured: let the page render and have
  // each call report its own failure, as it did before.
  if (q.isError) return { isConfigured: true, isLoading: false }
  return { isConfigured: q.data?.configured ?? true, isLoading: false }
}
