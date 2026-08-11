import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as RegexPatternsHttpError }

/* ─── Types (mirror backend modules/eventprocessing/dto regex_pattern) ── */

/**
 * A reusable regular expression referenced from parsing pipelines. Identity is
 * `patternId`, the name used to reference it. The whole vocabulary ships with
 * the release and is read-only, so there is no ownership to report.
 */
export interface RegexPattern {
  patternId: string
  patternDefinition: string
}

export interface RegexPatternListQuery {
  search?: string
  page?: number // 0-based
  size?: number
}

function listQuery(q: RegexPatternListQuery): string {
  const p = new URLSearchParams()
  if (q.search) p.set('search', q.search)
  p.set('page', String(q.page ?? 0))
  p.set('size', String(q.size ?? 20))
  return p.toString()
}

const BASE = '/eventprocessing/regex-pattern'

// Read-only. A shared vocabulary referenced from pipeline YAMLs as {{.name}};
// the API exposes no create, update or delete.
export const regexPatternsHttpService = {
  // Returns { data: RegexPattern[], total } — total comes from X-Total-Count.
  list: (q: RegexPatternListQuery = {}) => api.getPaged<RegexPattern[]>(`${BASE}?${listQuery(q)}`),
  get: (patternId: string) => api.get<RegexPattern>(`${BASE}/${encodeURIComponent(patternId)}`),
}
