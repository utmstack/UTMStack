import { ApiError, createApiClient } from '@/shared/lib/api-client'
import { FS_API_URL } from '@/shared/config/mode'
import type {
  FederationCreateUser,
  FederationTeamUser,
  FederationUpdateUser,
  FederationUserList,
} from '../types'

// FS team management — the FS's own /api/v1/users subpath, not a proxied instance.
const api = createApiClient(FS_API_URL)

export { ApiError as FederationTeamError }

export interface ListQuery {
  page?: number
  page_size?: number
  search?: string
}

function listQuery(q: ListQuery): string {
  const p = new URLSearchParams()
  if (q.page) p.set('page', String(q.page))
  if (q.page_size) p.set('page_size', String(q.page_size))
  if (q.search) p.set('search', q.search)
  const s = p.toString()
  return s ? `?${s}` : ''
}

export const federationTeamService = {
  list: (q: ListQuery = {}) => api.get<FederationUserList>(`/users${listQuery(q)}`),
  create: (input: FederationCreateUser) => api.post<FederationTeamUser>('/users', input),
  update: (id: number, input: FederationUpdateUser) =>
    api.put<FederationTeamUser>(`/users/${id}`, input),
  deactivate: (id: number) => api.delete<void>(`/users/${id}`),
  resetTfa: (id: number) => api.post<void>(`/users/${id}/tfa/disable`, {}),
  resendInvite: (id: number) => api.post<void>(`/users/${id}/resend-invite`, {}),
}
