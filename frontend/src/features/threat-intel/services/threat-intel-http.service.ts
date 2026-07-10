import { createApiClient } from '@/shared/lib/api-client'
import type {
  EntitySearchRequest,
  EntitySearchResponse,
  EntityDetail,
  EntityRelation,
  ThreatFeed,
  ChatRequest,
  ChatResponse,
  UsageInfo,
  TiResult,
  AdvancedSearchRequest,
  AdvancedSearchResponse,
} from '../domain/threat-intel.types'
import { isNotConfigured } from './ti-errors'

const api = createApiClient()
const BASE = '/threat-intel'

async function wrap<T>(fn: () => Promise<T>): Promise<TiResult<T>> {
  try {
    return { kind: 'ok', value: await fn() }
  } catch (e) {
    if (isNotConfigured(e)) return { kind: 'not-configured' }
    throw e
  }
}

export const threatIntelHttpService = {
  search: (body: EntitySearchRequest) =>
    wrap(() => api.post<EntitySearchResponse>(`${BASE}/search`, body)),
  searchAdvanced: (body: AdvancedSearchRequest, params?: { limit?: number; page?: number }) => {
    const qs = new URLSearchParams()
    if (params?.limit !== undefined) qs.set('limit', String(params.limit))
    if (params?.page !== undefined) qs.set('page', String(params.page))
    const query = qs.toString()
    const path = `${BASE}/search/advanced${query ? `?${query}` : ''}`
    return wrap(() => api.post<AdvancedSearchResponse>(path, body))
  },
  entity: (id: string) =>
    wrap(() => api.get<EntityDetail>(`${BASE}/entity/${encodeURIComponent(id)}`)),
  relations: (id: string) =>
    wrap(() => api.get<EntityRelation[]>(`${BASE}/entity/${encodeURIComponent(id)}/relations`)),
  feeds: () => wrap(() => api.get<ThreatFeed[]>(`${BASE}/feeds`)),
  chat: (body: ChatRequest) => wrap(() => api.post<ChatResponse>(`${BASE}/ai/chat`, body)),
  usage: () => wrap(() => api.get<UsageInfo>(`${BASE}/usage`)),
}
