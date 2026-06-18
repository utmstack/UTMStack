import { ReactNode, useEffect } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TooltipProvider } from '@/shared/components/ui/tooltip'
import { Toaster } from 'sonner'
import { AuthProvider } from '@/features/auth'
import { FederationAuthProvider } from '@/features/federation'
import { IS_FEDERATION } from '@/shared/config/mode'
import { BillingProvider } from '@/features/billing'
import { BrandingProvider } from '@/features/branding'
import { NotificationsProvider } from '@/features/notifications'
import { createApiClient } from '@/shared/lib/api-client'
import { setDateFormatPrefs, type DateFormatPrefs } from '@/shared/lib/datetime'
import { ThemeProvider, useThemeContext } from './ThemeProvider'

const queryClient = new QueryClient()

// Lives inside ThemeProvider so sonner follows the app's dark/light theme instead
// of defaulting to light.
function ThemedToaster() {
  const { theme } = useThemeContext()
  return <Toaster richColors position="top-right" closeButton theme={theme} />
}

// Loads the org-wide timestamp display preference once so the app renders dates
// in the configured timezone/format. Public endpoint — works pre-login too.
function DateFormatLoader() {
  useEffect(() => {
    createApiClient()
      .get<DateFormatPrefs>('/date-format')
      .then(setDateFormatPrefs)
      .catch(() => {
        /* keep UTC/medium defaults when unreachable */
      })
  }, [])
  return null
}

// Federation mode swaps the instance login/session for an FS-backed one, but
// populates the same AuthContext so the rest of the shell is unchanged.
const AuthShell = IS_FEDERATION ? FederationAuthProvider : AuthProvider

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
    <ThemeProvider>
      <BrandingProvider>
        <TooltipProvider delayDuration={0}>
          <AuthShell>
            <BillingProvider>
              <NotificationsProvider>
                <DateFormatLoader />
                {children}
                <ThemedToaster />
              </NotificationsProvider>
            </BillingProvider>
          </AuthShell>
        </TooltipProvider>
      </BrandingProvider>
    </ThemeProvider>
    </QueryClientProvider>
  )
}

export { useThemeContext } from './ThemeProvider'
