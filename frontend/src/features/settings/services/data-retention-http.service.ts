import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { PolicySettings } from '../types/data-retention.types'

const api = createApiClient()

export { ApiError as DataRetentionHttpError }

/**
 * Data retention is OpenSearch ISM (Index State Management). GET returns 200 with
 * an EMPTY body when no policy is configured yet, and the route 404s entirely when
 * OpenSearch isn't configured. Both cases are handled by the caller.
 */
export const dataRetentionHttpService = {
  get: () => api.get<PolicySettings | ''>('/opensearch/index-policy/policy'),
  update: (input: PolicySettings) => api.put<unknown>('/opensearch/index-policy/policy', input),
}
