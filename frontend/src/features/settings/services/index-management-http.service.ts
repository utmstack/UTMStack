import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { IndexInfo, IndexPattern, IndexPatternUpsert } from '../types/index-management.types'

const api = createApiClient()

export { ApiError as IndexHttpError }

export const indicesHttpService = {
  // Real OpenSearch indices, paginated. Returns { data, total } from X-Total-Count.
  // page is 1-based (the backend's index/all uses 1-based paging).
  listPaged: (page = 1, size = 25, pattern = '', includeSystem = false) => {
    const p = new URLSearchParams({ page: String(page), size: String(size) })
    if (pattern) p.set('pattern', pattern)
    if (includeSystem) p.set('includeSystemIndex', 'true')
    return api.getPaged<IndexInfo[]>(`/opensearch/index/all?${p.toString()}`)
  },
  // How many real indices a glob matches (used to validate a pattern before create).
  countMatching: async (pattern: string, includeSystem = false) => {
    const p = new URLSearchParams({ page: '1', size: '1', pattern })
    if (includeSystem) p.set('includeSystemIndex', 'true')
    const { total } = await api.getPaged<IndexInfo[]>(`/opensearch/index/all?${p.toString()}`)
    return total
  },
  remove: (names: string[]) => api.post<void>('/opensearch/index/delete-index', names),
}

export const dataViewsHttpService = {
  // Index patterns, paginated server-side. `search` maps to pattern.contains.
  // page is 0-based here (the index-patterns endpoint uses 0-based paging).
  listPaged: (page = 0, size = 25, search = '') => {
    const p = new URLSearchParams({ page: String(page), size: String(size), sort: 'pattern,asc' })
    if (search) p.set('pattern.contains', search)
    return api.getPaged<IndexPattern[]>(`/opensearch/index-patterns?${p.toString()}`)
  },
  // True when a pattern with this exact glob already exists (duplicate guard).
  existsByPattern: async (pattern: string) => {
    const p = new URLSearchParams({ page: '0', size: '1', 'pattern.equals': pattern })
    const { total } = await api.getPaged<IndexPattern[]>(`/opensearch/index-patterns?${p.toString()}`)
    return total > 0
  },
  create: (input: IndexPatternUpsert) => api.post<IndexPattern>('/opensearch/index-patterns', input),
  update: (input: IndexPatternUpsert) => api.put<IndexPattern>('/opensearch/index-patterns', input),
  remove: (id: number) => api.delete<void>(`/opensearch/index-patterns/${id}`),
}
