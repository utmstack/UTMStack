import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useAuth } from '@/features/auth'
import { storageHttpService } from './storage-http.service'

interface StorageEnabledContextValue {
  enabled: boolean
  isLoading: boolean
}

const StorageEnabledContext = createContext<StorageEnabledContextValue | undefined>(undefined)

/**
 * Asks /storage/enabled once the user is authenticated so the session headers
 * exist, and exposes the answer app-wide: the settings sidebar hides the
 * data-retention entry and the page itself bounces to /home when it is off.
 */
export function StorageEnabledProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const [enabled, setEnabled] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const refresh = useCallback(async () => {
    setIsLoading(true)
    try {
      const res = await storageHttpService.retention_enabled()
      setEnabled(res.enabled)
    } catch {
      setEnabled(false)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated) {
      setEnabled(false)
      return
    }
    void refresh()
  }, [isAuthenticated, refresh])

  const value = useMemo<StorageEnabledContextValue>(
    () => ({ enabled, isLoading }),
    [enabled, isLoading]
  )

  return <StorageEnabledContext.Provider value={value}>{children}</StorageEnabledContext.Provider>
}

export function useStorageEnabled(): StorageEnabledContextValue {
  const ctx = useContext(StorageEnabledContext)
  if (!ctx) throw new Error('useStorageEnabled must be used within StorageEnabledProvider')
  return ctx
}
