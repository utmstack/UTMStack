import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useAuth } from '@/features/auth'
import { presetRange } from '@/shared/components/ui/time-range-picker'
import type { LogExplorerTabConfig } from '../types/log-explorer.types'

const TABS_KEY = (uid: number | string) => `log-explorer:tabs:${uid}`
const ACTIVE_KEY = (uid: number | string) => `log-explorer:active-tab:${uid}`

function newId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `tab-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
}

function defaultTab(overrides: Partial<LogExplorerTabConfig> = {}): LogExplorerTabConfig {
  return {
    id: newId(),
    name: 'New tab',
    dataset: 'logs',
    patternStr: null,
    range: presetRange('24h'),
    filters: [],
    columns: [],
    searchInput: '',
    appliedQuery: '',
    sqlMode: false,
    sqlInput: '',
    appliedSql: '',
    viewMode: 'table',
    ...overrides,
  }
}

function readTabs(uid: number | string): { tabs: LogExplorerTabConfig[]; activeId: string } {
  try {
    const rawTabs = localStorage.getItem(TABS_KEY(uid))
    const rawActive = localStorage.getItem(ACTIVE_KEY(uid))
    const tabs = rawTabs ? (JSON.parse(rawTabs) as LogExplorerTabConfig[]) : []
    if (!Array.isArray(tabs) || tabs.length === 0) {
      const seed = defaultTab()
      return { tabs: [seed], activeId: seed.id }
    }
    const activeId = rawActive && tabs.some((t) => t.id === rawActive) ? rawActive : tabs[0].id
    return { tabs, activeId }
  } catch {
    const seed = defaultTab()
    return { tabs: [seed], activeId: seed.id }
  }
}

export interface UseLogExplorerTabs {
  tabs: LogExplorerTabConfig[]
  activeId: string
  activeTab: LogExplorerTabConfig
  setActive: (id: string) => void
  addTab: (config?: Partial<LogExplorerTabConfig>) => string
  removeTab: (id: string) => void
  renameTab: (id: string, name: string) => void
  updateActive: (patch: Partial<LogExplorerTabConfig>) => void
}

export function useLogExplorerTabs(): UseLogExplorerTabs {
  const { user } = useAuth()
  const uid = user?.id ?? 'anon'

  const [tabs, setTabs] = useState<LogExplorerTabConfig[]>(() => readTabs(uid).tabs)
  const [activeId, setActiveId] = useState<string>(() => readTabs(uid).activeId)
  const lastUidRef = useRef(uid)

  // Re-hydrate when the user changes (sign-out / sign-in as different user).
  useEffect(() => {
    if (lastUidRef.current === uid) return
    lastUidRef.current = uid
    const next = readTabs(uid)
    setTabs(next.tabs)
    setActiveId(next.activeId)
  }, [uid])

  // Write-through persistence on every change.
  useEffect(() => {
    try {
      localStorage.setItem(TABS_KEY(uid), JSON.stringify(tabs))
    } catch {
      /* quota or disabled storage — silently skip */
    }
  }, [uid, tabs])

  useEffect(() => {
    try {
      localStorage.setItem(ACTIVE_KEY(uid), activeId)
    } catch {
      /* ignore */
    }
  }, [uid, activeId])

  const setActive = useCallback((id: string) => setActiveId(id), [])

  const addTab = useCallback((config: Partial<LogExplorerTabConfig> = {}) => {
    const tab = defaultTab(config)
    setTabs((cur) => [...cur, tab])
    setActiveId(tab.id)
    return tab.id
  }, [])

  const removeTab = useCallback((id: string) => {
    setTabs((cur) => {
      const idx = cur.findIndex((t) => t.id === id)
      if (idx < 0) return cur
      const next = cur.filter((t) => t.id !== id)
      if (next.length === 0) {
        const seed = defaultTab()
        setActiveId(seed.id)
        return [seed]
      }
      setActiveId((curActive) => {
        if (curActive !== id) return curActive
        const neighbour = next[Math.max(0, idx - 1)] ?? next[0]
        return neighbour.id
      })
      return next
    })
  }, [])

  const renameTab = useCallback((id: string, name: string) => {
    const trimmed = name.trim() || 'Untitled'
    setTabs((cur) => cur.map((t) => (t.id === id ? { ...t, name: trimmed } : t)))
  }, [])

  const updateActive = useCallback(
    (patch: Partial<LogExplorerTabConfig>) => {
      setTabs((cur) => cur.map((t) => (t.id === activeId ? { ...t, ...patch, id: t.id } : t)))
    },
    [activeId]
  )

  const activeTab = useMemo(
    () => tabs.find((t) => t.id === activeId) ?? tabs[0],
    [tabs, activeId]
  )

  return { tabs, activeId, activeTab, setActive, addTab, removeTab, renameTab, updateActive }
}
