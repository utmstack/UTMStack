import { useEffect, useLayoutEffect, useRef, useState, type RefObject } from 'react'

export interface FloatingPanelPosition {
  left: number
  top: number
}

/**
 * Positions and dismisses a fixed dropdown panel anchored to a trigger.
 * The panel is clamped to the viewport and flipped above the trigger when
 * there is not enough room below. It closes on outside mousedown, Escape,
 * scroll, or resize. Returns `null` until the first layout measurement, so
 * callers render a trigger-rect fallback on the initial paint.
 */
export function useFloatingPanel(
  open: boolean,
  onClose: () => void,
  triggerRef: RefObject<HTMLElement | null>,
  panelRef: RefObject<HTMLElement | null>,
  align: 'start' | 'end' = 'start',
): FloatingPanelPosition | null {
  const [position, setPosition] = useState<FloatingPanelPosition | null>(null)
  const onCloseRef = useRef(onClose)
  useEffect(() => {
    onCloseRef.current = onClose
  })

  useLayoutEffect(() => {
    if (!open) {
      setPosition(null)
      return
    }
    const menu = panelRef.current
    const trigger = triggerRef.current
    if (!menu || !trigger) return
    const rect = trigger.getBoundingClientRect()
    const left = align === 'end'
      ? Math.max(8, Math.min(rect.right - menu.offsetWidth, window.innerWidth - menu.offsetWidth - 8))
      : Math.max(8, Math.min(rect.left, window.innerWidth - menu.offsetWidth - 8))
    let top = rect.bottom + 4
    if (top + menu.offsetHeight > window.innerHeight - 8) top = rect.top - 4 - menu.offsetHeight
    setPosition({ left, top: Math.max(8, top) })
  }, [open, align, panelRef, triggerRef])

  useEffect(() => {
    if (!open) return
    const close = () => {
      setPosition(null)
      onCloseRef.current()
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') close()
    }
    const onMouseDown = (event: MouseEvent) => {
      const target = event.target as Node
      if (!triggerRef.current?.contains(target) && !panelRef.current?.contains(target)) close()
    }
    const onScroll = (event: Event) => {
      // Ignore scrolling inside the panel itself (it may have its own overflow).
      if (panelRef.current?.contains(event.target as Node)) return
      close()
    }
    document.addEventListener('mousedown', onMouseDown)
    document.addEventListener('keydown', onKeyDown)
    window.addEventListener('scroll', onScroll, true)
    window.addEventListener('resize', close)
    return () => {
      document.removeEventListener('mousedown', onMouseDown)
      document.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('scroll', onScroll, true)
      window.removeEventListener('resize', close)
    }
  }, [open, panelRef, triggerRef])

  return position
}
