import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  DataTypeOption,
  Filter,
  FilterListQuery,
  IngestionQuery,
  IngestionStats,
  IngestionTimeline,
  SaveFilterRequest,
} from '../types/data-processing.types'

const api = createApiClient()

export { ApiError as DataProcessingHttpError }

function filterQuery(q: FilterListQuery): string {
  const p = new URLSearchParams()
  if (q.relPathContains) p.set('relPath.contains', q.relPathContains)
  if (q.isActive != null) p.set('isActive.equals', String(q.isActive))
  if (q.system != null) p.set('system.equals', String(q.system))
  if (q.dataType) p.set('dataType.equals', q.dataType)
  p.set('page', String(q.page ?? 1))
  p.set('size', String(q.size ?? 50))
  return `?${p.toString()}`
}

function ingestionQuery(q: IngestionQuery): string {
  const p = new URLSearchParams()
  if (q.groupBy) p.set('groupBy', q.groupBy)
  if (q.status) p.set('status', q.status)
  if (q.from) p.set('from', q.from)
  if (q.to) p.set('to', q.to)
  if (q.interval) p.set('interval', q.interval)
  if (q.top != null) p.set('top', String(q.top))
  const s = p.toString()
  return s ? `?${s}` : ''
}

export const filtersHttpService = {
  // Returns { data: Filter[], total } — total comes from X-Total-Count.
  list: (q: FilterListQuery = {}) => api.getPaged<Filter[]>(`/eventprocessing/filters${filterQuery(q)}`),
  dataTypes: () => api.get<string[]>('/eventprocessing/filters/data-types'),
  // Catalog of known dataTypes (from integrations) to pick from while authoring.
  dataTypeCatalog: () => api.get<DataTypeOption[]>('/integrations/data-types'),
  find: (relPath: string) =>
    api.get<Filter>(`/eventprocessing/filters/find?relPath=${encodeURIComponent(relPath)}`),
  create: (input: SaveFilterRequest) => api.post<Filter>('/eventprocessing/filters', input),
  update: (input: SaveFilterRequest) => api.put<Filter>('/eventprocessing/filters', input),
  remove: (relPath: string) =>
    api.delete<{ message: string }>(`/eventprocessing/filters?relPath=${encodeURIComponent(relPath)}`),
  activate: (relPath: string, active: boolean) =>
    api.put<{ message: string }>(
      `/eventprocessing/filters/activate?relPath=${encodeURIComponent(relPath)}&active=${active}`,
    ),
  setOrder: (relPath: string, order: number) =>
    api.put<Filter>('/eventprocessing/filters/order', { relPath, order }),
}

export const ingestionHttpService = {
  stats: (q: IngestionQuery) => api.get<IngestionStats>(`/eventprocessing/ingestion-stats${ingestionQuery(q)}`),
  timeline: (q: IngestionQuery) =>
    api.get<IngestionTimeline>(`/eventprocessing/ingestion-stats/timeline${ingestionQuery(q)}`),
}
