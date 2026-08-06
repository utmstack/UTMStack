/* Types mirror the backend iam DTOs (modules/iam/dto). */

export type UserStatus = 'pending' | 'active' | 'inactive' | 'suspended'

export interface RoleDigest {
  name: string
  display_name: string
}

/** Base user shape (UserResponse). The email is the login. */
export interface UserBase {
  id: string
  email: string
  name?: string
  status: UserStatus
  lang_key?: string
  image_url?: string
  /** Signs in through an identity provider, so there is no local password to
   * reset and the directory may rewrite the roles at the next login. */
  federated: boolean
  tfa_enabled: boolean
}

/** Row in GET /users (UserListItem). */
export interface UserListItem extends UserBase {
  roles?: RoleDigest[]
}

/** GET /users/:id (UserDetailResponse). */
export interface UserDetail extends UserBase {
  created_at: string
  updated_at: string
  roles?: RoleDigest[]
}

export interface PageInfo {
  page: number
  page_size: number
  total_items: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface UserListResponse {
  data: UserListItem[]
  page_info: PageInfo
}

export interface ListUsersQuery {
  page?: number
  page_size?: number
  search?: string
}

export interface CreateUserRequest {
  email: string
  name?: string
  lang_key?: string
  role_names?: string[]
}

export interface UpdateUserRequest {
  email?: string
  name?: string
  lang_key?: string
  status?: UserStatus
}

export interface AssignRolesRequest {
  role_names: string[]
}

/* ---- roles ---- */

export interface Role {
  id: string
  name: string
  display_name: string
  description?: string
  /** Seeded with the platform and the same for every tenant, so it cannot be
   * edited or deleted. */
  system: boolean
}

export interface Permission {
  name: string
  description?: string
}

export interface RoleDetail extends Role {
  permissions: Permission[]
}

export interface RoleUpsertRequest {
  name: string
  display_name?: string
  description?: string
  permissions: string[]
}
