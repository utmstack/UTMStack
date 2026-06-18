import { ApiError, createApiClient } from '@/shared/lib/api-client'
import { FS_API_URL } from '@/shared/config/mode'
import type { TfaInitResponse } from '@/features/auth/types/auth.types'
import type {
  FederationLoginResponse,
  FederationSession,
  FederationTokenPair,
  FederationUser,
} from '../types'

// Talks to the FS's own /api/v1 subpaths (auth), NOT a proxied instance.
const api = createApiClient(FS_API_URL)

export { ApiError as FederationHttpError }

export const federationAuthService = {
  login: (login: string, password: string) =>
    api.post<FederationLoginResponse>('/auth/login', { login, password }),
  refresh: (refreshToken: string) =>
    api.post<FederationTokenPair>('/auth/refresh', { refreshToken }),
  requestPasswordReset: (email: string) =>
    api.post<void>('/auth/reset-password/init', { email }),
  finishPasswordReset: (key: string, newPassword: string) =>
    api.post<void>('/auth/reset-password/finish', { key, new_password: newPassword }),
  me: () => api.get<FederationUser>('/auth/me'),
  updateMe: (input: { firstName?: string; lastName?: string; email?: string; langKey?: string }) =>
    api.put<FederationUser>('/auth/me', input),
  uploadAvatar: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post<FederationUser>('/auth/me/avatar', fd)
  },
  removeAvatar: () => api.delete<FederationUser>('/auth/me/avatar'),
  changePassword: (currentPassword: string, newPassword: string) =>
    api.post<void>('/auth/change-password', { currentPassword, newPassword }),
  listSessions: () => api.get<FederationSession[]>('/auth/sessions'),
  revokeSession: (id: number) => api.delete<void>(`/auth/sessions/${id}`),
  revokeOtherSessions: () => api.delete<void>('/auth/sessions'),
  // 2FA (TOTP only). Enrollment endpoints return snake_case so they reuse the
  // shared TfaInitResponse type and the shared TwoFactorBody renders unchanged.
  verifyTfaCode: (preAuthToken: string, code: string) =>
    api.post<FederationLoginResponse>('/auth/tfa/verify-code', { preAuthToken, code }),
  tfaInit: () => api.post<TfaInitResponse>('/tfa/init', { method: 'TOTP' }),
  tfaVerify: (code: string) => api.post<void>('/tfa/verify', { code }),
  tfaComplete: () => api.post<void>('/tfa/complete', { method: 'TOTP' }),
  tfaDisable: (password: string) => api.post<void>('/tfa/disable', { password }),
}
