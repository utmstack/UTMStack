import { createApiClient } from '@/shared/lib/api-client'
import type { IndexProperty } from '@/features/dashboard/types'

const BASE_URL = '/log-analyzer'

interface TopValuesResponse {
  total: number
  top: { value: string; count: number; percent: number }[]
}

/**
 * What a widget can be built from, read from the event store itself: the
 * datasets it holds, the fields those records carry, and the values a field
 * actually takes. This replaced the index-pattern registry — a table someone
 * had to keep in step with the data by hand.
 */
export interface DatasetService {
  listDatasets(): Promise<string[]>
  listFields(dataset: string): Promise<IndexProperty[]>
  topValues(dataset: string, field: string, top?: number): Promise<string[]>
}

export function createDatasetService(baseUrl?: string): DatasetService {
  const api = createApiClient(baseUrl)

  return {
    listDatasets: () => api.get<string[]>(`${BASE_URL}/datasets`),

    listFields: (dataset) =>
      api.get<IndexProperty[]>(`${BASE_URL}/datasets/${encodeURIComponent(dataset)}/fields`),

    topValues: async (dataset, field, top = 50) => {
      const res = await api.post<TopValuesResponse>(
        `${BASE_URL}/top-x-values/${encodeURIComponent(dataset)}/${encodeURIComponent(field)}/${top}`,
        []
      )
      return (res.top ?? []).map((v) => v.value)
    },
  }
}
