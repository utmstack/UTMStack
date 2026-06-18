import { createApiClient } from '@/shared/lib/api-client'
import type { IndexProperty } from '@/features/dashboard/types'

const BASE_URL = '/opensearch/index/properties'

export interface IndexPropertiesService {
  getIndexProperties(indexPattern: string): Promise<IndexProperty[]>
}

export function createIndexPropertiesService(baseUrl?: string): IndexPropertiesService {
  const api = createApiClient(baseUrl)
  return {
    getIndexProperties: (indexPattern: string) =>
      api.get<IndexProperty[]>(`${BASE_URL}?indexPattern=${encodeURIComponent(indexPattern)}`),
  }
}
