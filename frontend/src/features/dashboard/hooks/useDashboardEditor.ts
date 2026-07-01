import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
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
  // layout rows arrive asynchronously. We must NOT depend on the `initialItems`
  // array identity: callers usually pass a freshly-derived array every render,
  // so depending on it would re-run this effect (which setStates) on every
  // render → infinite loop. Instead we sync only when the *content* changes,
  // tracked by a stable string signature, and read the latest array via a ref.
  const itemsRef = useRef(initialItems)
  itemsRef.current = initialItems
  const pendingRemovalsRef = useRef(pendingRemovals)
  pendingRemovalsRef.current = pendingRemovals

  const signature = useMemo(
    () => JSON.stringify(initialItems.map((it) => [it.i, it.x, it.y, it.w, it.h])),
    [initialItems],
  )

  useEffect(() => {
    const items = itemsRef.current
    if (!editing) {
      setWorking(items)
      setBaseline(items)
      setPendingRemovals([])
      return
    }
    // While editing, pick up items that were created externally (e.g. via
    // "Add widget"): anything present in the fresh layout rows but neither in
    // working nor pending-removal gets appended so the grid reflects it live.
    const removedIds = new Set(pendingRemovalsRef.current.map(String))
    setWorking((curr) => {
      const known = new Set(curr.map((c) => c.i))
      const extras = items.filter((it) => !known.has(it.i) && !removedIds.has(it.i))
      return extras.length ? [...curr, ...extras] : curr
    })
    setBaseline((curr) => {
      const known = new Set(curr.map((c) => c.i))
      const extras = items.filter((it) => !known.has(it.i) && !removedIds.has(it.i))
      return extras.length ? [...curr, ...extras] : curr
    })
    // `signature` (a value-based string) stands in for `initialItems`; refs
    // are intentionally not dependencies.
  }, [editing, signature])

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
