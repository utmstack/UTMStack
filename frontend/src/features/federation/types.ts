/** Mirrors the Federation Service DTOs (camelCase JSON). */

import type { TfaFactorType } from '@/features/auth/types/auth.types'

export interface FederationUser {
  id: number
  login: string
  email?: string
  firstName?: string
  lastName?: string
  imageUrl?: string
  langKey?: string
  tfaEnabled?: boolean
  tfaMethod?: TfaFactorType
}

export interface FederationTokenPair {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresAt: string
  refreshExpiresAt: string
}

export interface FederationLoginResponse extends Partial<FederationTokenPair> {
  user: FederationUser
  // Present instead of the token pair when the account has 2FA enabled.
  tfaRequired?: boolean
  tfaMethod?: TfaFactorType
  preAuthToken?: string
}

/** Mirrors the FS SessionResponse (snake_case, matches the shared Session type). */
export interface FederationSession {
  id: number
  ip?: string
  user_agent?: string
  created_at: string
  expires_at: string
  current: boolean
}

export interface FederationInstance {
  id: number
  name: string
  baseUrl: string
  tlsSkipVerify: boolean
  createdAt: string
  updatedAt: string
}

export interface FederationInstanceInput {
  name: string
  baseUrl: string
  apiKey: string
  tlsSkipVerify: boolean
}

/* ---- team (flat: every member is an FS admin) ---- */

export interface FederationTeamUser {
  id: number
  login: string
  email?: string
  first_name?: string
  last_name?: string
  image_url?: string
  activated: boolean
  /** Invited but hasn't set a password yet. */
  pending: boolean
  tfa_enabled: boolean
  created_by?: string
  created_at: string
}

export interface FederationPageInfo {
  page: number
  page_size: number
  total_items: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface FederationUserList {
  data: FederationTeamUser[]
  page_info: FederationPageInfo
}

export interface FederationCreateUser {
  login: string
  email: string
  first_name?: string
  last_name?: string
}

export interface FederationUpdateUser {
  email?: string
  first_name?: string
  last_name?: string
  activated?: boolean
}
