import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { mergeAdvancedRequests } from '../services/advanced-query'
import { useAlertIocsFragment } from './use-alert-iocs'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

const BASE: AdvancedSearchRequest = {
  query: { must: [{ range: { lastSeen: { gte: 'now-24h', lte: 'now' } } }] },
  aggs: { hourly_iocs: { date_histogram: { field: 'lastSeen', interval: 'hour' } } },
}

export function useTiIocs24h() {
  const observed = useAlertIocsFragment()
  return useQuery({
    queryKey: ['ti', 'iocs-24h', observed],
    queryFn: () => threatIntelHttpService.searchAdvanced(
      mergeAdvancedRequests(BASE, observed),
      { limit: 0, page: 1 },
    ),
    enabled: !!observed,
  })
}
