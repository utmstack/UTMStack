import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createIndexPropertiesService } from '@/features/dashboard/service/index-properties.service'
import type { IndexProperty } from '@/features/dashboard/types'

export const INDEX_PROPERTIES_QUERY_KEYS = {
  all: ['index-properties'] as const,
  forPattern: (pattern: string) => [...INDEX_PROPERTIES_QUERY_KEYS.all, pattern] as const,
}

export function useIndexProperties(indexPattern: string | null | undefined) {
  const service = useMemo(() => createIndexPropertiesService(), [])
  const enabled = !!indexPattern?.trim()
  return useQuery<IndexProperty[]>({
    queryKey: INDEX_PROPERTIES_QUERY_KEYS.forPattern(indexPattern ?? ''),
    queryFn: () => service.getIndexProperties(indexPattern as string),
    enabled,
    staleTime: 60_000,
  })
}
