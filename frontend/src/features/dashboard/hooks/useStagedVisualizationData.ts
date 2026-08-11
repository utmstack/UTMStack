import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createQueryService } from '@/features/dashboard/service/query.service'
import { resolveRangeToISO } from '@/features/dashboard/utils/datemath'
import { resultToRows, specIsComplete } from '@/features/dashboard/utils/spec'
import { scopeSpec } from '@/features/dashboard/hooks/useVisualizationData'
import type { Row, VizSpec } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export interface StagedVisualizationData {
  rows: Row[]
  total: number
}

/** The editor's preview: the spec being built, answered as it changes. */
export function useStagedVisualizationData(spec: VizSpec | null, time: TimeRange) {
  const service = useMemo(() => createQueryService(), [])
  const { fromISO, toISO } = useMemo(
    () => resolveRangeToISO(time.from, time.to),
    [time.from, time.to]
  )

  const scoped = useMemo(() => {
    if (!spec || !specIsComplete(spec)) return null
    return scopeSpec(spec, fromISO, toISO, [])
  }, [spec, fromISO, toISO])

  return useQuery<StagedVisualizationData>({
    queryKey: ['staged-visualization-data', JSON.stringify(scoped)],
    queryFn: async () => {
      if (!scoped) return { rows: [], total: 0 }
      const result = await service.run(scoped)
      const rows = resultToRows(result, scoped)
      return { rows, total: result.total ?? rows.length }
    },
    enabled: scoped != null,
    staleTime: 30_000,
    retry: false,
  })
}
