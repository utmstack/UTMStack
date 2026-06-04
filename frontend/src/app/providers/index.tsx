import { ReactNode } from 'react'
import { TooltipProvider } from '@/shared/components/ui/tooltip'
import { Toaster } from 'sonner'
import { AuthProvider } from '@/features/auth'
import { WorkspaceProvider } from '@/features/workspace'
import { ThemeProvider } from './ThemeProvider'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider>
      <TooltipProvider delayDuration={0}>
        <AuthProvider>
          <WorkspaceProvider>
            {children}
            <Toaster richColors position="top-right" closeButton />
          </WorkspaceProvider>
        </AuthProvider>
      </TooltipProvider>
    </ThemeProvider>
  )
}

export { useThemeContext } from './ThemeProvider'
