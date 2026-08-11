import { createApiClient, type Paged } from '@/shared/lib/api-client'
import type {
  Dashboard,
  DashboardCreateInput,
  DashboardListParams,
  DashboardUpdateInput,
} from '@/features/dashboard/types'

const BASE_URL = '/dashboards'

function buildQuery(params: DashboardListParams): string {
  const p = new URLSearchParams()
  if (params.name) p.set('name', params.name)
  if (params.page != null) p.set('page', String(params.page))
  if (params.size != null) p.set('size', String(params.size))
  const q = p.toString()
  return q ? `?${q}` : ''
}

export interface DashboardsService {
  listDashboards(params?: DashboardListParams): Promise<Paged<Dashboard[]>>
  getDashboard(id: string): Promise<Dashboard>
  createDashboard(data: DashboardCreateInput): Promise<Dashboard>
  updateDashboard(data: DashboardUpdateInput): Promise<Dashboard>
  deleteDashboard(id: string): Promise<void>
}

export function createDashboardsService(baseUrl?: string): DashboardsService {
  const api = createApiClient(baseUrl)

  return {
    listDashboards: (params = {}) =>
      api.getPaged<Dashboard[]>(`${BASE_URL}${buildQuery(params)}`),

    getDashboard: (id: string) => api.get<Dashboard>(`${BASE_URL}/${id}`),

    createDashboard: (data: DashboardCreateInput) =>
      api.post<Dashboard>(BASE_URL, data),

    updateDashboard: (data: DashboardUpdateInput) =>
      api.put<Dashboard>(BASE_URL, data),

    deleteDashboard: (id: string) => api.delete<void>(`${BASE_URL}/${id}`),
  }
}
