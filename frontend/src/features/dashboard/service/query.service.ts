import { createApiClient } from '@/shared/lib/api-client'
import type { QueryResult, VizSpec } from '@/features/dashboard/types'

const BASE_URL = '/visualizations/query'

/**
 * Answering a widget: the spec says what to ask, the backend asks the event
 * store. The tenant is not part of the spec — it comes from the session — so
 * the same widget on two accounts answers about two different sets of records.
 */
export interface QueryService {
  run(spec: VizSpec): Promise<QueryResult>
}

export function createQueryService(baseUrl?: string): QueryService {
  const api = createApiClient(baseUrl)
  return {
    run: (spec) => api.post<QueryResult>(BASE_URL, spec),
  }
}
