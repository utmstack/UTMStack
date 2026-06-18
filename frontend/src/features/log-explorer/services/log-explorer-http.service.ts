import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ChartView,
  FilterType,
  IndexField,
  IndexPattern,
  LogDocument,
  TopValues,
} from '../types/log-explorer.types'

const api = createApiClient()

export { ApiError as LogExplorerHttpError }

/** Top results cap the backend search honours (matches the legacy MAX_SEARCH_RESULTS). */
export const MAX_SEARCH_RESULTS = 10_000

export const logExplorerHttpService = {
  // Active index patterns for the selector (few — fetch all).
  patterns: async (): Promise<IndexPattern[]> => {
    const { data } = await api.getPaged<IndexPattern[]>(
      '/opensearch/index-patterns?isActive.equals=true&page=0&size=200&sort=pattern,asc'
    )
    return data ?? []
  },

  // Field list (name + type) for the selected pattern.
  fields: (indexPattern: string) =>
    api.get<IndexField[]>(
      `/opensearch/index/properties?indexPattern=${encodeURIComponent(indexPattern)}`
    ),

  // Main document search. Returns { data: _source[], total } from X-Total-Count.
  search: (params: {
    indexPattern: string
    filters: FilterType[]
    page: number // 1-based
    size: number
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
  }) => {
    const q = new URLSearchParams({
      indexPattern: params.indexPattern,
      page: String(params.page),
      size: String(params.size),
      top: String(MAX_SEARCH_RESULTS),
      sortBy: params.sortBy ?? '@timestamp',
      sortOrder: params.sortOrder ?? 'desc',
    })
    return api.postPaged<LogDocument[]>(`/opensearch/search?${q.toString()}`, params.filters)
  },

  // SQL mode.
  searchSql: (query: string, page: number, size: number) => {
    const q = new URLSearchParams({ page: String(page), size: String(size) })
    return api.postPaged<LogDocument[]>(`/opensearch/search/sql?${q.toString()}`, { query })
  },

  // Events-over-time histogram (date_histogram). interval is a word: Day/Hour/Minute/...
  chartView: (body: {
    indexPattern: string
    field: string
    fieldDataType: string
    filters: FilterType[]
    interval: string
    top: number
  }) => api.post<ChartView>('/log-analyzer/chart-view', body),

  // Top values for a field (sidebar). `top` defaults to 5 like the legacy.
  topValues: (indexPattern: string, field: string, filters: FilterType[], top = 5) =>
    api.post<TopValues>(
      `/log-analyzer/top-x-values/${encodeURIComponent(indexPattern)}/${encodeURIComponent(
        field
      )}/${top}?sort=${encodeURIComponent('@timestamp,desc')}`,
      filters
    ),

  // CSV export of the current query → triggers a browser download.
  exportCsv: async (body: {
    indexPattern: string
    filters: FilterType[]
    columns: { label: string; field: string; type: string; visible: boolean }[]
  }) => {
    const blob = await api.post<Blob>(
      '/opensearch/search/csv',
      { ...body, top: MAX_SEARCH_RESULTS },
      { responseType: 'blob' }
    )
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `logs-${new Date().toISOString().slice(0, 19)}.csv`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  },
}
