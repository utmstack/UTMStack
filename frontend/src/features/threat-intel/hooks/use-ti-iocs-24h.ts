import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

const REQ: AdvancedSearchRequest = {
  query: { must: [{ range: { lastSeen: { gte: 'now-24h', lte: 'now' } } }] },
  aggs: { hourly_iocs: { date_histogram: { field: 'lastSeen', interval: 'hour' } } },
}

export function useTiIocs24h() {
  return useQuery({
    queryKey: ['ti', 'iocs-24h'],
    queryFn: () => threatIntelHttpService.searchAdvanced(REQ, { limit: 0, page: 1 }),
  })
}
