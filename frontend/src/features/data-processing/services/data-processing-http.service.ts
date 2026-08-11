import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  DataTypeOption,
  Pipeline,
  PipelineListQuery,
  IngestionQuery,
  IngestionStats,
  IngestionTimeline,
  SavePipelineRequest,
} from '../types/data-processing.types'

const api = createApiClient()

export { ApiError as DataProcessingHttpError }

function pipelineQuery(q: PipelineListQuery): string {
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

export const pipelinesHttpService = {
  // Returns { data: Pipeline[], total } — total comes from X-Total-Count.
  list: (q: PipelineListQuery = {}) => api.getPaged<Pipeline[]>(`/eventprocessing/pipelines${pipelineQuery(q)}`),
  dataTypes: () => api.get<string[]>('/eventprocessing/pipelines/data-types'),
  // Catalog of known dataTypes (from integrations) to pick from while authoring.
  dataTypeCatalog: () => api.get<DataTypeOption[]>('/integrations/data-types'),
  find: (relPath: string) =>
    api.get<Pipeline>(`/eventprocessing/pipelines/find?relPath=${encodeURIComponent(relPath)}`),
  create: (input: SavePipelineRequest) => api.post<Pipeline>('/eventprocessing/pipelines', input),
  update: (input: SavePipelineRequest) => api.put<Pipeline>('/eventprocessing/pipelines', input),
  remove: (relPath: string) =>
    api.delete<{ message: string }>(`/eventprocessing/pipelines?relPath=${encodeURIComponent(relPath)}`),
  activate: (relPath: string, active: boolean) =>
    api.put<{ message: string }>(
      `/eventprocessing/pipelines/activate?relPath=${encodeURIComponent(relPath)}&active=${active}`,
    ),
  // The whole sequence, not one position: the backend stores it as an ordered
  // list of pipeline names in the tenant's own config, so a partial update
  // could not be resolved.
  setOrder: (order: string[]) => api.put<void>('/eventprocessing/pipelines/order', { order }),
}

export const ingestionHttpService = {
  stats: (q: IngestionQuery) => api.get<IngestionStats>(`/eventprocessing/ingestion-stats${ingestionQuery(q)}`),
  timeline: (q: IngestionQuery) =>
    api.get<IngestionTimeline>(`/eventprocessing/ingestion-stats/timeline${ingestionQuery(q)}`),
}
