import { createApiClient, type Paged } from '@/shared/lib/api-client'
import type {
  IndexPattern,
  IndexPatternFieldsResponse,
  IndexPatternListParams,
} from '@/features/dashboard/types'

const BASE_URL = '/opensearch/index-patterns'

function buildQuery(params: IndexPatternListParams): string {
  const p = new URLSearchParams()
  if (params.isActive != null) p.set('isActive.equals', String(params.isActive))
  if (params.page != null) p.set('page', String(params.page))
  if (params.size != null) p.set('size', String(params.size))
  if (params.sort) p.set('sort', params.sort)
  const q = p.toString()
  return q ? `?${q}` : ''
}

export interface IndexPatternsService {
  listIndexPatterns(params?: IndexPatternListParams): Promise<Paged<IndexPattern[]>>
  getIndexPatternFields(patternId: number): Promise<IndexPatternFieldsResponse>
}

export function createIndexPatternsService(baseUrl?: string): IndexPatternsService {
  const api = createApiClient(baseUrl)

  return {
    listIndexPatterns: (params = {}) =>
      api.getPaged<IndexPattern[]>(`${BASE_URL}${buildQuery(params)}`),

    getIndexPatternFields: (patternId: number) =>
      api.get<IndexPatternFieldsResponse>(`${BASE_URL}/fields?id=${patternId}`),
  }
}
