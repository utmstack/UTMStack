import { useMemo } from 'react'
import { useDatasetFields } from '@/features/dashboard/hooks/useDatasets'
import {
  filterAggregatableFields,
  groupableFields,
} from '@/features/dashboard/utils/aggregatable-fields'
import type { IndexProperty } from '@/features/dashboard/types'

/**
 * Returns the dataset fields offered as a widget GROUP BY / filter.
 *
 * Derived purely from the dataset's own field list (one request) with a static
 * filter ({@link filterAggregatableFields}) that drops the sub-document
 * containers, keeping the paths inside them. We deliberately do NOT probe each
 * candidate with its own query — that fired one request per field (dozens of
 * calls, many failures) every time the editor opened. The chart preview already
 * runs the real question and surfaces any field that does not resolve.
 */
export function useAggregatableFields(dataset: string | null | undefined) {
  const propsQuery = useDatasetFields(dataset)

  const fields = useMemo<IndexProperty[]>(
    () => filterAggregatableFields(propsQuery.data ?? []),
    [propsQuery.data]
  )

  // Subset a chart can break down by. Filters and columns still use `fields`.
  const groupable = useMemo<IndexProperty[]>(() => groupableFields(fields), [fields])

  return {
    fields,
    groupableFields: groupable,
    isLoading: propsQuery.isFetching && !!dataset?.trim(),
  }
}
