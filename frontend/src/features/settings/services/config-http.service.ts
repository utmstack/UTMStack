import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

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

export const configHttpService = {
  list: () => api.get<ConfigEntry[]>('/config'),
  get: (key: string) => api.get<ConfigEntry>(`/config/${encodeURIComponent(key)}`),
  set: (key: string, input: UpsertConfig) =>
    api.put<ConfigEntry>(`/config/${encodeURIComponent(key)}`, input),
  rotate: (key: string) => api.post<ConfigEntry>(`/config/${encodeURIComponent(key)}/rotate`),
  remove: (key: string) => api.delete<void>(`/config/${encodeURIComponent(key)}`),
}
