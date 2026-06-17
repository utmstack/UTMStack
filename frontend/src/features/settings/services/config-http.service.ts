import { createApiClient } from '@/shared/lib/api-client'
import { FS_API_URL, IS_FEDERATION } from '@/shared/config/mode'

// In federation mode this config lives on the FS itself (/fs/config), not on a
// proxied instance.
const api = createApiClient(IS_FEDERATION ? FS_API_URL : undefined)

export interface ConfigEntry {
  key: string
  value?: string
  is_secret: boolean
  is_set: boolean
  label?: string
  description?: string
  updated_at: string
  updated_by?: string
}

export interface UpsertConfig {
  value: string
  is_secret?: boolean
  label?: string
  description?: string
}

/** One SMTP config tested by POST /config/checkEmailConfiguration (mirrors domain.MailConfig). */
export interface MailCheckConfig {
  host: string
  port: number
  username: string
  password: string
  authType: string
  from: string
}

export interface CheckMailResponse {
  success: boolean
  message: string
}

export const configHttpService = {
  list: () => api.get<ConfigEntry[]>('/config'),
  get: (key: string) => api.get<ConfigEntry>(`/config/${encodeURIComponent(key)}`),
  set: (key: string, input: UpsertConfig) =>
    api.put<ConfigEntry>(`/config/${encodeURIComponent(key)}`, input),
  rotate: (key: string) => api.post<ConfigEntry>(`/config/${encodeURIComponent(key)}/rotate`),
  remove: (key: string) => api.delete<void>(`/config/${encodeURIComponent(key)}`),
  // Sends a test email with the supplied SMTP config (an empty password reuses
  // the stored one server-side). The test message goes to each config's `from`.
  checkMail: (configs: MailCheckConfig[]) =>
    api.post<CheckMailResponse>('/config/checkEmailConfiguration', configs),
}
