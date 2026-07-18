import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { TestFilterRequest, TestFilterResponse, TestRuleRequest, TestRuleResponse } from '../types'

const api = createApiClient()

export { ApiError as PlaygroundHttpError }

// Shared playground routes live under the eventprocessing module group.
const BASE = '/eventprocessing/playground'

export const playgroundHttpService = {
  testFilter: (req: TestFilterRequest) => api.post<TestFilterResponse>(`${BASE}/test-filter`, req),
  testRule: (req: TestRuleRequest) => api.post<TestRuleResponse>(`${BASE}/test-rule`, req),
}
