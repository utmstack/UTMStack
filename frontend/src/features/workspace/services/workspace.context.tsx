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
import { workspaceHttpService } from './workspace-http.service'
import type { Workspace } from '../types/workspace.types'

const STORAGE_KEY = 'utmstack_current_workspace'

interface WorkspaceContextValue {
  workspaces: Workspace[]
  current: Workspace | null
  isLoading: boolean
  setCurrent: (id: number) => void
  refresh: () => Promise<void>
}

const WorkspaceContext = createContext<WorkspaceContextValue | undefined>(undefined)

function readStoredId(): number | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const n = Number(raw)
    return Number.isFinite(n) && n > 0 ? n : null
  } catch {
    return null
  }
}

function writeStoredId(id: number | null) {
  if (typeof window === 'undefined') return
  try {
    if (id == null) localStorage.removeItem(STORAGE_KEY)
    else localStorage.setItem(STORAGE_KEY, String(id))
  } catch {
    // localStorage unavailable — silently fall back to in-memory only
  }
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [currentId, setCurrentId] = useState<number | null>(() => readStoredId())
  const [isLoading, setIsLoading] = useState(false)

  const refresh = useCallback(async () => {
    setIsLoading(true)
    try {
      const rows = await workspaceHttpService.list()
      setWorkspaces(rows)
    } catch {
      // intentionally silent — the workspace selector is non-critical chrome.
      // The /settings/workspaces page surfaces fetch errors with toasts.
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated) {
      setWorkspaces([])
      return
    }
    refresh()
  }, [isAuthenticated, refresh])

  // Resolve the active workspace: persisted id → default → first row.
  const current = useMemo<Workspace | null>(() => {
    if (workspaces.length === 0) return null
    if (currentId != null) {
      const found = workspaces.find((w) => w.id === currentId)
      if (found) return found
    }
    return workspaces.find((w) => w.is_default) ?? workspaces[0]
  }, [workspaces, currentId])

  // If the persisted id is gone (workspace deleted) and we resolved to a
  // fallback, re-anchor localStorage so a refresh sticks.
  useEffect(() => {
    if (current && current.id !== currentId) {
      writeStoredId(current.id)
      setCurrentId(current.id)
    }
  }, [current, currentId])

  const setCurrent = useCallback((id: number) => {
    setCurrentId(id)
    writeStoredId(id)
  }, [])

  const value = useMemo<WorkspaceContextValue>(
    () => ({ workspaces, current, isLoading, setCurrent, refresh }),
    [workspaces, current, isLoading, setCurrent, refresh],
  )

  return <WorkspaceContext.Provider value={value}>{children}</WorkspaceContext.Provider>
}

export function useWorkspaces(): WorkspaceContextValue {
  const ctx = useContext(WorkspaceContext)
  if (!ctx) throw new Error('useWorkspaces must be used within WorkspaceProvider')
  return ctx
}
