export interface User {
  id: number
  login: string
  email?: string
  first_name?: string
  last_name?: string
  activated: boolean
  lang_key?: string
  image_url?: string
  tfa_enabled?: boolean
  tfa_method?: TfaMethod
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_at: string
  refresh_expires_at: string
}

export type TfaMethod = 'EMAIL' | 'TOTP'

/**
 * Raw response of POST /auth/login and POST /auth/tfa/verify-code. It is either a
 * completed login (TokenPair fields + `user`) or a 2FA challenge (`tfa_required`
 * + `pre_auth_token`), never both — hence every field is optional here.
 */
export interface LoginResponse {
  access_token?: string
  refresh_token?: string
  token_type?: string
  expires_at?: string
  refresh_expires_at?: string
  user?: User
  tfa_required?: boolean
  tfa_method?: TfaMethod
  pre_auth_token?: string
}

export interface LoginRequest {
  login: string
  password: string
}

export interface TfaVerifyCodeRequest {
  pre_auth_token: string
  code: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

export interface ResetPasswordInitRequest {
  email: string
}

export interface ResetPasswordFinishRequest {
  key: string
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

/* ---- TFA enrollment (enabling 2FA on the current account) ---- */

export interface TfaInitRequest {
  method: TfaMethod
}

export interface TfaInitResponse {
  method: TfaMethod
  qr_data_url?: string
  otp_auth_url?: string
  email_sent?: boolean
  expires_at: string
}

export interface TfaVerifyRequest {
  method: TfaMethod
  code: string
}

export interface TfaCompleteRequest {
  method: TfaMethod
}

export interface TfaDisableRequest {
  password: string
}

export interface TfaRefreshResponse {
  email_sent: boolean
  expires_at: string
  cooldown_until?: string
}

export interface TfaEnrollmentRequest {
  stage: 'INIT' | 'VERIFY' | 'COMPLETE'
  method: TfaMethod
  code?: string
}

export interface TfaEnrollmentResponse {
  stage: string
  init?: TfaInitResponse
  verified?: boolean
  enabled?: boolean
}
