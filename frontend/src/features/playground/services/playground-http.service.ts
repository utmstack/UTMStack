import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type { TestPipelineRequest, TestPipelineResponse, TestRuleRequest, TestRuleResponse } from '../types'

const api = createApiClient()

export { ApiError as PlaygroundHttpError }

// Shared playground routes live under the eventprocessing module group.
const BASE = '/eventprocessing/playground'

export const playgroundHttpService = {
  testPipeline: (req: TestPipelineRequest) => api.post<TestPipelineResponse>(`${BASE}/test-pipeline`, req),
  testRule: (req: TestRuleRequest) => api.post<TestRuleResponse>(`${BASE}/test-rule`, req),
}
