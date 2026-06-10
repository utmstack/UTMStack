/* Types mirror the backend iam DTOs (modules/iam/dto). */

export type TfaMethod = 'EMAIL' | 'TOTP'

export interface RoleDigest {
  name: string
  display_name: string
}

/** Base user shape (UserResponse). */
export interface UserBase {
  id: number
  login: string
  email?: string
  first_name?: string
  last_name?: string
  activated: boolean
  lang_key?: string
  image_url?: string
  tfa_enabled: boolean
  tfa_method?: TfaMethod
}

/** Row in GET /users (UserListItem): base + roles + default_password. */
export interface UserListItem extends UserBase {
  default_password: boolean
  roles?: RoleDigest[]
}

/** GET /users/:id (UserDetailResponse). */
export interface UserDetail extends UserBase {
  default_password: boolean
  created_by?: string
  created_date?: string
  last_modified_by?: string
  last_modified_date?: string
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
  login: string
  email: string
  first_name?: string
  last_name?: string
  lang_key?: string
  role_names?: string[]
}

export interface UpdateUserRequest {
  email?: string
  first_name?: string
  last_name?: string
  lang_key?: string
  activated?: boolean
}

export interface AssignRolesRequest {
  role_names: string[]
}

/* ---- roles ---- */

export interface Role {
  name: string
  display_name: string
  description?: string
}

export interface Permission {
  id: number
  name: string
  resource: string
  action: string
  description?: string
}

export interface RoleDetail extends Role {
  permissions: Permission[]
}
