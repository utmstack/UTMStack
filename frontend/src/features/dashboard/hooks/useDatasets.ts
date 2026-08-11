import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createDatasetService } from '@/features/dashboard/service/dataset.service'
import type { IndexProperty } from '@/features/dashboard/types'

export const DATASET_QUERY_KEYS = {
  all: ['datasets'] as const,
  list: () => [...DATASET_QUERY_KEYS.all, 'list'] as const,
  fields: (dataset: string) => [...DATASET_QUERY_KEYS.all, 'fields', dataset] as const,
  values: (dataset: string, field: string) =>
    [...DATASET_QUERY_KEYS.all, 'values', dataset, field] as const,
}

/** The datasets the store holds — what a widget reads from. */
export function useDatasets() {
  const service = useMemo(() => createDatasetService(), [])
  return useQuery<string[]>({
    queryKey: DATASET_QUERY_KEYS.list(),
    queryFn: () => service.listDatasets(),
    staleTime: 60_000,
  })
}

/** Every field a dataset's records carry. */
export function useDatasetFields(dataset: string | null | undefined) {
  const service = useMemo(() => createDatasetService(), [])
  const enabled = !!dataset?.trim()
  return useQuery<IndexProperty[]>({
    queryKey: DATASET_QUERY_KEYS.fields(dataset ?? ''),
    queryFn: () => service.listFields(dataset as string),
    enabled,
    staleTime: 60_000,
  })
}

/** The values a field actually takes, most frequent first. */
export function useFieldValues(dataset: string, field: string, enabled = true) {
  const service = useMemo(() => createDatasetService(), [])
  return useQuery<string[]>({
    queryKey: DATASET_QUERY_KEYS.values(dataset, field),
    queryFn: () => service.topValues(dataset, field),
    enabled: enabled && !!dataset && !!field,
    staleTime: 60_000,
  })
}
