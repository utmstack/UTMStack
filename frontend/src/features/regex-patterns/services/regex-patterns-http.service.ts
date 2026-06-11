import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as RegexPatternsHttpError }

/* ─── Types (mirror backend modules/eventprocessing/dto regex_pattern) ── */

/**
 * A reusable regular expression referenced from parsing filters. Identity is
 * `patternId` (the name used to reference it). System patterns are read-only.
 */
export interface RegexPattern {
  patternId: string
  patternDescription: string
  patternDefinition: string
  systemOwner: boolean
}

export interface SaveRegexPatternRequest {
  patternId: string
  patternDescription: string
  patternDefinition: string
}

export interface RegexPatternListQuery {
  search?: string
  system?: boolean // nil = both; true = system only; false = user only
  page?: number // 0-based
  size?: number
}

function listQuery(q: RegexPatternListQuery): string {
  const p = new URLSearchParams()
  if (q.search) p.set('search', q.search)
  if (q.system != null) p.set('system', String(q.system))
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 20))
  return p.toString()
}

const BASE = '/eventprocessing/regex-pattern'

export const regexPatternsHttpService = {
  // Returns { data: RegexPattern[], total } — total comes from X-Total-Count.
  list: (q: RegexPatternListQuery = {}) => api.getPaged<RegexPattern[]>(`${BASE}?${listQuery(q)}`),
  get: (patternId: string) => api.get<RegexPattern>(`${BASE}/${encodeURIComponent(patternId)}`),
  create: (input: SaveRegexPatternRequest) => api.post<RegexPattern>(BASE, input),
  update: (input: SaveRegexPatternRequest) => api.put<RegexPattern>(BASE, input),
  remove: (patternId: string) => api.delete<{ message: string }>(`${BASE}/${encodeURIComponent(patternId)}`),
}
