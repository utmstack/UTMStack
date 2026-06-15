import { useCallback, useEffect, useState } from 'react'
import type { GridLayoutItem } from '@/features/dashboard/types'

export interface EditorState {
  editing: boolean
  working: GridLayoutItem[]
  dirty: boolean
  pendingRemovals: number[]
}

export function useDashboardEditor(initialItems: GridLayoutItem[]) {
  const [editing, setEditing] = useState(false)
  const [working, setWorking] = useState<GridLayoutItem[]>(initialItems)
  const [baseline, setBaseline] = useState<GridLayoutItem[]>(initialItems)
  const [pendingRemovals, setPendingRemovals] = useState<number[]>([])

  // Keep working/baseline in sync with the latest data while NOT editing — the
  // layout rows arrive asynchronously, so without `initialItems` in the deps the
  // working copy stays stale (empty) and edit mode would render nothing.
  useEffect(() => {
    if (!editing) {
      setWorking(initialItems)
      setBaseline(initialItems)
      setPendingRemovals([])
    }
  }, [editing, initialItems])

  const enter = useCallback(() => {
    // Start the edit session from the current items (not a stale working copy).
    setWorking(initialItems)
    setBaseline(initialItems)
    setPendingRemovals([])
    setEditing(true)
  }, [initialItems])

  const discard = useCallback(() => {
    setWorking(baseline)
    setPendingRemovals([])
    setEditing(false)
  }, [baseline])

  const commit = useCallback(() => {
    setBaseline(working)
    setPendingRemovals([])
    setEditing(false)
  }, [working])

  const replace = useCallback((items: GridLayoutItem[]) => {
    setWorking(items)
  }, [])

  const remove = useCallback((id: number) => {
    setWorking((curr) => curr.filter((it) => it.i !== String(id)))
    setPendingRemovals((curr) => (curr.includes(id) ? curr : [...curr, id]))
  }, [])

  // Set a widget's size preset (w = column span 1–3, h = height 1–2).
  const resize = useCallback((id: string, w: number, h: number) => {
    setWorking((curr) => curr.map((it) => (it.i === id ? { ...it, w, h } : it)))
  }, [])

  // Move a widget one slot back (-1) or forward (+1) in the display order.
  const move = useCallback((id: string, dir: -1 | 1) => {
    setWorking((curr) => {
      const idx = curr.findIndex((it) => it.i === id)
      const next = idx + dir
      if (idx < 0 || next < 0 || next >= curr.length) return curr
      const copy = [...curr]
      ;[copy[idx], copy[next]] = [copy[next], copy[idx]]
      return copy
    })
  }, [])

  const dirty =
    pendingRemovals.length > 0 ||
    working.length !== baseline.length ||
    working.some((it, idx) => {
      const b = baseline[idx]
      return !b || b.i !== it.i || b.x !== it.x || b.y !== it.y || b.w !== it.w || b.h !== it.h
    })

  return {
    editing,
    working,
    baseline,
    dirty,
    pendingRemovals,
    enter,
    discard,
    commit,
    replace,
    remove,
    move,
    resize,
  }
}
