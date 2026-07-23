import { useQuery } from '@tanstack/react-query'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { mergeAdvancedRequests } from '../services/advanced-query'
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

export function useTiIocsOverview(
  extra: AdvancedSearchRequest | undefined,
  interval: string,
) {
  const base: AdvancedSearchRequest = {
    query: { must: [{ terms: { 'type.keyword': IOC_TYPES } }] },
    aggs: {
      histogram: { date_histogram: { field: 'lastSeen', interval } },
      by_types: { terms: { field: 'type.keyword', size: 50 } },
    },
  }
  return useQuery({
    queryKey: ['ti', 'iocs-overview', extra, interval],
    queryFn: () => threatIntelHttpService.searchAdvanced(
      mergeAdvancedRequests(base, extra),
      { limit: 0, page: 1 },
    ),
    enabled: !!extra,
  })
}
