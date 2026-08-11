import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createQueryService } from '@/features/dashboard/service/query.service'
import { resolveRangeToISO } from '@/features/dashboard/utils/datemath'
import {
  filterTypesToSpec,
  parseSpec,
  resultToRows,
  specIsComplete,
} from '@/features/dashboard/utils/spec'
import type { FilterType, Row, Visualization, VizSpec } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

const EMPTY_FILTERS: FilterType[] = []

export const VISUALIZATION_DATA_QUERY_KEYS = {
  all: ['visualization-data'] as const,
  forViz: (vizId: string, fromISO: string, toISO: string, filtersKey: string) =>
    [...VISUALIZATION_DATA_QUERY_KEYS.all, vizId, fromISO, toISO, filtersKey] as const,
}

export interface VisualizationData {
  rows: Row[]
  total: number
}

/**
 * The dashboard's time range and its filter bar are not part of a saved widget:
 * they are what the person looking at it chose, so they are merged into the
 * spec on the way out rather than stored with it.
 */
export function scopeSpec(spec: VizSpec, fromISO: string, toISO: string, filters: FilterType[]): VizSpec {
  const bar = filterTypesToSpec(filters)
  return {
    ...spec,
    from: fromISO,
    to: toISO,
    filters: bar.length > 0 ? [...(spec.filters ?? []), ...bar] : spec.filters,
  }
}

export function useVisualizationData(
  visualization: Visualization | null,
  time: TimeRange,
  filters: FilterType[] = EMPTY_FILTERS,
  refetchIntervalMs: number | null = null
) {
  const service = useMemo(() => createQueryService(), [])

  const { fromISO, toISO } = useMemo(
    () => resolveRangeToISO(time.from, time.to),
    [time.from, time.to]
  )

  const filtersKey = useMemo(() => JSON.stringify(filters), [filters])

  const spec = useMemo(() => {
    const parsed = parseSpec(visualization?.spec)
    if (!parsed || !specIsComplete(parsed)) return null
    return scopeSpec(parsed, fromISO, toISO, filters)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visualization?.spec, fromISO, toISO, filtersKey])

  return useQuery<VisualizationData>({
    queryKey: visualization
      ? VISUALIZATION_DATA_QUERY_KEYS.forViz(visualization.id, fromISO, toISO, filtersKey)
      : [...VISUALIZATION_DATA_QUERY_KEYS.all, 'noop'],
    queryFn: async () => {
      if (!spec) return { rows: [], total: 0 }
      const result = await service.run(spec)
      const rows = resultToRows(result, spec)
      return { rows, total: result.total ?? rows.length }
    },
    enabled: visualization != null && spec != null,
    staleTime: 30_000,
    refetchInterval: refetchIntervalMs ?? false,
  })
}
