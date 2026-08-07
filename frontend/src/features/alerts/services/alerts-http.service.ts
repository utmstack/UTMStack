import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { Alert, AlertTag, FilterType, RelatedLogsResponse } from '../types/alert.types'

const api = createApiClient()

export { ApiError as AlertsHttpError }

// The alerts dataset of the event store. It used to be an OpenSearch index
// pattern; the store has named datasets and knows its own columns.
const ALERT_DATASET = 'alerts'
const CSV_MAX = 10_000

// The search endpoint caps a page at 500, so the export walks pages rather than
// asking for everything at once — a single oversized request comes back
// silently truncated, which is the one thing an export must not do.
const SEARCH_PAGE_MAX = 500

interface TopValues {
  total: number
  top: { value: string; count: number; percent: number }[]
}
interface ChartView {
  categories: string[]
  values: number[]
}

export interface AlertListParams {
  page: number // 0-based, like the search endpoint
  size: number
  filters: FilterType[]
}

export const alertsHttpService = {
  // Flat, paginated alert list, newest first.
  list: async ({ page, size, filters }: AlertListParams) => {
    const r = await api.post<{ data: Alert[]; total: number }>('/log-analyzer/search', {
      dataset: ALERT_DATASET,
      filters,
      page,
      size,
      sortBy: '@timestamp',
      order: 'desc',
    })
    return { data: r.data ?? [], total: r.total ?? 0 }
  },

  // Counts by a field (severity / status) honouring the current filters.
  counts: (field: 'severity' | 'status', filters: FilterType[]) =>
    api.post<TopValues>(
      `/log-analyzer/top-x-values/${ALERT_DATASET}/${encodeURIComponent(field)}/10`,
      filters
    ),

  // Existing distinct values of an arbitrary field — to populate the filter
  // value picker (so the user selects real values, not free text). Text fields
  // need their .keyword sub-field for the terms aggregation; we fall back to it.
  fieldValues: async (field: string, top = 100): Promise<{ value: string; count: number }[]> => {
    // An alert with no parent has parentId '', not a missing parentId: the column
    // is a plain String, so "has no parent" is an equality test and an existence
    // test matches nothing.
    const body: FilterType[] = [{ field: 'parentId', operator: 'IS', value: '' }]
    const url = (f: string) =>
      `/log-analyzer/top-x-values/${ALERT_DATASET}/${encodeURIComponent(f)}/${top}`
    const tryFetch = async (f: string) => {
      const r = await api.post<TopValues>(url(f), body)
      return r.top ?? []
    }
    try {
      const r = await tryFetch(field)
      if (r.length || field.endsWith('.keyword')) return r
      return await tryFetch(`${field}.keyword`)
    } catch {
      if (field.endsWith('.keyword')) return []
      try {
        return await tryFetch(`${field}.keyword`)
      } catch {
        return []
      }
    }
  },

  // Alerts-over-time histogram.
  timeline: (filters: FilterType[], interval: string) =>
    api.post<ChartView>('/log-analyzer/chart-view', {
      dataset: ALERT_DATASET,
      field: '@timestamp',
      fieldDataType: 'date',
      filters,
      interval,
      top: 60,
    }),

  // Re-fetch a single alert (e.g. to refresh its history after an action).
  getById: async (id: string): Promise<Alert | null> => {
    const r = await api.post<{ data: Alert[] }>('/log-analyzer/search', {
      dataset: ALERT_DATASET,
      filters: [{ field: 'id', operator: 'IS', value: id }],
      // The endpoint counts pages from 0. Asking for page 1 with size 1 skipped
      // the only row there was, so this always answered "no such alert".
      page: 0,
      size: 1,
    })
    return r.data?.[0] ?? null
  },

  countOpen: () => api.get<number>('/utm-alerts/count-open-alerts'),

  // Paginated child "echoes" of a parent alert (the dedup children the main
  // list hides). Backend sorts newest-first by default.
  echoes: (parentId: string, page: number, size: number) =>
    api.getPaged<Alert[]>(
      `/utm-alerts/${encodeURIComponent(parentId)}/echoes?page=${page}&size=${size}`,
    ),

  // All logs the Event Processor correlated for this alert (reproduced server-side
  // without the engine's 10-hit cap) — for the "view all related logs" deep-link.
  relatedLogs: (alertId: string) =>
    api.get<RelatedLogsResponse>(`/utm-alerts/related-logs?alertId=${encodeURIComponent(alertId)}`),

  // Tag catalog (for the tag picker).
  tags: () => api.get<AlertTag[]>('/utm-alert-tags?page=1&size=500'),

  // Register a brand-new tag in the catalog (name + color).
  createTag: (tagName: string, tagColor: string) =>
    api.post<AlertTag>('/utm-alert-tags', { tagName, tagColor }),

  // Rename / recolor an existing catalog tag. The backend rejects system-owned
  // tags (e.g. "False positive") with 403.
  updateTag: (id: string, tagName: string, tagColor: string) =>
    api.put<AlertTag>('/utm-alert-tags', { id, tagName, tagColor }),

  // Remove a catalog tag. Backend rejects system-owned tags with 403.
  deleteTag: (id: string) => api.delete<void>(`/utm-alert-tags/${id}`),

  // CSV export of the current alert list (honours the active scope filters).
  // Mirrors the Log Explorer downloader: hits /opensearch/search/csv and triggers
  // a browser download.
  // CSV of the current alert list. Built here rather than server-side: the
  // event store has no CSV endpoint, and the export is the page's own view of
  // the data — the same rows, the same columns, in the order shown.
  exportCsv: async (filters: FilterType[]) => {
    const columns: { label: string; field: keyof Alert | 'tags' }[] = [
      { label: 'Timestamp', field: '@timestamp' },
      { label: 'Name', field: 'name' },
      { label: 'Severity', field: 'severity' },
      { label: 'Status', field: 'status' },
      { label: 'Category', field: 'category' },
      { label: 'Technique', field: 'technique' },
      { label: 'Source', field: 'dataSource' },
      { label: 'Tags', field: 'tags' },
      { label: 'Notes', field: 'notes' },
    ]

    const rows: Alert[] = []
    // From 0: the endpoint's first page. Starting at 1 dropped the newest
    // SEARCH_PAGE_MAX alerts from every export.
    for (let page = 0; rows.length < CSV_MAX; page++) {
      const r = await api.post<{ data: Alert[] }>('/log-analyzer/search', {
        dataset: ALERT_DATASET,
        filters,
        page,
        size: Math.min(SEARCH_PAGE_MAX, CSV_MAX - rows.length),
        sortBy: '@timestamp',
        order: 'desc',
      })
      const batch = r.data ?? []
      rows.push(...batch)
      if (batch.length < SEARCH_PAGE_MAX) break
    }

    const cell = (v: unknown) => {
      const t = Array.isArray(v) ? v.join(', ') : v == null ? '' : String(v)
      return /[",\n]/.test(t) ? `"${t.replace(/"/g, '""')}"` : t
    }
    const csv = [
      columns.map((c) => c.label).join(','),
      ...rows.map((a) => columns.map((c) => cell(a[c.field as keyof Alert])).join(',')),
    ].join('\n')

    const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
    const a = document.createElement('a')
    a.href = url
    a.download = `alerts-${new Date().toISOString().slice(0, 19)}.csv`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  },

  // Actions (operate on one or many alert ids).
  updateStatus: (alertIds: string[], status: string, statusObservation = '', addFalsePositiveTag = false) =>
    api.post<void>('/utm-alerts/status', { alertIds, status, statusObservation, addFalsePositiveTag }),

  updateTags: (alertIds: string[], tags: string[], createRule = false) =>
    api.post<void>('/utm-alerts/tags', { alertIds, tags, createRule }),

  // Notes is a raw-string body keyed by alertId query param.
  updateNotes: (alertId: string, notes: string) =>
    api.post<void>(`/utm-alerts/notes?alertId=${encodeURIComponent(alertId)}`, notes, {
      headers: { 'Content-Type': 'text/plain' },
    }),

  // Assign / unassign an alert (empty assignee clears it).
  updateAssignee: (alertId: string, assignee: string) =>
    api.post<void>('/utm-alerts/assignee', { alertId, assignee }),

  // Flag alert docs as belonging to an incident (after the incident exists).
  convertToIncident: (eventIds: string[], incidentName: string, incidentId: string) =>
    api.post<void>('/utm-alerts/convert-to-incident', {
      eventIds,
      incidentName,
      incidentId,
      incidentSource: 'incidents',
    }),
}
