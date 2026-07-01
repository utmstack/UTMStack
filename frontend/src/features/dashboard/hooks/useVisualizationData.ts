import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createOpenSearchService, type Row } from '@/features/dashboard/service/opensearch.service'
import { resolveRangeToISO } from '@/features/dashboard/utils/datemath'
import { applySqlTemplate } from '@/features/dashboard/utils/sql-template'
import type { FilterType, Visualization } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

const EMPTY_FILTERS: FilterType[] = []

export const VISUALIZATION_DATA_QUERY_KEYS = {
  all: ['visualization-data'] as const,
  forViz: (vizId: number, fromISO: string, toISO: string, filtersKey: string) =>
    [...VISUALIZATION_DATA_QUERY_KEYS.all, vizId, fromISO, toISO, filtersKey] as const,
}

export interface VisualizationData {
  rows: Row[]
  total: number
}

export function useVisualizationData(
  visualization: Visualization | null,
  time: TimeRange,
  filters: FilterType[] = EMPTY_FILTERS
) {
  const service = useMemo(() => createOpenSearchService(), [])

  const { fromISO, toISO } = useMemo(
    () => resolveRangeToISO(time.from, time.to),
    [time.from, time.to]
  )

  const resolvedQuery = useMemo(() => {
    if (!visualization?.sqlQuery?.trim()) return null
    return applySqlTemplate({ sql: visualization.sqlQuery, fromISO, toISO })
  }, [visualization?.sqlQuery, fromISO, toISO])

  const filtersKey = useMemo(() => JSON.stringify(filters), [filters])

  return useQuery<VisualizationData>({
    queryKey: visualization
      ? VISUALIZATION_DATA_QUERY_KEYS.forViz(visualization.id, fromISO, toISO, filtersKey)
      : [...VISUALIZATION_DATA_QUERY_KEYS.all, 'noop'],
    queryFn: async () => {
      if (!resolvedQuery) return { rows: [], total: 0 }
      const { data, total } = await service.searchSql({ query: resolvedQuery, filters })
      return { rows: data, total }
    },
    enabled: visualization != null && resolvedQuery != null,
    staleTime: 30_000,
  })
}
