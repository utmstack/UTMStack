import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'

export function useTiEntity(id: string | null | undefined) {
  return useQuery({
    queryKey: ['ti', 'entity', id],
    queryFn: () => threatIntelHttpService.entity(id as string),
    enabled: !!id,
    staleTime: 5 * 60_000,
  })
}
