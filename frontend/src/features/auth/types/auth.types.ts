export interface User {
  id: number
  login: string
  email?: string
  first_name?: string
  last_name?: string
  activated: boolean
  lang_key?: string
  image_url?: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_at: string
  refresh_expires_at: string
}

export interface LoginResponse extends TokenPair {
  user: User
}

export interface LoginRequest {
  login: string
  password: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export interface UpdateMeRequest {
  first_name?: string
  last_name?: string
  email?: string
  lang_key?: string
}

export interface Session {
  id: number
  ip?: string
  user_agent?: string
  created_at: string
  expires_at: string
  current: boolean
}
