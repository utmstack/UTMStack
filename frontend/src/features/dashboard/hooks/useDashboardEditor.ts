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

  useEffect(() => {
    if (!editing) {
      setWorking(initialItems)
      setBaseline(initialItems)
      setPendingRemovals([])
    }
  }, [editing])

  const enter = useCallback(() => {
    setBaseline(working)
    setEditing(true)
  }, [working])

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
  }
}
