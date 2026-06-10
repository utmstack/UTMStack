import { ReactNode } from 'react'
import { TooltipProvider } from '@/shared/components/ui/tooltip'
import { Toaster } from 'sonner'
import { AuthProvider } from '@/features/auth'
import { BillingProvider } from '@/features/billing'
import { NotificationsProvider } from '@/features/notifications'
import { ThemeProvider } from './ThemeProvider'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider>
      <TooltipProvider delayDuration={0}>
        <AuthProvider>
          <BillingProvider>
            <NotificationsProvider>
              {children}
              <Toaster richColors position="top-right" closeButton />
            </NotificationsProvider>
          </BillingProvider>
        </AuthProvider>
      </TooltipProvider>
    </ThemeProvider>
  )
}

export { useThemeContext } from './ThemeProvider'
