import { useCallback, useEffect, useRef, useState } from 'react'
import type { MouseEvent as ReactMouseEvent } from 'react'

export type ColSize = string | number

interface Opts {
  min?: number | number[]
  storageKey?: string
  // Column index that absorbs the table's leftover width instead of taking an
  // explicit px of its own. Without one, every column has a fixed width and a
  // table forced wider than their sum (e.g. by w-full/min-w-full on a screen
  // wider than the columns need) has nothing to give the extra space to —
  // table-layout:fixed then scales every declared width up proportionally to
  // fill it, so dragging one column visibly resizes all the others too. With
  // a flex column, only it grows on a wide screen and only it shrinks to make
  // room when you widen a neighbor; every other column's own width is never
  // touched by anything but its own handle.
  flex?: number
}

const DEFAULT_MIN = 40

const minAt = (min: number | number[] | undefined, index: number): number => {
  if (min == null) return DEFAULT_MIN
  if (typeof min === 'number') return min
  return min[index] ?? DEFAULT_MIN
}

// Per-column min widths derived from the header label so a drag/resize can't
// crop the label. Pass a number in the labels array for icon/checkbox columns
// where no label exists — that number is used verbatim as the min.
// ponytail: 15px/char + 30px padding is a rough uppercase-heading approximation;
// tune the charPx/padding options if a specific font stack diverges.
export function colMins(
  labels: Array<string | number>,
  opts: { floor?: number; charPx?: number; padding?: number } = {},
): number[] {
  const { floor = 60, charPx = 15, padding = 30 } = opts
  return labels.map((l) =>
    typeof l === 'number' ? l : Math.max(floor, Math.round(l.length * charPx) + padding),
  )
}

export function useResizableColumns(initial: ColSize[], opts: Opts = {}) {
  const min = opts.min
  const storageKey = opts.storageKey
  const flex = opts.flex
  const [widths, setWidths] = useState<ColSize[]>(initial)
  const dragRef = useRef<{ target: number; sign: 1 | -1; startX: number; startY: number; startW: number; measured: number[]; moved: boolean } | null>(null)

  const rawWidths = widths.length === initial.length ? widths : initial
  // The flex column never carries an explicit width — [undefined] tells the
  // header components to leave it unset (table-layout:fixed / CSS Grid then
  // give it 100% of whatever's left over) instead of rendering a stale number
  // that was only ever a placeholder for array-length bookkeeping.
  const effectiveWidths: (ColSize | undefined)[] = rawWidths.map((w, i) => (i === flex ? undefined : w))

  useEffect(() => {
    setWidths(initial)
  }, [initial.length, storageKey])

  useEffect(() => {
    const onMove = (e: globalThis.MouseEvent) => {
      const d = dragRef.current
      if (!d) return
      if (!d.moved && Math.abs(e.clientX - d.startX) < 2 && Math.abs(e.clientY - d.startY) < 2) return
      d.moved = true
      e.preventDefault()
      const minWidth = minAt(min, d.target)
      const rawWidth = d.startW + d.sign * (e.clientX - d.startX)
      const w = Math.max(minWidth, rawWidth)
      setWidths((prev) => {
        const next = d.measured.length === prev.length ? d.measured.slice() : prev.slice()
        next[d.target] = w
        return next
      })
    }
    const onUp = () => {
      if (!dragRef.current) return
      dragRef.current = null
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [min])

  const startDrag = useCallback(
    // index is the boundary being dragged (it always resizes the column to
    // its left, matching the handle rendered at the START of the column to
    // its right — see resizable-table-header.tsx). If that left column is
    // the flex one, there's no explicit width to change on it, so the drag
    // instead shrinks/grows the column on the OTHER side of the boundary by
    // the same amount, sign-flipped: moving the boundary right still makes
    // the flex column visually wider by taking that space from its neighbor.
    (index: number) => (e: ReactMouseEvent<HTMLElement>) => {
      e.preventDefault()
      e.stopPropagation()
      const isFlexBoundary = index === flex
      const target = isFlexBoundary ? index + 1 : index
      const sign: 1 | -1 = isFlexBoundary ? -1 : 1
      const handle = e.currentTarget as HTMLElement
      const cell = handle.closest('[data-resizable-col]') as HTMLElement | null ?? handle.parentElement
      const row = cell?.parentElement
      const measured = row
        ? Array.from(row.children)
            .filter((child): child is HTMLElement => child instanceof HTMLElement && child.hasAttribute('data-resizable-col'))
            .map((child, i) => Math.max(minAt(min, i), child.getBoundingClientRect().width))
        : []
      const startW = Math.max(minAt(min, target), measured[target] ?? cell?.getBoundingClientRect().width ?? minAt(min, target))
      dragRef.current = { target, sign, startX: e.clientX, startY: e.clientY, startW, measured, moved: false }
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    },
    [min, flex],
  )

  const mins = effectiveWidths.map((_, i) => minAt(min, i))
  // ponytail: wrap simple fr tracks in minmax so a narrow viewport can't
  // collapse labels below their per-column min. Fixed px tracks and strings
  // that already declare their own function (e.g. `minmax(120px, 1fr)`) are
  // passed through unchanged. The flex column (undefined width) becomes a
  // real 1fr track here, for consumers built on CSS Grid.
  const template = effectiveWidths
    .map((w, i) => {
      if (w == null) return `minmax(${mins[i]}px, 1fr)`
      if (typeof w === 'number') return `${w}px`
      return /^[\d.]+fr$/.test(w.trim()) ? `minmax(${mins[i]}px, ${w})` : w
    })
    .join(' ')

  return { widths: effectiveWidths, template, setWidths, startDrag, mins }
}
