import { createContext, useContext, type ReactNode } from 'react'
import { useTheme } from '@/shared/hooks/useTheme'

type Theme = 'dark' | 'light'
type ThemeMode = 'dark' | 'light' | 'system'

interface ThemeContextValue {
  theme: Theme
  mode: ThemeMode
  setTheme: (theme: Theme) => void
  setMode: (mode: ThemeMode) => void
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({ children }: { children: ReactNode }) {
  const value = useTheme()
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useThemeContext() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useThemeContext must be used within ThemeProvider')
  return ctx
}
