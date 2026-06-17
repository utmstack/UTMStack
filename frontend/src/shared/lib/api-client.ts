import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { FS_API_URL, IS_FEDERATION } from '@/shared/config/mode'
import { getCurrentInstanceId } from '@/shared/lib/current-instance'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

export interface AuthTokens {
  access_token: string
  refresh_token: string
  expires_at?: string
  refresh_expires_at?: string
}

const TOKEN_KEY = 'utmstack_tokens'

export function getStoredTokens(): AuthTokens | null {
  const raw = localStorage.getItem(TOKEN_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as AuthTokens
  } catch {
    return null
  }
}

export function storeTokens(tokens: AuthTokens): void {
  localStorage.setItem(TOKEN_KEY, JSON.stringify(tokens))
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY)
}

let logoutHandler: ((reason: string) => void) | null = null
export function setLogoutHandler(handler: (reason: string) => void): void {
  logoutHandler = handler
}

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

let refreshPromise: Promise<string | null> | null = null

async function refreshAccessToken(): Promise<string | null> {
  const tokens = getStoredTokens()
  if (!tokens?.refresh_token) return null
  try {
    const response = await axios.post<AuthTokens>(`${API_URL}/auth/refresh`, {
      refresh_token: tokens.refresh_token,
    })
    storeTokens(response.data)
    return response.data.access_token
  } catch {
    return null
  }
}

// Federation rotates the FS access token via /fs/auth/refresh, which speaks
// camelCase. The proxy authenticates every instance call with the FS JWT, so a
// 401 from either the FS API or a proxied instance means the FS token expired.
async function refreshFederationToken(): Promise<string | null> {
  const tokens = getStoredTokens()
  if (!tokens?.refresh_token) return null
  try {
    const response = await axios.post<{
      accessToken: string
      refreshToken: string
      expiresAt?: string
      refreshExpiresAt?: string
    }>(`${FS_API_URL}/auth/refresh`, { refreshToken: tokens.refresh_token })
    const next: AuthTokens = {
      access_token: response.data.accessToken,
      refresh_token: response.data.refreshToken,
      expires_at: response.data.expiresAt,
      refresh_expires_at: response.data.refreshExpiresAt,
    }
    storeTokens(next)
    return next.access_token
  } catch {
    return null
  }
}

export class ApiError extends Error {
  constructor(message: string, public status: number, public code?: string) {
    super(message)
    this.name = 'ApiError'
  }
}

interface RequestOptions {
  body?: unknown
  headers?: Record<string, string>
  responseType?: 'json' | 'text' | 'blob'
}

const instances: Record<string, AxiosInstance> = {}

function getAxiosInstance(baseURL: string): AxiosInstance {
  if (instances[baseURL]) return instances[baseURL]

  const instance = axios.create({ baseURL })

  instance.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
    const tokens = getStoredTokens()
    if (tokens?.access_token) {
      config.headers.Authorization = `Bearer ${tokens.access_token}`
    }
    // Federation: stamp the selected instance so the FS proxy can route the call.
    // The FS's own API (/fs) isn't proxied, so it's left untouched.
    if (IS_FEDERATION && baseURL !== FS_API_URL) {
      const instanceId = getCurrentInstanceId()
      if (instanceId != null) {
        config.headers['X-UTM-Instance'] = String(instanceId)
      }
    }
    return config
  })

  instance.interceptors.response.use(
    (r) => r,
    async (error: AxiosError) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
      const url = originalRequest?.url ?? ''
      const isAuthEndpoint = url.includes('/auth/login') || url.includes('/auth/refresh')

      if (error.response?.status === 401 && !originalRequest._retry && !isAuthEndpoint) {
        originalRequest._retry = true
        const doRefresh = IS_FEDERATION ? refreshFederationToken : refreshAccessToken
        const newToken = refreshPromise ?? (refreshPromise = doRefresh())
        const tk = await newToken
        refreshPromise = null
        if (tk) {
          originalRequest.headers.Authorization = `Bearer ${tk}`
          return instance(originalRequest)
        }
        clearTokens()
        if (logoutHandler) {
          logoutHandler('Session expired')
        }
      }

      const data = error.response?.data as { error?: string } | undefined
      const message = data?.error ?? error.message ?? 'Request failed'
      throw new ApiError(message, error.response?.status ?? 0)
    }
  )

  instances[baseURL] = instance
  return instance
}

export async function apiRequest<T>(
  baseUrl: string,
  endpoint: string,
  method: HttpMethod = 'GET',
  options: RequestOptions = {}
): Promise<T> {
  const instance = getAxiosInstance(baseUrl)
  const response = await instance.request<T>({
    url: endpoint,
    method,
    data: options.body,
    headers: options.headers,
    responseType: options.responseType ?? 'json',
  })
  return response.data
}

export interface Paged<T> {
  data: T
  /** From the X-Total-Count response header (falls back to the array length). */
  total: number
}

/** GET that also surfaces the X-Total-Count header for classic pagination. */
export async function apiRequestPaged<T>(
  baseUrl: string,
  endpoint: string,
  options: RequestOptions = {}
): Promise<Paged<T>> {
  const instance = getAxiosInstance(baseUrl)
  const response = await instance.request<T>({
    url: endpoint,
    method: 'GET',
    headers: options.headers,
    responseType: options.responseType ?? 'json',
  })
  const raw = response.headers['x-total-count']
  const total = raw != null ? Number(raw) : Array.isArray(response.data) ? response.data.length : 0
  return { data: response.data, total: Number.isFinite(total) ? total : 0 }
}

/** POST that also surfaces the X-Total-Count header for classic pagination. */
export async function apiRequestPagedPost<T>(
  baseUrl: string,
  endpoint: string,
  body?: unknown,
  options: RequestOptions = {}
): Promise<Paged<T>> {
  const instance = getAxiosInstance(baseUrl)
  const response = await instance.request<T>({
    url: endpoint,
    method: 'POST',
    data: body,
    headers: options.headers,
    responseType: options.responseType ?? 'json',
  })
  const raw = response.headers['x-total-count']
  const total = raw != null ? Number(raw) : Array.isArray(response.data) ? response.data.length : 0
  return { data: response.data, total: Number.isFinite(total) ? total : 0 }
}

export function createApiClient(baseUrl: string = API_URL) {
  return {
    get: <T,>(endpoint: string, options?: RequestOptions) =>
      apiRequest<T>(baseUrl, endpoint, 'GET', options),
    getPaged: <T,>(endpoint: string, options?: RequestOptions) =>
      apiRequestPaged<T>(baseUrl, endpoint, options),
    postPaged: <T,>(endpoint: string, body?: unknown, options?: RequestOptions) =>
      apiRequestPagedPost<T>(baseUrl, endpoint, body, options),
    post: <T,>(endpoint: string, body?: unknown, options?: RequestOptions) =>
      apiRequest<T>(baseUrl, endpoint, 'POST', { ...options, body }),
    put: <T,>(endpoint: string, body?: unknown, options?: RequestOptions) =>
      apiRequest<T>(baseUrl, endpoint, 'PUT', { ...options, body }),
    patch: <T,>(endpoint: string, body?: unknown, options?: RequestOptions) =>
      apiRequest<T>(baseUrl, endpoint, 'PATCH', { ...options, body }),
    delete: <T,>(endpoint: string, options?: RequestOptions) =>
      apiRequest<T>(baseUrl, endpoint, 'DELETE', options),
  }
}
