import { useRef, type MouseEvent } from 'react'

/**
 * Backdrop click-to-close handlers for a modal overlay. Unlike a plain
 * `onClick={onClose}` on the backdrop, these don't fire when the user starts
 * a text selection drag inside the modal and releases the mouse outside it.
 *
 * That case would otherwise close the modal because the browser's synthetic
 * "click" event targets the nearest common ancestor of the mousedown and
 * mouseup targets — once the drag crosses the modal's boundary, that ancestor
 * is the backdrop itself, indistinguishable from a genuine backdrop click.
 * Requiring the mousedown to *also* have started directly on the backdrop
 * rules that out while still closing on an actual click there.
 *
 * Spread the result onto the backdrop element; no stopPropagation is needed
 * on the modal card itself.
 */
export function useBackdropDismiss(onClose: () => void) {
  const mouseDownOnBackdrop = useRef(false)

  return {
    onMouseDown: (e: MouseEvent) => {
      mouseDownOnBackdrop.current = e.target === e.currentTarget
    },
    onMouseUp: (e: MouseEvent) => {
      if (mouseDownOnBackdrop.current && e.target === e.currentTarget) onClose()
      mouseDownOnBackdrop.current = false
    },
  }
}
