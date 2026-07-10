import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'

export function useTiUsage() {
  return useQuery({
    queryKey: ['ti', 'usage'],
    queryFn: () => threatIntelHttpService.usage(),
    staleTime: 60_000,
  })
}
