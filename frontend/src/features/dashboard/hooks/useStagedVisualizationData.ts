import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createOpenSearchService, type Row } from '@/features/dashboard/service/opensearch.service'
import { resolveRangeToISO } from '@/features/dashboard/utils/datemath'
import { applySqlTemplate } from '@/features/dashboard/utils/sql-template'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export interface StagedVisualizationData {
  rows: Row[]
  total: number
}

export function useStagedVisualizationData(sql: string, time: TimeRange) {
  const service = useMemo(() => createOpenSearchService(), [])
  const { fromISO, toISO } = useMemo(
    () => resolveRangeToISO(time.from, time.to),
    [time.from, time.to]
  )
  const resolved = useMemo(() => {
    if (!sql.trim()) return null
    return applySqlTemplate({ sql, fromISO, toISO })
  }, [sql, fromISO, toISO])

  return useQuery<StagedVisualizationData>({
    queryKey: ['staged-visualization-data', sql, fromISO, toISO],
    queryFn: async () => {
      if (!resolved) return { rows: [], total: 0 }
      const { data, total } = await service.searchSql({ query: resolved })
      return { rows: data, total }
    },
    enabled: resolved != null,
    staleTime: 30_000,
    retry: false,
  })
}
