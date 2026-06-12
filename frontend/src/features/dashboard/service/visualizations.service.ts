import { createApiClient, type Paged } from '@/shared/lib/api-client'
import type {
  Visualization,
  VisualizationCreateInput,
  VisualizationListParams,
  VisualizationUpdateInput,
} from '@/features/dashboard/types'

const BASE_URL = '/visualizations'

function buildQuery(params: VisualizationListParams): string {
  const p = new URLSearchParams()
  if (params.name) p.set('name', params.name)
  if (params.page != null) p.set('page', String(params.page))
  if (params.size != null) p.set('size', String(params.size))
  const q = p.toString()
  return q ? `?${q}` : ''
}

export interface VisualizationsService {
  listVisualizations(params?: VisualizationListParams): Promise<Paged<Visualization[]>>
  getVisualization(id: number): Promise<Visualization>
  createVisualization(data: VisualizationCreateInput): Promise<Visualization>
  updateVisualization(data: VisualizationUpdateInput): Promise<Visualization>
  deleteVisualization(id: number): Promise<void>
}

export function createVisualizationsService(baseUrl?: string): VisualizationsService {
  const api = createApiClient(baseUrl)

  return {
    listVisualizations: (params = {}) =>
      api.getPaged<Visualization[]>(`${BASE_URL}${buildQuery(params)}`),

    getVisualization: (id: number) => api.get<Visualization>(`${BASE_URL}/${id}`),

    createVisualization: (data: VisualizationCreateInput) =>
      api.post<Visualization>(BASE_URL, data),

    updateVisualization: (data: VisualizationUpdateInput) =>
      api.put<Visualization>(BASE_URL, data),

    deleteVisualization: (id: number) => api.delete<void>(`${BASE_URL}/${id}`),
  }
}
