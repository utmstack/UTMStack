import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'
import type {
  Datasource,
  IngestionTimeline,
  IngestionTotals,
  ListResponse,
  DatasourceCount,
  SourceKind,
} from '../types/datasource.types'

const api = createApiClient()

export { ApiError as DatasourcesHttpError }

/** Filter DSL: clauses are `field.op.value` joined by `&` inside search_query.
 * The backend SplitN(.,3) keeps dotted values (IPs) intact. */
function buildSearchQuery(parts: (string | null | undefined)[]): string {
  return parts.filter(Boolean).join('&')
}

export interface ListParams {
  page: number // 1-based
  size: number
  search?: string
  kind?: SourceKind
  label?: string // matches the comma-separated labels column (ILIKE %label%)
  staleBefore?: string // RFC3339 → last_ping_at.lt.<iso> (the "offline" filter)
  pingNull?:boolean,
  sort?: string // e.g. "last_ping_at.desc"
}

export const datasourcesHttpService = {
  list: (p: ListParams) => {
    const q = new URLSearchParams({
      page_number: String(p.page),
      page_size: String(p.size),
      sort_by: p.sort ?? 'last_ping_at.desc',
    })
    const filter = buildSearchQuery([
      p.search ? `asset_name.like.${p.search}` : null,
      p.kind ? `source_kind.eq.${p.kind}` : null,
      p.label ? `labels.like.${p.label}` : null,
      p.staleBefore ? `last_ping_at.lt.${p.staleBefore}` : null,
      p.pingNull ? `last_ping_at.null` : null,
    ])
    if (filter) q.set('search_query', filter)
    return api.get<ListResponse<Datasource>>(`/datasources?${q.toString()}`)
  },
  get: (id: string) => api.get<Datasource>(`/datasources/${id}`),
  remove: (id: string) => api.delete<void>(`/datasources/${id}`),
  updateLabels: (id: string, labels: string) =>
    api.put<void>('/datasources/labels', { id, labels }),
  updateSensitivity: (
    id: string,
    cia: { assetConfidentiality: number; assetIntegrity: number; assetAvailability: number },
  ) => api.put<void>('/datasources/sensitivity', { id, ...cia }),

  count: () => api.get<DatasourceCount>('/datasources/count'),

  // Live ingestion stats from v11-statistics-* (no DB). Keyed by dataSource name.S
  ingestionTotals: () =>
    api.get<IngestionTotals>('/eventprocessing/ingestion-stats?groupBy=dataSource&status=received&top=500'),
  ingestionTimeline: (range?: TimeRange) => {
    const interval_time = range?.interval
    const qs = new URLSearchParams({
      status: 'received',
      interval: interval_time?`1${interval_time[0]}`:'auto',
    })
    if (range?.from) qs.set('from', range.from)
    if (range?.to) qs.set('to', range.to)
    return api.get<IngestionTimeline>(`/eventprocessing/ingestion-stats/timeline?${qs.toString()}`)
  },
  // Per-source ingestion trend (scopes the timeline to one dataSource by name).
  ingestionTimelineFor: (name: string) =>
    api.get<IngestionTimeline>(
      `/eventprocessing/ingestion-stats/timeline?status=received&interval=auto&dataSource=${encodeURIComponent(name)}`,
    ),
}
