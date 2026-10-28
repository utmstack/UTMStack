import { useRef, type ReactNode, type RefObject } from 'react'
import { createPortal } from 'react-dom'
import { useFloatingPanel } from '@/shared/hooks/use-floating-panel'

/**
 * Fixed dropdown panel portaled to `document.body`, positioned next to its
 * trigger by {@link useFloatingPanel} (viewport clamping, flip above when
 * needed) and closed on outside click, Escape, scroll, or resize. The root
 * stops click propagation so portaled content does not bubble into whatever
 * row or cell hosts the trigger.
 */
export function FloatingPanel({
  open,
  onClose,
  triggerRef,
  align = 'start',
  className = 'z-[95] rounded-md border border-border bg-popover py-1 shadow-lg',
  children,
}: {
  open: boolean
  onClose: () => void
  triggerRef: RefObject<HTMLElement | null>
  align?: 'start' | 'end'
  className?: string
  children: ReactNode
}) {
  const panelRef = useRef<HTMLDivElement>(null)
  const position = useFloatingPanel(open, onClose, triggerRef, panelRef, align)

  if (!open) return null
  const triggerRect = triggerRef.current?.getBoundingClientRect()
  const fallback = triggerRect
    ? { left: Math.max(8, triggerRect.left), top: Math.max(8, triggerRect.bottom + 4) }
    : { left: 8, top: 8 }

  return createPortal(
    <div
      ref={panelRef}
      data-floating-panel
      style={{ position: 'fixed', left: position?.left ?? fallback.left, top: position?.top ?? fallback.top }}
      onClick={(event) => event.stopPropagation()}
      className={className}
    >
      {children}
    </div>,
    document.body,
  )
}
