import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { Alert, AlertTag, FilterType, RelatedLogsResponse } from '../types/alert.types'

const api = createApiClient()

export { ApiError as AlertsHttpError }

const ALERT_INDEX = 'v11-alert-*'
const MAX = 10_000

interface TopValues {
  total: number
  top: { value: string; count: number; percent: number }[]
}
interface ChartView {
  categories: string[]
  values: number[]
}

export interface AlertListParams {
  page: number // 1-based
  size: number
  filters: FilterType[]
}

export const alertsHttpService = {
  // Flat, paginated alert list straight from OpenSearch (v11-alert-*), newest first.
  list: ({ page, size, filters }: AlertListParams) => {
    const q = new URLSearchParams({
      indexPattern: ALERT_INDEX,
      page: String(page),
      size: String(size),
      top: String(MAX),
      sortBy: '@timestamp',
      sortOrder: 'desc',
    })
    return api.postPaged<Alert[]>(`/opensearch/search?${q.toString()}`, filters)
  },

  // Counts by a field (severity / status) honouring the current filters.
  counts: (field: 'severity' | 'status', filters: FilterType[]) =>
    api.post<TopValues>(
      `/log-analyzer/top-x-values/${encodeURIComponent(ALERT_INDEX)}/${field}/10?sort=${encodeURIComponent('@timestamp,desc')}`,
      filters
    ),

  // Existing distinct values of an arbitrary field — to populate the filter
  // value picker (so the user selects real values, not free text). Text fields
  // need their .keyword sub-field for the terms aggregation; we fall back to it.
  fieldValues: async (field: string, top = 100): Promise<{ value: string; count: number }[]> => {
    const body: FilterType[] = [{ field: 'parentId', operator: 'DOES_NOT_EXIST' }]
    const url = (f: string) =>
      `/log-analyzer/top-x-values/${encodeURIComponent(ALERT_INDEX)}/${encodeURIComponent(f)}/${top}?sort=${encodeURIComponent('@timestamp,desc')}`
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
      indexPattern: ALERT_INDEX,
      field: '@timestamp',
      fieldDataType: 'date',
      filters,
      interval,
      top: 60,
    }),

  // Re-fetch a single alert (e.g. to refresh its history after an action).
  getById: async (id: string): Promise<Alert | null> => {
    const q = new URLSearchParams({ indexPattern: ALERT_INDEX, page: '1', size: '1', top: '1' })
    const { data } = await api.postPaged<Alert[]>(`/opensearch/search?${q.toString()}`, [
      { field: 'id', operator: 'IS', value: id },
    ])
    return data?.[0] ?? null
  },

  countOpen: () => api.get<number>('/utm-alerts/count-open-alerts'),

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
  updateTag: (id: number, tagName: string, tagColor: string) =>
    api.put<AlertTag>('/utm-alert-tags', { id, tagName, tagColor }),

  // Remove a catalog tag. Backend rejects system-owned tags with 403.
  deleteTag: (id: number) => api.delete<void>(`/utm-alert-tags/${id}`),

  // CSV export of the current alert list (honours the active scope filters).
  // Mirrors the Log Explorer downloader: hits /opensearch/search/csv and triggers
  // a browser download.
  exportCsv: async (filters: FilterType[]) => {
    const columns = [
      { label: 'Timestamp', field: '@timestamp', type: 'date', visible: true },
      { label: 'Name', field: 'name', type: 'text', visible: true },
      { label: 'Severity', field: 'severity', type: 'number', visible: true },
      { label: 'Status', field: 'status', type: 'number', visible: true },
      { label: 'Category', field: 'category', type: 'text', visible: true },
      { label: 'Technique', field: 'technique', type: 'text', visible: true },
      { label: 'Source', field: 'dataSource', type: 'text', visible: true },
      { label: 'Tags', field: 'tags', type: 'text', visible: true },
      { label: 'Notes', field: 'notes', type: 'text', visible: true },
    ]
    const blob = await api.post<Blob>(
      '/opensearch/search/csv',
      { indexPattern: ALERT_INDEX, filters, top: MAX, columns },
      { responseType: 'blob' }
    )
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `alerts-${new Date().toISOString().slice(0, 19)}.csv`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  },

  // Actions (operate on one or many alert ids).
  updateStatus: (alertIds: string[], status: number, statusObservation = '', addFalsePositiveTag = false) =>
    api.post<void>('/utm-alerts/status', { alertIds, status, statusObservation, addFalsePositiveTag }),

  updateTags: (alertIds: string[], tags: string[], createRule = false) =>
    api.post<void>('/utm-alerts/tags', { alertIds, tags, createRule }),

  // Notes is a raw-string body keyed by alertId query param.
  updateNotes: (alertId: string, notes: string) =>
    api.post<void>(`/utm-alerts/notes?alertId=${encodeURIComponent(alertId)}`, notes, {
      headers: { 'Content-Type': 'text/plain' },
    }),

  // Flag alert docs as belonging to an incident (after the incident exists).
  convertToIncident: (eventIds: string[], incidentName: string, incidentId: number) =>
    api.post<void>('/utm-alerts/convert-to-incident', {
      eventIds,
      incidentName,
      incidentId,
      incidentSource: 'incidents',
    }),
}
