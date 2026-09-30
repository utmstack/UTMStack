import { createApiClient, type Paged } from '@/shared/lib/api-client'
import type {
  DashboardFilter,
  DashboardFilterCreateInput,
  DashboardFilterListParams,
  DashboardFilterUpdateInput,
} from '@/features/dashboard/types'

const BASE_URL = '/dashboards/filters'

function buildQuery(params: DashboardFilterListParams): string {
  const p = new URLSearchParams()
  if (params.dashboardId != null) p.set('dashboardId', String(params.dashboardId))
  if (params.page != null) p.set('page', String(params.page))
  if (params.size != null) p.set('size', String(params.size))
  const q = p.toString()
  return q ? `?${q}` : ''
}

export interface DashboardFiltersService {
  listDashboardFilters(params?: DashboardFilterListParams): Promise<Paged<DashboardFilter[]>>
  getDashboardFilter(id: string): Promise<DashboardFilter>
  createDashboardFilter(data: DashboardFilterCreateInput): Promise<DashboardFilter>
  updateDashboardFilter(data: DashboardFilterUpdateInput): Promise<DashboardFilter>
  deleteDashboardFilter(id: string): Promise<void>
}

export function createDashboardFiltersService(baseUrl?: string): DashboardFiltersService {
  const api = createApiClient(baseUrl)

  return {
    listDashboardFilters: (params = {}) => api.getPaged<DashboardFilter[]>(`${BASE_URL}${buildQuery(params)}`),

    getDashboardFilter: (id: string) => api.get<DashboardFilter>(`${BASE_URL}/${id}`),

    createDashboardFilter: (data: DashboardFilterCreateInput) => api.post<DashboardFilter>(BASE_URL, data),

    updateDashboardFilter: (data: DashboardFilterUpdateInput) => api.put<DashboardFilter>(BASE_URL, data),

    deleteDashboardFilter: (id: string) => api.delete<void>(BASE_URL + '/' + id),
  }
}
