import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as SocAiHttpError }

/** GET /soc-ai/config — current config with secrets masked. */
export interface SocAiConfig {
  configured: boolean
  provider: string
  model: string
  url: string
  apiKeySet: boolean
  authType: string
  authHeaderName: string
  customHeaders: Record<string, string> // values masked ("*****")
  maxTokens: number
  maxToolIterations: number
  autoAnalyze: boolean
  /** Enabled permission groups (alerts, incidents, soar, …). */
  capabilities: string[]
}

/** PUT /soc-ai/config body. Secrets sent as "*****" (or empty) keep the stored value. */
export interface SocAiConfigInput {
  provider: string
  model: string
  url: string
  apiKey: string
  authType: string
  authHeaderName: string
  customHeaders: Record<string, string>
  maxTokens: number
  maxToolIterations: number
  autoAnalyze: boolean
  capabilities: string[]
}

export const socaiHttpService = {
  get: () => api.get<SocAiConfig>('/soc-ai/config'),
  // The save runs a live connection check server-side; a 400 ApiError carries the
  // verification message (bad key, unreachable endpoint, no tool-calling…).
  update: (input: SocAiConfigInput) => api.put<SocAiConfig>('/soc-ai/config', input),
}
