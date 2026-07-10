import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'

export function useTiFeeds() {
  return useQuery({
    queryKey: ['ti', 'feeds'],
    queryFn: () => threatIntelHttpService.feeds(),
    staleTime: 60_000,
  })
}
