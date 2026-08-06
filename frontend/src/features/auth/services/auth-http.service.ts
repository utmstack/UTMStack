import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ChangePasswordRequest,
  LoginRequest,
  LoginResponse,
  ResetPasswordFinishRequest,
  Session,
  TfaDisableRequest,
  TfaEnrollmentRequest,
  TfaEnrollmentResponse,
  TfaVerifyCodeRequest,
  TokenPair,
  UpdateMeRequest,
  User,
} from '../types/auth.types'

const api = createApiClient()

export { ApiError as AuthHttpError }

export const authHttpService = {
  /* ---- login / session lifecycle ---- */
  login: (input: LoginRequest) => api.post<LoginResponse>('/auth/login', input),
  // Second factor: exchange the pre-auth token + code for a real token pair.
  verifyTfaCode: (input: TfaVerifyCodeRequest) =>
    api.post<LoginResponse>('/auth/tfa/verify-code', input),
  refresh: (refresh_token: string) =>
    api.post<TokenPair>('/auth/refresh', { refresh_token }),
  logout: (refresh_token: string) => api.post<void>('/auth/logout', { refresh_token }),

  /* ---- password reset (unauthenticated) ---- */
  requestPasswordReset: (email: string) =>
    api.post<void>('/auth/reset-password/init', { email }),
  finishPasswordReset: (input: ResetPasswordFinishRequest) =>
    api.post<void>('/auth/reset-password/finish', input),

  /* ---- current user (me) ---- */
  me: () => api.get<User>('/auth/me'),
  // Lightweight identity probe — returns the login string as text/plain.
  authenticate: () => api.get<string>('/auth/authenticate', { responseType: 'text' }),
  updateMe: (input: UpdateMeRequest) => api.put<User>('/auth/me', input),
  changePassword: (input: ChangePasswordRequest) =>
    api.post<void>('/auth/change-password', input),
  uploadAvatar: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    // axios sets the multipart boundary automatically when given FormData.
    return api.post<User>('/auth/me/avatar', fd)
  },
  removeAvatar: () => api.delete<User>('/auth/me/avatar'),

  /* ---- active sessions ---- */
  listSessions: () => api.get<Session[]>('/auth/sessions'),
  revokeSession: (id: string) => api.delete<void>(`/auth/sessions/${id}`),
  revokeOtherSessions: () => api.delete<void>('/auth/sessions'),

  /* ---- TFA enrollment (authenticated; enabling 2FA on the account) ---- */
  // Tear down 2FA — re-authenticates with the current password.
  tfaDisable: (input: TfaDisableRequest) => api.post<void>('/tfa/disable', input),
  // Unified single-call enrollment (stage = INIT | VERIFY | COMPLETE).
  // One endpoint for the three stages: the separate init/verify/complete routes
  // did the same work by another name.
  tfaEnrollment: (input: TfaEnrollmentRequest) =>
    api.post<TfaEnrollmentResponse>('/tfa/enroll', input),
}
