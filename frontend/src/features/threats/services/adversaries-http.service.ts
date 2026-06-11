import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { AdversaryResponse } from '../types/adversary.types'

const api = createApiClient()

export { ApiError as AdversariesHttpError }

export const adversariesHttpService = {
  // POST /adversary/alerts — alerts grouped by adversary host. Body is an optional
  // []FilterType ({ field, operator, value }); empty = no extra filters.
  list: (filters: unknown[] = []) => api.post<AdversaryResponse[]>('/adversary/alerts', filters),
}
