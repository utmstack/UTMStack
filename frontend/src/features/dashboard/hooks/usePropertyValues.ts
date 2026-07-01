import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createPropertyValuesService } from '@/features/dashboard/service/property-values.service'

export const PROPERTY_VALUES_QUERY_KEYS = {
  all: ['property-values'] as const,
  for: (indexPattern: string, keyword: string) =>
    [...PROPERTY_VALUES_QUERY_KEYS.all, indexPattern, keyword] as const,
}

export function usePropertyValues(indexPattern: string, keyword: string, enabled = true) {
  const service = useMemo(() => createPropertyValuesService(), [])
  return useQuery<string[]>({
    queryKey: PROPERTY_VALUES_QUERY_KEYS.for(indexPattern, keyword),
    queryFn: () => service.getValues({ indexPattern, keyword }),
    enabled: enabled && !!indexPattern && !!keyword,
    staleTime: 60_000,
  })
}
