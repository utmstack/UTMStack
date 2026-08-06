export type UserStatus = 'pending' | 'active' | 'inactive' | 'suspended'

/** The kinds of second factor an account can hold. `recovery` is a spent-once
 * backup code, so it is never enrolled directly from the profile page. */
export type TfaFactorType = 'email' | 'totp' | 'recovery'

export interface User {
  id: string
  email: string
  name?: string
  status: UserStatus
  lang_key?: string
  image_url?: string
  /** True when the account signs in through an identity provider, in which case
   * it has no local password to change. */
  federated: boolean
  tfa_enabled: boolean
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_at: string
  refresh_expires_at: string
}

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
  /** Which factor to ask for, so the screen names the right one. */
  tfa_type?: TfaFactorType
  pre_auth_token?: string
}

export interface LoginRequest {
  login: string
  password: string
  /** Which directory to bind against, when the user picked one. Omitted, the
   * backend tries every directory the tenant has. */
  provider_id?: string
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
  email?: string
  name?: string
  lang_key?: string
}

export interface Session {
  id: string
  ip?: string
  user_agent?: string
  created_at: string
  expires_at: string
  current: boolean
}

/* ---- TFA enrollment (enabling 2FA on the current account) ---- */

export interface TfaInitResponse {
  type: TfaFactorType
  factor_id: string
  qr_data_url?: string
  otp_auth_url?: string
  email_sent?: boolean
  expires_at: string
}

export interface TfaDisableRequest {
  password: string
}

/** The three stages go to one endpoint, POST /tfa/enroll. */
export interface TfaEnrollmentRequest {
  stage: 'INIT' | 'VERIFY' | 'COMPLETE'
  type: TfaFactorType
  code?: string
}

export interface TfaEnrollmentResponse {
  stage: string
  init?: TfaInitResponse
  verified?: boolean
  enabled?: boolean
}
