import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { mergeAdvancedRequests } from '../services/advanced-query'
import { useAlertIocsFragment } from './use-alert-iocs'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

const IOC_TYPES = [
  'domain', 'hostname', 'url', 'link', 'github-organization', 'github-repository',
  'ip', 'cidr', 'malware',
  'sha1', 'sha224', 'sha256', 'sha384', 'sha512', 'sha512-224', 'sha512-256',
  'sha3-224', 'sha3-256', 'sha3-384', 'sha3-512',
  'authentihash', 'cdhash', 'md5',
  'profile-photo', 'facebook-profile', 'tiktok-profile', 'twitter-profile',
  'filename',
]

const BASE: AdvancedSearchRequest = {
  query: {
    must: [
      { range: { lastSeen: { gte: 'now-24h', lte: 'now' } } },
      { terms: { 'type.keyword': IOC_TYPES } },
    ],
  },
  aggs: {
    hourly_iocs: { date_histogram: { field: 'lastSeen', interval: 'hour' } },
    by_types: { terms: { field: 'type.keyword', size: 50 } },
  },
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
