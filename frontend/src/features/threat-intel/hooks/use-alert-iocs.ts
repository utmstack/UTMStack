import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

const IOC_FIELD_MAP: { field: string; twAttr: string }[] = [
  { field: 'adversary.ip',     twAttr: 'ip' },
  { field: 'target.ip',        twAttr: 'ip' },
  { field: 'adversary.host',   twAttr: 'hostname' },
  { field: 'target.host',      twAttr: 'hostname' },
  { field: 'adversary.domain', twAttr: 'domain' },
  { field: 'target.domain',    twAttr: 'domain' },
  { field: 'adversary.hash',   twAttr: 'hash' },
  { field: 'target.hash',      twAttr: 'hash' },
  { field: 'adversary.url',    twAttr: 'url' },
  { field: 'target.url',       twAttr: 'url' },
  { field: 'adversary.email',  twAttr: 'email' },
  { field: 'target.email',     twAttr: 'email' },
]

const TOP_TOTAL = 1000

export interface AlertIocs {
  byAttr: Record<string, string[]>
  total: number
}

export function useAlertIocs() {
  return useQuery<AlertIocs>({
    queryKey: ['ti', 'alert-iocs', TOP_TOTAL],
    queryFn: async () => {
      const perField = await Promise.all(
        IOC_FIELD_MAP.map(async ({ field, twAttr }) => ({
          twAttr,
          values: await alertsHttpService.fieldValues(field, TOP_TOTAL).catch(() => []),
        })),
      )
      const flat = perField.flatMap(({ twAttr, values }) =>
        values.map((v) => ({ twAttr, value: v.value, count: v.count })),
      )
      flat.sort((a, b) => b.count - a.count)
      const byAttr: Record<string, string[]> = {}
      let total = 0
      for (const entry of flat) {
        if (total >= TOP_TOTAL) break
        const bucket = byAttr[entry.twAttr] ?? []
        if (bucket.includes(entry.value)) continue
        bucket.push(entry.value)
        byAttr[entry.twAttr] = bucket
        total++
      }
      return { byAttr, total }
    },
    staleTime: 5 * 60_000,
  })
}

export function alertIocsFragment(iocs: AlertIocs): AdvancedSearchRequest {
  const entries = Object.entries(iocs.byAttr).filter(([, v]) => v.length > 0)
  if (entries.length === 0) {
    return { query: { must: [{ terms: { id: ['__no_observed_iocs__'] } }] } }
  }
  return {
    query: {
      should: entries.map(([attr, values]) => ({
        terms: { [`attributes.${attr}`]: values },
      })),
    },
  }
}

export function useAlertIocsFragment(): AdvancedSearchRequest | undefined {
  const q = useAlertIocs()
  return useMemo(() => (q.data ? alertIocsFragment(q.data) : undefined), [q.data])
}
