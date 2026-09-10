import { useCallback, useEffect, useRef, useState } from 'react'
import type { MouseEvent as ReactMouseEvent } from 'react'

export type ColSize = string | number

interface Opts {
  min?: number
  storageKey?: string
}

export function useResizableColumns(initial: ColSize[], opts: Opts = {}) {
  const min = opts.min ?? 40
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
      const w = Math.max(min, d.startW + (e.clientX - d.startX))
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
            .map((child) => Math.max(min, child.getBoundingClientRect().width))
        : []
      const startW = measured[index] ?? cell?.getBoundingClientRect().width ?? min
      if (measured.length > 0) setWidths(measured)
      dragRef.current = { index, startX: e.clientX, startW, measured }
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    },
    [min],
  )

  const template = widths.map((w) => (typeof w === 'number' ? `${w}px` : w)).join(' ')

  return { widths, template, setWidths, startDrag }
}
