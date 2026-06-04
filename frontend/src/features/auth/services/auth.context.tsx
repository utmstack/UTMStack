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
import { AuthHttpError, authHttpService } from './auth-http.service'
import type {
  ChangePasswordRequest,
  LoginRequest,
  UpdateMeRequest,
  User,
} from '../types/auth.types'

interface AuthContextValue {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (input: LoginRequest) => Promise<void>
  logout: () => Promise<void>
  changePassword: (input: ChangePasswordRequest) => Promise<void>
  updateMe: (input: UpdateMeRequest) => Promise<void>
  uploadAvatar: (file: File) => Promise<void>
  removeAvatar: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(() => getStoredTokens() !== null)

  const reset = useCallback(() => {
    clearTokens()
    setUser(null)
  }, [])

  const login = useCallback(async (input: LoginRequest) => {
    const result = await authHttpService.login(input)
    storeTokens({
      access_token: result.access_token,
      refresh_token: result.refresh_token,
      expires_at: result.expires_at,
      refresh_expires_at: result.refresh_expires_at,
    })
    setUser(result.user)
  }, [])

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

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: user !== null,
      isLoading,
      login,
      logout,
      changePassword,
      updateMe,
      uploadAvatar,
      removeAvatar,
    }),
    [user, isLoading, login, logout, changePassword, updateMe, uploadAvatar, removeAvatar]
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
