import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { LayoutGrid, X } from 'lucide-react'
import {
  closeAllPanels,
  closePanel,
  confinePanels,
  focusPanel,
  movePanel,
  tilePanels,
  toggleMinimised,
  useConsoleStore,
} from './store'
import { CanvasPanelFrame } from './canvas-panel'
import { CanvasContent } from './canvas-content'

export function Canvas({
  conversationId,
  onPivot,
  onAskAgent,
}: {
  conversationId: string
  onPivot(target: { id: string; type: string; value: string }): void
  onAskAgent(question: string): void
}) {
  const { t } = useTranslation()
  const surfaceRef = useRef<HTMLDivElement>(null)
  const store = useConsoleStore()
  const panels = store.canvases[conversationId] ?? []

  const tileToFit = () => {
    const rect = surfaceRef.current?.getBoundingClientRect()
    tilePanels(conversationId, { width: rect?.width ?? 1200, height: rect?.height ?? 800 })
  }

  useEffect(() => {
    const surface = surfaceRef.current
    if (!surface || typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver((entries) => {
      const { width, height } = entries[0].contentRect
      if (!width || !height) return
      confinePanels(conversationId, { width, height })
    })
    observer.observe(surface)
    return () => observer.disconnect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [conversationId])

  useEffect(() => {
    const surface = surfaceRef.current
    if (!surface) return
    const { width, height } = surface.getBoundingClientRect()
    if (!width || !height) return
    confinePanels(conversationId, { width, height })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [panels.length, conversationId])

  return (
    <section
      aria-label={t('tiAnalyst.canvas.title')}
      className="flex h-full min-h-0 flex-col border-l border-border"
    >
      <div className="flex shrink-0 items-center gap-2 border-b border-border bg-card px-2 py-1.5">
        <h2 className="text-[11px] uppercase tracking-wider text-muted-foreground">
          {t('tiAnalyst.canvas.title')}
        </h2>
        <div className="flex-1" />
        <button
          type="button"
          onClick={tileToFit}
          disabled={!panels.length}
          aria-label={t('tiAnalyst.canvas.tile')}
          className="flex items-center gap-1 rounded-md border border-border px-1.5 py-0.5 text-[11px] uppercase tracking-wider text-muted-foreground transition-colors hover:text-foreground disabled:opacity-40"
        >
          <LayoutGrid className="h-3 w-3" />
          {t('tiAnalyst.canvas.tile')}
        </button>
        <button
          type="button"
          onClick={() => closeAllPanels(conversationId)}
          disabled={!panels.length}
          aria-label={t('tiAnalyst.canvas.closeAll')}
          className="flex items-center gap-1 rounded-md border border-border px-1.5 py-0.5 text-[11px] uppercase tracking-wider text-muted-foreground transition-colors hover:text-destructive disabled:opacity-40"
        >
          <X className="h-3 w-3" />
          {t('tiAnalyst.canvas.closeAll')}
        </button>
      </div>

      <div ref={surfaceRef} className="relative min-h-0 flex-1 overflow-hidden bg-background/40">
        {!panels.length && (
          <p className="absolute inset-0 flex items-center justify-center px-6 text-center font-mono text-[11px] text-muted-foreground/70">
            {t('tiAnalyst.canvas.empty')}
          </p>
        )}

        {panels.map((panel) => (
          <CanvasPanelFrame
            key={panel.id}
            panel={panel}
            onMove={(id, geo) => movePanel(conversationId, id, geo)}
            onFocus={(id) => focusPanel(conversationId, id)}
            onClose={(id) => closePanel(conversationId, id)}
            onToggleMinimise={(id) => toggleMinimised(conversationId, id)}
          >
            <CanvasContent panel={panel} onPivot={onPivot} onAskAgent={onAskAgent} />
          </CanvasPanelFrame>
        ))}
      </div>
    </section>
  )
}
