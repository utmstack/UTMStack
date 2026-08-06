import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

/**
 * A saved explorer query, as the backend stores it (modules/loganalyzer).
 *
 * `filters` is free text there, so it carries the whole query state as JSON:
 * the filters proper plus the time range and the search box, which are just as
 * much part of "the search I had on screen".
 */
export interface SavedQuery {
  id: number
  name: string
  description: string
  owner: string
  createdAt: string
  updatedAt: string
  columns: string
  filters: string
  dataset: string
}

export interface SavedQueryInput {
  name: string
  description?: string
  columns?: string
  filters: string
  dataset: string
}

export const savedQueriesHttpService = {
  // page is 0-based here, and 200 is the largest size this endpoint honours.
  list: () => api.get<SavedQuery[]>('/log-analyzer/queries?page=0&size=200'),
  create: (input: SavedQueryInput) => api.post<SavedQuery>('/log-analyzer/queries', input),
  remove: (id: number) => api.delete<void>(`/log-analyzer/queries/${id}`),
}
