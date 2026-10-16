import { useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useTranslation } from 'react-i18next'
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { byRecency, createConversation, useConsoleStore } from './store'
import { ASK_EVENT } from './events'
import { ConversationRail } from './conversation-rail'
import { ConversationPanel } from './conversation-panel'
import { Canvas } from './canvas'
import {
  SplitHandle,
  useSplit,
  clampCanvasFraction,
} from './split-handle'
import { CANVAS_BREAKPOINT, AVAILABLE_MODELS, DEFAULT_MODEL } from './types'

export function ThreatIntelConsole() {
  const { t } = useTranslation()
  const store = useConsoleStore()

  const [conversationId, setConversationId] = useState<string | null>(null)
  const [model, setModel] = useState(DEFAULT_MODEL)
  const [mobileRailOpen, setMobileRailOpen] = useState(false)

  const canvasWide = useCanvasBreakpoint()
  const split = useSplit()
  const [workspaceRef, workspaceWidth] = useMeasuredWidth()

  useEffect(() => {
    if (conversationId || !store.conversations.length) return
    setConversationId([...store.conversations].sort(byRecency)[0].id)
  }, [conversationId, store.conversations])

  useEffect(() => {
    if (!conversationId) return
    if (!store.conversations.some((c) => c.id === conversationId)) setConversationId(null)
  }, [conversationId, store.conversations])

  const askInCurrentConversation = useCallback(
    (question: string) => {
      if (!conversationId) {
        const title = question.length > 40 ? `${question.slice(0, 38)}…` : question
        const conversation = createConversation(title)
        setConversationId(conversation.id)
        return
      }
      window.dispatchEvent(new CustomEvent(ASK_EVENT, { detail: question }))
    },
    [conversationId],
  )

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-2 border-b border-border/40 bg-muted/10 px-3 py-1.5 sm:px-4">
        <button
          type="button"
          aria-label={mobileRailOpen ? t('nav.closeMenu') : t('nav.openMenu')}
          aria-expanded={mobileRailOpen}
          onClick={() => setMobileRailOpen((prev) => !prev)}
          className="flex h-7 w-7 shrink-0 items-center justify-center rounded-sm border border-border/40 bg-muted/20 text-muted-foreground hover:text-foreground lg:hidden"
        >
          {mobileRailOpen ? <PanelLeftClose className="h-4 w-4" /> : <PanelLeftOpen className="h-4 w-4" />}
        </button>
        <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          {t('tiAnalyst.console.subtitle')}
        </span>
        <div className="flex-1" />
        <label className="flex items-center gap-1.5">
          <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
            {t('tiAnalyst.console.model')}
          </span>
          <select
            value={model}
            onChange={(e) => setModel(e.target.value)}
            className="h-6 rounded-sm border border-border/50 bg-background px-1.5 font-mono text-[10px]"
          >
            {AVAILABLE_MODELS.map((m) => (
              <option key={m.id} value={m.id}>
                {m.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      <div className="flex min-h-0 flex-1">
        {mobileRailOpen &&
          createPortal(
            <>
              <div
                className="fixed inset-0 z-50 bg-black/40 lg:hidden"
                onClick={() => setMobileRailOpen(false)}
              />
              <ConversationRail
                conversationId={conversationId}
                onSelectConversation={(id) => {
                  setConversationId(id)
                  setMobileRailOpen(false)
                }}
                className="fixed inset-y-0 left-0 z-50 bg-background"
              />
            </>,
            document.body,
          )}

        <ConversationRail
          conversationId={conversationId}
          onSelectConversation={setConversationId}
          className="hidden lg:flex"
        />

        <div ref={workspaceRef} className="flex min-w-0 flex-1">
          {conversationId ? (
            <ConversationPanel
              key={conversationId}
              conversationId={conversationId}
              model={model}
              canvasEnabled={canvasWide}
            />
          ) : (
            <div className="flex flex-1 items-center justify-center px-6 text-center">
              <p className="max-w-sm font-mono text-xs text-muted-foreground">
                {t('tiAnalyst.conv.selectPrompt')}
              </p>
            </div>
          )}

          {canvasWide && conversationId && (
            <>
              <SplitHandle
                fraction={split.fraction}
                containerWidth={workspaceWidth}
                onChange={(next) => split.setFraction(next, workspaceWidth)}
                onReset={split.reset}
              />
              <div
                className="shrink-0"
                style={{ width: `${clampCanvasFraction(split.fraction, workspaceWidth) * 100}%` }}
              >
                <Canvas
                  conversationId={conversationId}
                  onPivot={(target) => askInCurrentConversation(target.value || target.id)}
                  onAskAgent={askInCurrentConversation}
                />
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function useMeasuredWidth(): [(node: HTMLDivElement | null) => void, number] {
  const [width, setWidth] = useState(0)
  const observer = useRef<ResizeObserver | null>(null)

  const ref = useCallback((node: HTMLDivElement | null) => {
    observer.current?.disconnect()
    observer.current = null
    if (!node) return
    setWidth(node.getBoundingClientRect().width)
    if (typeof ResizeObserver === 'undefined') return
    observer.current = new ResizeObserver((entries) => setWidth(entries[0].contentRect.width))
    observer.current.observe(node)
  }, [])

  useEffect(() => () => observer.current?.disconnect(), [])

  return [ref, width]
}

function useCanvasBreakpoint(): boolean {
  const [wide, setWide] = useState(
    () => typeof window !== 'undefined' && window.innerWidth >= CANVAS_BREAKPOINT,
  )
  useEffect(() => {
    const query = window.matchMedia(`(min-width: ${CANVAS_BREAKPOINT}px)`)
    const update = () => setWide(query.matches)
    update()
    query.addEventListener('change', update)
    return () => query.removeEventListener('change', update)
  }, [])
  return wide
}
