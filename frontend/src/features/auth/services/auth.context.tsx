import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import {
  clearTokens,
  getStoredTokens,
  setLogoutHandler,
  storeTokens,
} from '@/shared/lib/api-client'
import i18n from '@/shared/i18n'
import { AuthHttpError, authHttpService } from './auth-http.service'
import type {
  ChangePasswordRequest,
  LoginRequest,
  LoginResponse,
  ResetPasswordFinishRequest,
  TfaMethod,
  UpdateMeRequest,
  User,
} from '../types/auth.types'

/**
 * Result of a login attempt: either fully authenticated, or a 2FA challenge that
 * must be completed with `verifyTfaCode`.
 */
export type LoginOutcome =
  | { status: 'authenticated' }
  | { status: 'tfa_required'; method: TfaMethod; preAuthToken: string }

interface AuthContextValue {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (input: LoginRequest) => Promise<LoginOutcome>
  verifyTfaCode: (preAuthToken: string, code: string) => Promise<void>
  logout: () => Promise<void>
  changePassword: (input: ChangePasswordRequest) => Promise<void>
  requestPasswordReset: (email: string) => Promise<void>
  finishPasswordReset: (input: ResetPasswordFinishRequest) => Promise<void>
  updateMe: (input: UpdateMeRequest) => Promise<void>
  uploadAvatar: (file: File) => Promise<void>
  removeAvatar: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(() => getStoredTokens() !== null)

  const reset = useCallback(() => {
    clearTokens()
    setUser(null)
  }, [])

  // Persist tokens + set the user from a completed login response (post-2FA or
  // plain login). Only called on the success branch, where the tokens are present.
  const completeLogin = useCallback((result: LoginResponse) => {
    storeTokens({
      access_token: result.access_token ?? '',
      refresh_token: result.refresh_token ?? '',
      expires_at: result.expires_at,
      refresh_expires_at: result.refresh_expires_at,
    })
    if (result.user) setUser(result.user)
  }, [])

  const login = useCallback(
    async (input: LoginRequest): Promise<LoginOutcome> => {
      const result = await authHttpService.login(input)
      if (result.tfa_required && result.pre_auth_token) {
        return {
          status: 'tfa_required',
          method: result.tfa_method ?? 'EMAIL',
          preAuthToken: result.pre_auth_token,
        }
      }
      completeLogin(result)
      return { status: 'authenticated' }
    },
    [completeLogin]
  )

  const verifyTfaCode = useCallback(
    async (preAuthToken: string, code: string) => {
      const result = await authHttpService.verifyTfaCode({
        pre_auth_token: preAuthToken,
        code,
      })
      completeLogin(result)
    },
    [completeLogin]
  )

  const logout = useCallback(async () => {
    const tokens = getStoredTokens()
    if (tokens?.refresh_token) {
      try {
        await authHttpService.logout(tokens.refresh_token)
      } catch {
        // best-effort; clearing tokens locally is what matters
      }
    }
    reset()
  }, [reset])

  const changePassword = useCallback(async (input: ChangePasswordRequest) => {
    await authHttpService.changePassword(input)
  }, [])

  const requestPasswordReset = useCallback(async (email: string) => {
    await authHttpService.requestPasswordReset(email)
  }, [])

  const finishPasswordReset = useCallback(async (input: ResetPasswordFinishRequest) => {
    await authHttpService.finishPasswordReset(input)
  }, [])

  const updateMe = useCallback(async (input: UpdateMeRequest) => {
    const updated = await authHttpService.updateMe(input)
    setUser(updated)
  }, [])

  const uploadAvatar = useCallback(async (file: File) => {
    const updated = await authHttpService.uploadAvatar(file)
    setUser(updated)
  }, [])

  const removeAvatar = useCallback(async () => {
    const updated = await authHttpService.removeAvatar()
    setUser(updated)
  }, [])

  // Re-fetch the current user (e.g. after enrolling or disabling 2FA).
  const refreshUser = useCallback(async () => {
    const u = await authHttpService.me()
    setUser(u)
  }, [])

  // Hydrate from stored tokens on mount.
  useEffect(() => {
    const tokens = getStoredTokens()
    if (!tokens?.access_token) {
      setIsLoading(false)
      return
    }
    let cancelled = false
    authHttpService
      .me()
      .then((u) => {
        if (!cancelled) setUser(u)
      })
      .catch((err) => {
        // 401 here means refresh also failed (interceptor already cleared tokens).
        if (err instanceof AuthHttpError && err.status === 401) {
          if (!cancelled) reset()
          return
        }
        if (!cancelled) reset()
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [reset])

  // Wire the api-client logout fallback (fired when refresh fails).
  useEffect(() => {
    setLogoutHandler(() => {
      setUser(null)
    })
  }, [])

  // Apply the user's saved language preference once known (lang_key → i18n).
  useEffect(() => {
    if (user?.lang_key) {
      void i18n.changeLanguage(user.lang_key)
    }
  }, [user?.lang_key])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: user !== null,
      isLoading,
      login,
      verifyTfaCode,
      logout,
      changePassword,
      requestPasswordReset,
      finishPasswordReset,
      updateMe,
      uploadAvatar,
      removeAvatar,
      refreshUser,
    }),
    [
      user,
      isLoading,
      login,
      verifyTfaCode,
      logout,
      changePassword,
      requestPasswordReset,
      finishPasswordReset,
      updateMe,
      uploadAvatar,
      removeAvatar,
      refreshUser,
    ]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
