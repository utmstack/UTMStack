import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  AssignRolesRequest,
  CreateUserRequest,
  ListUsersQuery,
  Permission,
  Role,
  RoleDetail,
  RoleUpsertRequest,
  UpdateUserRequest,
  UserDetail,
  UserListResponse,
} from '../types/team.types'

const api = createApiClient()

export { ApiError as TeamHttpError }

function listQuery(q: ListUsersQuery): string {
  const p = new URLSearchParams()
  if (q.page) p.set('page', String(q.page))
  if (q.page_size) p.set('page_size', String(q.page_size))
  if (q.search) p.set('search', q.search)
  const s = p.toString()
  return s ? `?${s}` : ''
}

export const usersHttpService = {
  list: (q: ListUsersQuery = {}) => api.get<UserListResponse>(`/users${listQuery(q)}`),
  get: (id: string) => api.get<UserDetail>(`/users/${id}`),
  create: (input: CreateUserRequest) => api.post<UserDetail>('/users', input),
  update: (id: string, input: UpdateUserRequest) => api.put<UserDetail>(`/users/${id}`, input),
  deactivate: (id: string) => api.delete<void>(`/users/${id}`),
  assignRoles: (id: string, role_names: string[]) =>
    api.put<void>(`/users/${id}/roles`, { role_names } satisfies AssignRolesRequest),
  // Admin reset of a user's 2FA (e.g. lost authenticator). The user re-enrolls at
  // next login. No password required — guarded by the users.write permission.
  resetTfa: (id: string) => api.post<void>(`/users/${id}/tfa/disable`, {}),
}

export const rolesHttpService = {
  list: () => api.get<Role[]>('/roles'),
  get: (id: string) => api.get<RoleDetail>(`/roles/${id}`),
  // The full catalog, which is what a role editor offers. A role's own
  // permissions are only ever a subset of it.
  listPermissions: () => api.get<Permission[]>('/roles/permissions'),
  create: (input: RoleUpsertRequest) => api.post<RoleDetail>('/roles', input),
  update: (id: string, input: RoleUpsertRequest) => api.put<RoleDetail>(`/roles/${id}`, input),
  remove: (id: string) => api.delete<void>(`/roles/${id}`),
}
