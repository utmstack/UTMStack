import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { clearTokens, getStoredTokens, setLogoutHandler, storeTokens } from '@/shared/lib/api-client'
import { setCurrentInstanceId } from '@/shared/lib/current-instance'
import i18n from '@/shared/i18n'
import {
  AuthContext,
  type AuthContextValue,
  type LoginOutcome,
} from '@/features/auth/services/auth.context'
import type { LoginRequest, User } from '@/features/auth/types/auth.types'
import { federationAuthService } from './services/federation-auth.service'
import type { FederationUser } from './types'

// Map the FS user onto the instance User shape so the shared shell (Topbar,
// ProtectedRoute, etc.) works unchanged.
function toUser(fu: FederationUser): User {
  return {
    id: fu.id,
    login: fu.login,
    email: fu.email,
    first_name: fu.firstName,
    last_name: fu.lastName,
    image_url: fu.imageUrl,
    lang_key: fu.langKey,
    activated: true,
    tfa_enabled: fu.tfaEnabled,
    tfa_method: fu.tfaMethod,
  }
}

// Stash the FS token pair from a (possibly 2FA-gated) login response.
function storeFromResponse(res: {
  accessToken?: string
  refreshToken?: string
  expiresAt?: string
  refreshExpiresAt?: string
}) {
  storeTokens({
    access_token: res.accessToken ?? '',
    refresh_token: res.refreshToken ?? '',
    expires_at: res.expiresAt,
    refresh_expires_at: res.refreshExpiresAt,
  })
}

/**
 * Federation-mode replacement for AuthProvider. It populates the SAME AuthContext
 * — so `useAuth()` works everywhere — but is backed by the FS's own session
 * (/api/v1/auth), not an instance login. Authz is permissive (the instance API key
 * enforces every permission server-side; `/auth/me` isn't reachable through the
 * proxy). Profile features (avatar, password, sessions, TOTP 2FA) hit the FS
 * directly.
 */
export function FederationAuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(() => getStoredTokens() !== null)

  const reset = useCallback(() => {
    clearTokens()
    setCurrentInstanceId(null)
    setUser(null)
  }, [])

  const login = useCallback(async (input: LoginRequest): Promise<LoginOutcome> => {
    const res = await federationAuthService.login(input.login, input.password)
    if (res.tfaRequired && res.preAuthToken) {
      return {
        status: 'tfa_required',
        method: res.tfaMethod ?? 'TOTP',
        preAuthToken: res.preAuthToken,
      }
    }
    storeFromResponse(res)
    setUser(toUser(res.user))
    return { status: 'authenticated' }
  }, [])

  const verifyTfaCode = useCallback(async (preAuthToken: string, code: string) => {
    const res = await federationAuthService.verifyTfaCode(preAuthToken, code)
    storeFromResponse(res)
    setUser(toUser(res.user))
  }, [])

  const logout = useCallback(async () => {
    reset()
  }, [reset])

  const refreshUser = useCallback(async () => {
    const fu = await federationAuthService.me()
    setUser(toUser(fu))
  }, [])

  // Hydrate from a stored FS token on mount.
  useEffect(() => {
    const tokens = getStoredTokens()
    if (!tokens?.access_token) {
      setIsLoading(false)
      return
    }
    let cancelled = false
    federationAuthService
      .me()
      .then((fu) => {
        if (!cancelled) setUser(toUser(fu))
      })
      .catch(() => {
        if (!cancelled) reset()
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [reset])

  // api-client logout fallback (fired when a request 401s in FS mode).
  useEffect(() => {
    setLogoutHandler(() => setUser(null))
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: user !== null,
      isLoading,
      roles: ['FS_ADMIN'],
      permissions: [],
      isAdmin: true,
      hasPermission: () => true,
      login,
      verifyTfaCode,
      logout,
      changePassword: async (input) => {
        await federationAuthService.changePassword(input.current_password, input.new_password)
      },
      requestPasswordReset: async (email) => {
        await federationAuthService.requestPasswordReset(email)
      },
      finishPasswordReset: async (input) => {
        await federationAuthService.finishPasswordReset(input.key, input.new_password)
      },
      updateMe: async (input) => {
        const fu = await federationAuthService.updateMe({
          firstName: input.first_name,
          lastName: input.last_name,
          email: input.email,
          langKey: input.lang_key,
        })
        setUser(toUser(fu))
        if (input.lang_key) void i18n.changeLanguage(input.lang_key)
      },
      uploadAvatar: async (file) => {
        setUser(toUser(await federationAuthService.uploadAvatar(file)))
      },
      removeAvatar: async () => {
        setUser(toUser(await federationAuthService.removeAvatar()))
      },
      refreshUser,
    }),
    [user, isLoading, login, verifyTfaCode, logout, refreshUser],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
