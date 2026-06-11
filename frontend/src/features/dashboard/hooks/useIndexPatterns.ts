import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { createIndexPatternsService } from '@/features/dashboard/service/index-patterns.service'
import type {
  IndexPattern,
  IndexPatternFieldsResponse,
  IndexPatternListParams,
} from '@/features/dashboard/types'
import type { Paged } from '@/shared/lib/api-client'

export const INDEX_PATTERNS_QUERY_KEYS = {
  all: ['index-patterns'] as const,
  list: (params: IndexPatternListParams) =>
    [...INDEX_PATTERNS_QUERY_KEYS.all, 'list', params] as const,
  fields: (patternId: number) =>
    [...INDEX_PATTERNS_QUERY_KEYS.all, 'fields', patternId] as const,
}

const DEFAULT_PARAMS: IndexPatternListParams = {
  isActive: true,
  page: 0,
  size: 200,
  sort: 'pattern,asc',
}

export function useIndexPatterns(params: IndexPatternListParams = DEFAULT_PARAMS) {
  const service = useMemo(() => createIndexPatternsService(), [])
  return useQuery<Paged<IndexPattern[]>>({
    queryKey: INDEX_PATTERNS_QUERY_KEYS.list(params),
    queryFn: () => service.listIndexPatterns(params),
    staleTime: 60_000,
  })
}

export function useIndexPatternFields(patternId: number | null) {
  const service = useMemo(() => createIndexPatternsService(), [])
  return useQuery<IndexPatternFieldsResponse>({
    queryKey: INDEX_PATTERNS_QUERY_KEYS.fields(patternId ?? 0),
    queryFn: () => service.getIndexPatternFields(patternId as number),
    enabled: patternId != null && patternId > 0,
    staleTime: 60_000,
  })
}
