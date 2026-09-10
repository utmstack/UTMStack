import { useCallback, useEffect, useRef, useState } from 'react'
import type { MouseEvent as ReactMouseEvent } from 'react'

export type ColSize = string | number

interface Opts {
  min?: number | number[]
  storageKey?: string
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
  const [widths, setWidths] = useState<ColSize[]>(initial)
  const dragRef = useRef<{ index: number; startX: number; startW: number; measured: number[] } | null>(null)

  useEffect(() => {
    setWidths(initial)
  }, [initial.length, storageKey])

  useEffect(() => {
    const onMove = (e: globalThis.MouseEvent) => {
      const d = dragRef.current
      if (!d) return
      e.preventDefault()
      const w = Math.max(minAt(min, d.index), d.startW + (e.clientX - d.startX))
      setWidths((prev) => {
        const next = d.measured.length === prev.length ? d.measured.slice() : prev.slice()
        next[d.index] = w
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
    (index: number) => (e: ReactMouseEvent<HTMLElement>) => {
      e.preventDefault()
      e.stopPropagation()
      const handle = e.currentTarget as HTMLElement
      const cell = handle.closest('[data-resizable-col]') as HTMLElement | null ?? handle.parentElement
      const row = cell?.parentElement
      const measured = row
        ? Array.from(row.children)
            .filter((child): child is HTMLElement => child instanceof HTMLElement && child.hasAttribute('data-resizable-col'))
            .map((child, i) => Math.max(minAt(min, i), child.getBoundingClientRect().width))
        : []
      const startW = measured[index] ?? cell?.getBoundingClientRect().width ?? minAt(min, index)
      if (measured.length > 0) setWidths(measured)
      dragRef.current = { index, startX: e.clientX, startW, measured }
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    },
    [min],
  )

  const mins = widths.map((_, i) => minAt(min, i))
  // ponytail: wrap simple fr tracks in minmax so a narrow viewport can't
  // collapse labels below their per-column min. Fixed px tracks and strings
  // that already declare their own function (e.g. `minmax(120px, 1fr)`) are
  // passed through unchanged.
  const template = widths
    .map((w, i) => {
      if (typeof w === 'number') return `${w}px`
      return /^[\d.]+fr$/.test(w.trim()) ? `minmax(${mins[i]}px, ${w})` : w
    })
    .join(' ')

  return { widths, template, setWidths, startDrag, mins }
}
