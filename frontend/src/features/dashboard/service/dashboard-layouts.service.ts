import { createApiClient, type Paged } from '@/shared/lib/api-client'
import type {
  DashboardLayoutCreateInput,
  DashboardLayoutListParams,
  DashboardLayoutUpdateInput,
  DashboardVisualization,
} from '@/features/dashboard/types'

const BASE_URL = '/dashboard-layouts'

function buildQuery(params: DashboardLayoutListParams): string {
  const p = new URLSearchParams()
  if (params.idDashboard != null) p.set('idDashboard', String(params.idDashboard))
  if (params.idVisualization != null) p.set('idVisualization', String(params.idVisualization))
  if (params.page != null) p.set('page', String(params.page))
  if (params.size != null) p.set('size', String(params.size))
  const q = p.toString()
  return q ? `?${q}` : ''
}

export interface DashboardLayoutsService {
  listLayouts(params?: DashboardLayoutListParams): Promise<Paged<DashboardVisualization[]>>
  getLayout(id: number): Promise<DashboardVisualization>
  createLayout(data: DashboardLayoutCreateInput): Promise<DashboardVisualization>
  updateLayout(data: DashboardLayoutUpdateInput): Promise<DashboardVisualization>
  deleteLayout(id: number): Promise<void>
}

export function createDashboardLayoutsService(baseUrl?: string): DashboardLayoutsService {
  const api = createApiClient(baseUrl)

  return {
    listLayouts: (params = {}) =>
      api.getPaged<DashboardVisualization[]>(`${BASE_URL}${buildQuery(params)}`),

    getLayout: (id: number) => api.get<DashboardVisualization>(`${BASE_URL}/${id}`),

    createLayout: (data: DashboardLayoutCreateInput) =>
      api.post<DashboardVisualization>(BASE_URL, data),

    updateLayout: (data: DashboardLayoutUpdateInput) =>
      api.put<DashboardVisualization>(BASE_URL, data),

    deleteLayout: (id: number) => api.delete<void>(`${BASE_URL}/${id}`),
  }
}
