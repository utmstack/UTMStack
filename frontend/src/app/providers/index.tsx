import { ReactNode } from 'react'
import { TooltipProvider } from '@/shared/components/ui/tooltip'
import { Toaster } from 'sonner'
import { AuthProvider } from '@/features/auth'
import { BillingProvider } from '@/features/billing'
import { NotificationsProvider } from '@/features/notifications'
import { ThemeProvider, useThemeContext } from './ThemeProvider'

// Lives inside ThemeProvider so sonner follows the app's dark/light theme instead
// of defaulting to light.
function ThemedToaster() {
  const { theme } = useThemeContext()
  return <Toaster richColors position="top-right" closeButton theme={theme} />
}

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider>
      <TooltipProvider delayDuration={0}>
        <AuthProvider>
          <BillingProvider>
            <NotificationsProvider>
              {children}
              <ThemedToaster />
            </NotificationsProvider>
          </BillingProvider>
        </AuthProvider>
      </TooltipProvider>
    </ThemeProvider>
  )
}

export { useThemeContext } from './ThemeProvider'
