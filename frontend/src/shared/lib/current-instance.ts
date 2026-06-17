import { useSyncExternalStore } from 'react'

/**
 * The federation "current instance" selection. Lives in a plain module (not React
 * state) because the api-client request interceptor — which runs outside React —
 * must read it to stamp the `X-UTM-Instance` header on every proxied call. A
 * React hook mirrors it for components.
 */
const STORAGE_KEY = 'utmstack_fs_instance'

function read(): number | null {
  if (typeof localStorage === 'undefined') return null
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  const n = Number(raw)
  return Number.isFinite(n) ? n : null
}

let current: number | null = read()
const listeners = new Set<() => void>()

export function getCurrentInstanceId(): number | null {
  return current
}

export function setCurrentInstanceId(id: number | null): void {
  current = id
  if (typeof localStorage !== 'undefined') {
    if (id == null) localStorage.removeItem(STORAGE_KEY)
    else localStorage.setItem(STORAGE_KEY, String(id))
  }
  listeners.forEach((l) => l())
}

function subscribe(fn: () => void): () => void {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

/** React binding for the current instance id (re-renders on switch). */
export function useCurrentInstanceId(): number | null {
  return useSyncExternalStore(subscribe, getCurrentInstanceId, () => null)
}
