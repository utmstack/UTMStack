import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  ChangePasswordRequest,
  LoginRequest,
  LoginResponse,
  Session,
  UpdateMeRequest,
  User,
} from '../types/auth.types'

const api = createApiClient()

export { ApiError as AuthHttpError }

export const authHttpService = {
  login: (input: LoginRequest) => api.post<LoginResponse>('/auth/login', input),
  me: () => api.get<User>('/auth/me'),
  updateMe: (input: UpdateMeRequest) => api.put<User>('/auth/me', input),
  logout: (refresh_token: string) => api.post<void>('/auth/logout', { refresh_token }),
  changePassword: (input: ChangePasswordRequest) =>
    api.post<void>('/auth/change-password', input),
  listSessions: () => api.get<Session[]>('/auth/sessions'),
  revokeSession: (id: number) => api.delete<void>(`/auth/sessions/${id}`),
  revokeOtherSessions: () => api.delete<void>('/auth/sessions'),
  uploadAvatar: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    // axios will set the multipart boundary automatically when given FormData.
    return api.post<User>('/auth/me/avatar', fd)
  },
  removeAvatar: () => api.delete<User>('/auth/me/avatar'),
}
