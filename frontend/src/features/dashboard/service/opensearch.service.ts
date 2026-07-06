import { createApiClient, type Paged } from '@/shared/lib/api-client'
import type { FilterType } from '@/features/dashboard/types'

const BASE_URL = '/opensearch'

export type Row = Record<string, unknown>

export interface SqlSearchInput {
  query: string
  fetchSize?: number
  page?: number
  size?: number
  filters?: FilterType[]
}

export interface OpenSearchService {
  searchSql(input: SqlSearchInput): Promise<Paged<Row[]>>
}

export function createOpenSearchService(baseUrl?: string): OpenSearchService {
  const api = createApiClient(baseUrl)

  const searchSql: OpenSearchService['searchSql'] = ({
    query,
    fetchSize,
    filters,
    page = 1,
    size = 100,
  }) => {
    const params = new URLSearchParams()
    params.set('page', String(page))
    params.set('size', String(size))
    return api.postPaged<Row[]>(`${BASE_URL}/search/sql?${params.toString()}`, {
      query,
      fetchSize,
      filters,
    })
  }

  return {
    searchSql,
  }
}
