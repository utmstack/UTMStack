import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  AssignRolesRequest,
  CreateUserRequest,
  ListUsersQuery,
  Role,
  RoleDetail,
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
  get: (id: number) => api.get<UserDetail>(`/users/${id}`),
  create: (input: CreateUserRequest) => api.post<UserDetail>('/users', input),
  update: (id: number, input: UpdateUserRequest) => api.put<UserDetail>(`/users/${id}`, input),
  deactivate: (id: number) => api.delete<void>(`/users/${id}`),
  assignRoles: (id: number, role_names: string[]) =>
    api.put<void>(`/users/${id}/roles`, { role_names } satisfies AssignRolesRequest),
}

export const rolesHttpService = {
  list: () => api.get<Role[]>('/roles'),
  get: (name: string) => api.get<RoleDetail>(`/roles/${name}`),
}
