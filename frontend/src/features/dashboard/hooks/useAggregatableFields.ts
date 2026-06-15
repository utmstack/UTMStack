import { useMemo } from 'react'
import { useIndexProperties } from '@/features/dashboard/hooks/useIndexProperties'
import { filterAggregatableFields } from '@/features/dashboard/utils/aggregatable-fields'
import type { IndexProperty } from '@/features/dashboard/types'

/**
 * Returns the index fields offered as a widget GROUP BY / filter.
 *
 * Derived purely from the index mapping (one `properties` request) with a static
 * filter ({@link filterAggregatableFields}) that drops nested/object containers and
 * `.keyword` multifield duplicates. We deliberately do NOT empirically probe each
 * candidate with its own SQL query — that fired one `SELECT … GROUP BY` request per
 * field (dozens of calls, many 400s) every time the editor opened. The chart preview
 * already runs the real SQL and surfaces any field that doesn't resolve, so eager
 * per-field validation was both redundant and a performance problem.
 */
export function useAggregatableFields(indexPattern: string | null | undefined) {
  const propsQuery = useIndexProperties(indexPattern)

  const fields = useMemo<IndexProperty[]>(
    () => filterAggregatableFields(propsQuery.data ?? []),
    [propsQuery.data]
  )

  return {
    fields,
    isLoading: propsQuery.isFetching && !!indexPattern?.trim(),
  }
}
