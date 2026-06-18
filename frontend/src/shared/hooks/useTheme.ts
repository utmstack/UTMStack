import { useEffect, useState } from 'react'

type Theme = 'dark' | 'light'
type ThemeMode = 'dark' | 'light' | 'system'

const MODE_KEY = 'utmstack-theme-mode'
const LEGACY_KEY = 'utmstack-theme' // keep for migration

function readSystem(): Theme {
  if (typeof window === 'undefined') return 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function readInitialMode(): ThemeMode {
  if (typeof window === 'undefined') return 'system'
  // Migration: prior versions stored 'dark'|'light' under utmstack-theme
  const stored = localStorage.getItem(MODE_KEY) as ThemeMode | null
  if (stored === 'dark' || stored === 'light' || stored === 'system') return stored
  const legacy = localStorage.getItem(LEGACY_KEY) as Theme | null
  if (legacy === 'dark' || legacy === 'light') return legacy
  return 'system'
}

export function useTheme() {
  const [mode, setModeState] = useState<ThemeMode>(readInitialMode)
  const [systemTheme, setSystemTheme] = useState<Theme>(readSystem)

  // Watch OS-level changes to dark/light preference
  useEffect(() => {
    if (typeof window === 'undefined') return
    const mql = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = (e: MediaQueryListEvent) => setSystemTheme(e.matches ? 'dark' : 'light')
    mql.addEventListener('change', handler)
    return () => mql.removeEventListener('change', handler)
  }, [])

  const theme: Theme = mode === 'system' ? systemTheme : mode

  // Apply effective theme to <html>
  useEffect(() => {
    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
      root.classList.remove('light')
    } else {
      root.classList.add('light')
      root.classList.remove('dark')
    }
  }, [theme])

  const setMode = (m: ThemeMode) => {
    setModeState(m)
    if (typeof window !== 'undefined') localStorage.setItem(MODE_KEY, m)
  }

  // Topbar 2-way toggle stays as-is: flips between light/dark and pins.
  const toggleTheme = () => setMode(theme === 'dark' ? 'light' : 'dark')

  return { theme, mode, setMode, setTheme: (t: Theme) => setMode(t), toggleTheme }
}
