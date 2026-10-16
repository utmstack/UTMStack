import { getStoredTokens } from '@/shared/lib/api-client'
import { IS_FEDERATION } from '@/shared/config/mode'
import { getCurrentInstanceId } from '@/shared/lib/current-instance'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'
const BASE = `${API_URL}/threat-intel`

export async function proxyFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(init.headers as Record<string, string>) }
  const tokens = getStoredTokens()
  if (tokens?.access_token) headers.Authorization = `Bearer ${tokens.access_token}`
  if (IS_FEDERATION) {
    const id = getCurrentInstanceId()
    if (id != null) headers['X-UTM-Instance'] = String(id)
  }
  const res = await fetch(`${BASE}${path}`, { ...init, headers, signal: init.signal })
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      msg = JSON.parse(await res.text()).error ?? msg
    } catch {
    }
    throw new Error(msg)
  }
  return (await res.json()) as T
}
