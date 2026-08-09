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
// ProtectedRoute, etc.) works unchanged. The FS still models a person as a
// login plus a first/last name; the instance moved to one display name, so the
// two are joined here and split again on the way back out.
function toUser(fu: FederationUser): User {
  const name = [fu.firstName, fu.lastName].filter(Boolean).join(' ')
  return {
    id: String(fu.id),
    email: fu.email ?? '',
    name: name || fu.login,
    status: 'active',
    image_url: fu.imageUrl,
    lang_key: fu.langKey,
    // The FS authenticates against its own user table and offers a password
    // change, so the account is never federated in the instance's sense.
    federated: false,
    tfa_enabled: fu.tfaEnabled ?? false,
  }
}

// splitName undoes the join above. Everything after the first space is the last
// name, so a multi-word surname survives the round trip.
function splitName(name?: string): { firstName?: string; lastName?: string } {
  if (!name) return {}
  const at = name.indexOf(' ')
  if (at < 0) return { firstName: name }
  return { firstName: name.slice(0, at), lastName: name.slice(at + 1) }
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
        method: res.tfaMethod ?? 'totp',
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
      // An FS admin administers instances, not the platform inside one. False
      // is also the safe direction: the platform-only routes stay closed rather
      // than opening on a guess.
      isPlatformAdmin: false,
      // An FS session belongs to no tenant — it is outside any single instance.
      tenantId: undefined,
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
          ...splitName(input.name),
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
      // adoptSession takes the token an instance SSO round trip hands back.
      // The FS has its own login and never makes that round trip, so there is
      // no session to adopt here.
      adoptSession: async () => {},
      refreshUser,
    }),
    [user, isLoading, login, verifyTfaCode, logout, refreshUser],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
