import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ChartView,
  FilterType,
  IndexField,
  LogDocument,
  TopValues,
} from '../types/log-explorer.types'

const api = createApiClient()

export { ApiError as LogExplorerHttpError }

/** Top results cap the backend search honours (matches the legacy MAX_SEARCH_RESULTS). */
export const MAX_SEARCH_RESULTS = 10_000

/** The explorer explores logs; alerts have their own page. */
const LOGS_DATASET = 'logs'

export const logExplorerHttpService = {
  // The explorer always reads logs; what an analyst picks between is the kind
  // of log. Read from the data, so a type that stopped arriving stops being
  // offered — which is what the index-pattern registry used to be for.
  dataTypes: () => api.get<string[]>(`/log-analyzer/datasets/${LOGS_DATASET}/data-types`),

  // Queryable fields of a dataset, straight from the store: a field that exists
  // is offered and one that does not is not.
  fields: () => api.get<IndexField[]>(`/log-analyzer/datasets/${LOGS_DATASET}/fields`),

  // Main document search, through the store rather than the index gateway.
  // `page` is 0-based here, matching the endpoint.
  search: async (params: {
    dataType: string | null
    filters: FilterType[]
    page: number
    size: number
    from?: string
    to?: string
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
  }): Promise<{ data: LogDocument[]; total: number }> => {
    const res = await api.post<{ data: LogDocument[]; total: number }>('/log-analyzer/search', {
      dataset: LOGS_DATASET,
      dataType: params.dataType ?? undefined,
      filters: params.filters,
      page: params.page,
      size: params.size,
      from: params.from,
      to: params.to,
      sortBy: params.sortBy ?? '@timestamp',
      order: params.sortOrder ?? 'desc',
    })
    return { data: res.data ?? [], total: res.total ?? 0 }
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
