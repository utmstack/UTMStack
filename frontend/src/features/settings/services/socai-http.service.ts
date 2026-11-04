import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as SocAiHttpError }

/** GET /soc-ai/config — current config with secrets masked. */
export interface SocAiConfig {
  configured: boolean
  /**
   * True when what is shown is the instance default rather than this tenant's
   * own configuration. Saving replaces it with an override; resetting drops the
   * override and goes back to following the instance.
   */
  inherited: boolean
  provider: string
  model: string
  url: string
  apiKeySet: boolean
  authType: string
  authHeaderName: string
  customHeaders: Record<string, string> // values masked ("*****")
  maxTokens: number
  maxToolIterations: number
  /** Triage every new alert automatically. */
  autoAnalyze: boolean
  /**
   * The triage agent may create/update incidents when it decides an alert
   * warrants one. Every other tool group (dashboards, compliance,
   * correlation, datasources, SOAR) only matters to the interactive chat,
   * where a person is already driving the conversation, so it's granted
   * unconditionally — configuring SOC-AI at all is the signal it's meant
   * to be used.
   */
  allowIncidents: boolean
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
  allowIncidents: boolean
}

/** GET /soc-ai/usage — what this tenant has spent today against its allowance. */
export interface SocAiUsage {
  /** Negative means no cap. Zero means the feature is denied. */
  limit: number
  used: number
  remaining?: number
  resetsAt: string
}

export const socaiHttpService = {
  get: () => api.get<SocAiConfig>('/soc-ai/config'),
  // The save runs a live connection check server-side; a 400 ApiError carries the
  // verification message (bad key, unreachable endpoint, no tool-calling…).
  update: (input: SocAiConfigInput) => api.put<SocAiConfig>('/soc-ai/config', input),
  // Drops this tenant's override so it follows the instance default again.
  // Refused (400) when there is no instance default to fall back to.
  resetToDefault: () => api.delete<SocAiConfig>('/soc-ai/config'),
  usage: () => api.get<SocAiUsage>('/soc-ai/usage'),
}
