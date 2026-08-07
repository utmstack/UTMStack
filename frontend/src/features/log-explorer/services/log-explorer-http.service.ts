import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ChartView,
  FilterType,
  IndexField,
  LogDocument,
  TopValues,
} from '../types/log-explorer.types'
import { flattenDoc } from '../domain/flatten'

const api = createApiClient()

export { ApiError as LogExplorerHttpError }

/** How many rows an export gathers at most. */
export const MAX_SEARCH_RESULTS = 10_000

// The search endpoint caps a page at 500, so anything wanting more walks pages.
// Asking for more in one call comes back silently truncated.
const SEARCH_PAGE_MAX = 500

/**
 * A column of the export. It carries a resolver rather than a field name
 * because that is what the table's columns are: "source" is whichever of a
 * handful of candidates the row has, and "message" falls back to a summary
 * built from the record when no message field exists — which, for normalized
 * logs, is most of them. A CSV that only reads field names comes out blank
 * wherever the screen is at its most useful.
 */
export interface ExportColumn {
  label: string
  value: (flat: Record<string, unknown>) => string
}

/** Every field the rows carry, first seen first — the header for a statement,
 *  whose columns are whatever it selected. */
function columnsOf(rows: LogDocument[]): ExportColumn[] {
  const seen = new Set<string>()
  for (const r of rows) for (const k of Object.keys(flattenDoc(r))) seen.add(k)
  return [...seen].map((f) => ({
    label: f,
    value: (flat: Record<string, unknown>) => (flat[f] == null ? '' : String(flat[f])),
  }))
}

function toCsv(rows: LogDocument[], columns: ExportColumn[]): string {
  const cell = (v: string) => (/[",\n]/.test(v) ? `"${v.replace(/"/g, '""')}"` : v)
  return [
    columns.map((c) => cell(c.label)).join(','),
    ...rows.map((doc) => {
      const flat = flattenDoc(doc)
      return columns.map((c) => cell(c.value(flat))).join(',')
    }),
  ].join('\n')
}

function download(csv: string, name: string) {
  // Excel reads a bare UTF-8 CSV as the local codepage and mangles anything
  // non-ASCII; the BOM is what tells it otherwise.
  const blob = new Blob(['﻿', csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${name}-${new Date().toISOString().slice(0, 19)}.csv`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

/** The explorer reads whatever the store holds; the dataset comes from the
 *  caller, because "which table" and "which kind of record inside it" are two
 *  questions and the selector asks both. */
export const DEFAULT_DATASET = 'logs'

export const logExplorerHttpService = {
  // The explorer always reads logs; what an analyst picks between is the kind
  // of log. Read from the data, so a type that stopped arriving stops being
  // offered — which is what the index-pattern registry used to be for.
  datasets: () => api.get<string[]>('/log-analyzer/datasets'),

  dataTypes: (dataset: string) =>
    api.get<string[]>(`/log-analyzer/datasets/${encodeURIComponent(dataset)}/data-types`),

  // Queryable fields of a dataset, straight from the store: a field that exists
  // is offered and one that does not is not.
  fields: (dataset: string) =>
    api.get<IndexField[]>(`/log-analyzer/datasets/${encodeURIComponent(dataset)}/fields`),

  // Main document search, through the store rather than the index gateway.
  // `page` is 0-based here, matching the endpoint.
  search: async (params: {
    dataset: string
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
      dataset: params.dataset,
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
  // SQL mode. The datasets are named `logs` and `alerts`; both resolve to the
  // caller's own tenant, and naming a real table or another database is
  // refused by the backend rather than scoped.
  searchSql: async (query: string, page: number, size: number) => {
    const q = new URLSearchParams({ page: String(page), size: String(size) })
    const r = await api.post<{ data: LogDocument[]; total: number }>(
      `/log-analyzer/search-sql?${q.toString()}`,
      { query }
    )
    return { data: r.data ?? [], total: r.total ?? 0 }
  },

  // Events-over-time histogram (date_histogram). interval is a word: Day/Hour/Minute/...
  chartView: (body: {
    dataset: string
    dataType: string | null
    field: string
    fieldDataType: string
    filters: FilterType[]
    interval: string
    top: number
  }) =>
    api.post<ChartView>('/log-analyzer/chart-view', {
      ...body,
      dataType: body.dataType ?? undefined,
    }),

  // Top values for a field (sidebar). `top` defaults to 5 like the legacy.
  // The path names the dataset, not the kind of log inside it: that is a query
  // parameter. Passing the data type where the dataset goes asks the store for
  // a dataset it has never heard of.
  topValues: (dataset: string, dataType: string | null, field: string, filters: FilterType[], top = 5) => {
    const q = dataType ? `?dataType=${encodeURIComponent(dataType)}` : ''
    return api.post<TopValues>(
      `/log-analyzer/top-x-values/${encodeURIComponent(dataset)}/${encodeURIComponent(field)}/${top}${q}`,
      filters
    )
  },

  // CSV of the current query. Built here rather than server-side: the event
  // store has no CSV endpoint, and the export is the page's own view of the
  // data — the same rows, the same columns, in the order shown.
  exportCsv: async (body: {
    dataset: string
    dataType: string | null
    filters: FilterType[]
    columns: ExportColumn[]
    from?: string
    to?: string
  }) => {
    const rows: LogDocument[] = []
    for (let page = 0; rows.length < MAX_SEARCH_RESULTS; page++) {
      const r = await api.post<{ data: LogDocument[] }>('/log-analyzer/search', {
        dataset: body.dataset,
        dataType: body.dataType ?? undefined,
        filters: body.filters,
        page,
        size: SEARCH_PAGE_MAX,
        from: body.from,
        to: body.to,
        sortBy: '@timestamp',
        order: 'desc',
      })
      const batch = r.data ?? []
      rows.push(...batch)
      if (batch.length < SEARCH_PAGE_MAX) break
    }
    download(toCsv(rows.slice(0, MAX_SEARCH_RESULTS), body.columns), body.dataset)
  },

  // A statement has its own columns — whatever it selected — so the export runs
  // it rather than the filters, and takes its header from the rows.
  exportSqlCsv: async (query: string) => {
    const rows: LogDocument[] = []
    for (let page = 0; rows.length < MAX_SEARCH_RESULTS; page++) {
      const q = new URLSearchParams({ page: String(page), size: String(SEARCH_PAGE_MAX) })
      const r = await api.post<{ data: LogDocument[] }>(`/log-analyzer/search-sql?${q.toString()}`, {
        query,
      })
      const batch = r.data ?? []
      rows.push(...batch)
      if (batch.length < SEARCH_PAGE_MAX) break
    }
    const capped = rows.slice(0, MAX_SEARCH_RESULTS)
    download(toCsv(capped, columnsOf(capped)), 'query')
  },
}
