import { createApiClient } from '@/shared/lib/api-client'

const BASE_URL = '/opensearch'

export interface PropertyValuesInput {
  keyword: string
  indexPattern: string
}

export interface PropertyValuesService {
  getValues(input: PropertyValuesInput): Promise<string[]>
}

export function createPropertyValuesService(baseUrl?: string): PropertyValuesService {
  const api = createApiClient(baseUrl)
  return {
    getValues: ({ keyword, indexPattern }) => {
      const params = new URLSearchParams()
      params.set('keyword', keyword)
      params.set('indexPattern', indexPattern)
      return api.get<string[]>(`${BASE_URL}/property/values?${params.toString()}`)
    },
  }
}
