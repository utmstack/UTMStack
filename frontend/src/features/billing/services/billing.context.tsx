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
import { billingHttpService } from './billing-http.service'
import type { License, VersionInfo } from '../types/billing.types'

interface BillingContextValue {
  license: License | null
  version: VersionInfo | null
  isLoading: boolean
  error: boolean
  refresh: () => Promise<void>
}

const BillingContext = createContext<BillingContextValue | undefined>(undefined)

/**
 * Loads the license + version once the user is authenticated and exposes them
 * app-wide (Topbar badge, License page, enterprise-feature gating). Refreshing
 * after a license upload keeps every consumer in sync.
 */
export function BillingProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const [license, setLicense] = useState<License | null>(null)
  const [version, setVersion] = useState<VersionInfo | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState(false)

  const refresh = useCallback(async () => {
    setIsLoading(true)
    setError(false)
    try {
      const [lic, ver] = await Promise.all([
        billingHttpService.getLicense(),
        billingHttpService.getVersion(),
      ])
      setLicense(lic)
      setVersion(ver)
    } catch {
      setError(true)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated) {
      setLicense(null)
      setVersion(null)
      return
    }
    void refresh()
  }, [isAuthenticated, refresh])

  const value = useMemo<BillingContextValue>(
    () => ({ license, version, isLoading, error, refresh }),
    [license, version, isLoading, error, refresh]
  )

  return <BillingContext.Provider value={value}>{children}</BillingContext.Provider>
}

export function useBilling(): BillingContextValue {
  const ctx = useContext(BillingContext)
  if (!ctx) throw new Error('useBilling must be used within BillingProvider')
  return ctx
}
