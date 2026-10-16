import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Minus, Square, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { CanvasPanel, PanelGeometry } from './types'

export interface CanvasHandle {
  onMove(id: string, geometry: Partial<PanelGeometry>): void
  onFocus(id: string): void
  onClose(id: string): void
  onToggleMinimise(id: string): void
}

export function CanvasPanelFrame({
  panel,
  onMove,
  onFocus,
  onClose,
  onToggleMinimise,
  children,
}: CanvasHandle & { panel: CanvasPanel; children: React.ReactNode }) {
  const { t } = useTranslation()
  const name = t('tiAnalyst.canvas.panelName', {
    title: panel.title,
    kind: t(`tiAnalyst.canvas.kind.${panel.kind}`),
  })
  const gesture = useRef<'drag' | 'resize' | null>(null)
  const [dragging, setDragging] = useState(false)
  const [resizing, setResizing] = useState(false)
  const origin = useRef({ pointerX: 0, pointerY: 0, x: 0, y: 0, width: 0, height: 0 })

  const beginDrag = useCallback(
    (e: React.PointerEvent) => {
      if (e.button !== 0) return
      try {
        e.currentTarget.setPointerCapture(e.pointerId)
      } catch {
      }
      origin.current = { pointerX: e.clientX, pointerY: e.clientY, x: panel.x, y: panel.y, width: panel.width, height: panel.height }
      gesture.current = 'drag'
      setDragging(true)
      onFocus(panel.id)
    },
    [panel, onFocus],
  )

  const beginResize = useCallback(
    (e: React.PointerEvent) => {
      if (e.button !== 0) return
      e.stopPropagation()
      try {
        e.currentTarget.setPointerCapture(e.pointerId)
      } catch {
      }
      origin.current = { pointerX: e.clientX, pointerY: e.clientY, x: panel.x, y: panel.y, width: panel.width, height: panel.height }
      gesture.current = 'resize'
      setResizing(true)
      onFocus(panel.id)
    },
    [panel, onFocus],
  )

  const onPointerMove = useCallback(
    (e: React.PointerEvent) => {
      const active = gesture.current
      if (!active) return
      const dx = e.clientX - origin.current.pointerX
      const dy = e.clientY - origin.current.pointerY
      if (active === 'drag') {
        onMove(panel.id, { x: Math.max(0, origin.current.x + dx), y: Math.max(0, origin.current.y + dy) })
      } else {
        onMove(panel.id, {
          width: Math.max(300, origin.current.width + dx),
          height: Math.max(200, origin.current.height + dy),
        })
      }
    },
    [onMove, panel.id],
  )

  const endGesture = useCallback(
    (e: React.PointerEvent) => {
      gesture.current = null
      if (e.currentTarget.hasPointerCapture(e.pointerId)) {
        e.currentTarget.releasePointerCapture(e.pointerId)
      }
      setDragging(false)
      setResizing(false)
    },
    [],
  )

  const onKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      const step = e.shiftKey ? 40 : 10
      const resize = e.altKey
      const moves: Record<string, Partial<PanelGeometry>> = {
        ArrowLeft: resize ? { width: panel.width - step } : { x: panel.x - step },
        ArrowRight: resize ? { width: panel.width + step } : { x: panel.x + step },
        ArrowUp: resize ? { height: panel.height - step } : { y: panel.y - step },
        ArrowDown: resize ? { height: panel.height + step } : { y: panel.y + step },
      }
      const change = moves[e.key]
      if (!change) return
      e.preventDefault()
      onMove(panel.id, change)
    },
    [panel, onMove],
  )

  useEffect(() => {
    if (!dragging && !resizing) return
    const previous = document.body.style.userSelect
    document.body.style.userSelect = 'none'
    return () => {
      document.body.style.userSelect = previous
    }
  }, [dragging, resizing])

  return (
    <section
      role="group"
      aria-label={name}
      onPointerDown={() => onFocus(panel.id)}
      className={cn(
        'absolute flex flex-col overflow-hidden rounded-sm border bg-card shadow-lg',
        dragging || resizing ? 'border-primary/50' : 'border-border/60',
      )}
      style={{
        left: panel.x,
        top: panel.y,
        width: panel.width,
        height: panel.minimised ? undefined : panel.height,
        zIndex: panel.z,
      }}
    >
      <header
        onPointerDown={beginDrag}
        onPointerMove={onPointerMove}
        onPointerUp={endGesture}
        onPointerCancel={endGesture}
        onKeyDown={onKeyDown}
        tabIndex={0}
        className="flex shrink-0 cursor-move items-center gap-2 border-b border-border/40 bg-muted/30 px-2 py-1.5 focus:outline-none focus-visible:ring-1 focus-visible:ring-primary"
      >
        <span className="min-w-0 flex-1 truncate font-mono text-[10px] uppercase tracking-widest text-muted-foreground">{name}</span>

        <button
          type="button"
          onPointerDown={(e) => e.stopPropagation()}
          onClick={() => onToggleMinimise(panel.id)}
          aria-label={panel.minimised ? t('tiAnalyst.canvas.expand', { title: name }) : t('tiAnalyst.canvas.collapse', { title: name })}
          className="rounded-sm p-0.5 text-muted-foreground transition-colors hover:text-foreground"
        >
          {panel.minimised ? <Square className="h-3 w-3" /> : <Minus className="h-3 w-3" />}
        </button>
        <button
          type="button"
          onPointerDown={(e) => e.stopPropagation()}
          onClick={() => onClose(panel.id)}
          aria-label={t('tiAnalyst.canvas.closePanel', { title: name })}
          className="rounded-sm p-0.5 text-muted-foreground transition-colors hover:text-destructive"
        >
          <X className="h-3 w-3" />
        </button>
      </header>

      {!panel.minimised && (
        <>
          <div className="min-h-0 flex-1 overflow-auto p-2">{children}</div>

          <button
            type="button"
            onPointerDown={beginResize}
            onPointerMove={onPointerMove}
            onPointerUp={endGesture}
            onPointerCancel={endGesture}
            aria-label={t('tiAnalyst.canvas.resizePanel', { title: name })}
            className="absolute bottom-0 right-0 h-4 w-4 cursor-nwse-resize text-muted-foreground/40 hover:text-primary"
          >
            <svg viewBox="0 0 16 16" className="h-full w-full" aria-hidden>
              <path d="M11 15 L15 11 M6 15 L15 6" stroke="currentColor" strokeWidth="1.5" fill="none" />
            </svg>
          </button>
        </>
      )}
    </section>
  )
}
