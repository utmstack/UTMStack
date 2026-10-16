import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ArrowUp,
  ChevronRight,
  Clock,
  ExternalLink,
  Square,
  Sparkles,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { MarkdownMessage } from '@/features/soc-ai/components/MarkdownMessage'
import { useConsoleStore, openPanel } from './store'
import { ASK_EVENT } from './events'
import { focusConversation, useConversationLoop } from './use-conversation-loop'
import { panelForStep } from './step-panel'
import { ToolResult } from './tool-result'
import type { AgentStep, LoopTurn } from './types'

const SUGGESTION_KEYS = [
  'tiAnalyst.agent.suggestionInvestigate',
  'tiAnalyst.agent.suggestionInfrastructure',
  'tiAnalyst.agent.suggestionCounts',
] as const

export function ConversationPanel({
  conversationId,
  model,
  canvasEnabled,
}: {
  conversationId: string
  model: string
  
  canvasEnabled: boolean
}) {
  const { t } = useTranslation()
  const store = useConsoleStore()
  const conversation = store.conversations.find((c) => c.id === conversationId)
  const loop = useConversationLoop(conversationId, conversation)
  const [input, setInput] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)
  const turns = conversation?.turns ?? []

  const ask = useCallback(
    (question: string) => {
      if (!question.trim() || loop.liveState.busy) return
      setInput('')
      void loop.ask(question, model, (panel) =>
        canvasEnabled ? openPanel(conversationId, panel) : undefined,
      )
    },
    [conversationId, loop, model, canvasEnabled],
  )

  const askRef = useRef(ask)
  askRef.current = ask
  useEffect(() => {
    const onAsk = (e: Event) => {
      const q = (e as CustomEvent<string>).detail
      if (q) askRef.current(q)
    }
    window.addEventListener(ASK_EVENT, onAsk)
    return () => window.removeEventListener(ASK_EVENT, onAsk)
  }, [])

  useEffect(() => {
    focusConversation(conversationId)
    return () => focusConversation(null)
  }, [conversationId])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [turns])

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div
        ref={scrollRef}
        className="min-h-0 flex-1 space-y-5 overflow-y-auto px-3 py-4 sm:px-4"
        role="log"
        aria-live="polite"
        aria-label={t('tiAnalyst.conv.transcript')}
      >
        {!turns.length && <Empty onAsk={ask} />}

        {turns.map((turn) => (
          <TurnView key={turn.id} turn={turn} canvasEnabled={canvasEnabled} conversationId={conversationId} />
        ))}

        {loop.liveState.busy && (
          <div className="flex items-center gap-2 rounded-md border border-primary/30 bg-primary/5 px-3 py-2 font-mono text-[11px] text-primary">
            <Clock className="h-3.5 w-3.5 shrink-0" />
            {t('tiAnalyst.agent.thinking')}
          </div>
        )}
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault()
          ask(input)
        }}
        className="shrink-0 border-t border-border bg-card/80 px-3 py-3 sm:px-4"
      >
        <div className="flex items-center gap-2">
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={loop.liveState.busy}
            placeholder={t('tiAnalyst.agent.placeholder')}
            aria-label={t('tiAnalyst.agent.ask')}
            className="h-10 min-w-0 flex-1 rounded-md border border-border bg-muted/30 px-3 font-mono text-sm placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-primary disabled:opacity-50"
          />
          {loop.liveState.busy ? (
            <button
              type="button"
              aria-label={t('tiAnalyst.agent.stop')}
              onClick={() => loop.stop()}
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border border-destructive/40 text-destructive transition-colors hover:bg-destructive/10"
            >
              <Square className="h-3.5 w-3.5" />
            </button>
          ) : (
            <button
              type="submit"
              aria-label={t('tiAnalyst.agent.send')}
              disabled={!input.trim()}
              className={cn(
                'flex h-10 w-10 shrink-0 items-center justify-center rounded-md border transition-colors',
                input.trim()
                  ? 'border-primary/40 text-primary hover:bg-primary/10'
                  : 'border-border text-muted-foreground/40',
              )}
            >
              <ArrowUp className="h-4 w-4" />
            </button>
          )}
        </div>
      </form>
    </div>
  )
}

function TurnView({
  turn,
  canvasEnabled,
  conversationId,
}: {
  turn: LoopTurn
  canvasEnabled: boolean
  conversationId: string
}) {
  const { t } = useTranslation()

  return (
    <div className="space-y-2.5">
      <div className="flex justify-end">
        <div className="max-w-[85%] rounded-md border border-primary/25 bg-primary/10 px-3 py-2 font-mono text-xs text-foreground">
          {turn.question}
        </div>
      </div>

      {canvasEnabled
        ? turn.steps.map((step) => (
            <ToolChip
              key={step.id}
              step={step}
              onOpenPanel={() => {
                const panel = panelForStep(step)
                if (panel) openPanel(conversationId, panel)
              }}
            />
          ))
        : turn.steps.map((step) => <ToolResult key={step.id} step={step} onPivot={() => undefined} />)}

      {turn.text && (
        <div className="rounded-md border border-border bg-card px-3 py-2.5">
          <MarkdownMessage text={turn.text} />
        </div>
      )}

      {turn.status === 'running' && !turn.text && !turn.steps.length && (
        <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
          <Sparkles className="h-3 w-3 animate-pulse" />
          {t('tiAnalyst.agent.thinking')}
        </div>
      )}

      {turn.error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 font-mono text-xs text-destructive">
          {turn.error}
        </div>
      )}
    </div>
  )
}

function ToolChip({
  step,
  onOpenPanel,
}: {
  step: AgentStep
  onOpenPanel(): void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const bodyId = `${step.id}-body`
  const panel = useMemo(() => panelForStep(step), [step])

  return (
    <div className="overflow-hidden rounded-md border border-border bg-card">
      <div className="flex items-center gap-2 px-2 py-1.5">
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          aria-expanded={open}
          aria-controls={bodyId}
          aria-label={open ? t('tiAnalyst.tool.collapse') : t('tiAnalyst.tool.expand')}
          className="flex min-w-0 flex-1 items-center gap-2 text-start text-[11px] uppercase tracking-wider text-muted-foreground transition-colors hover:text-foreground"
        >
          <ChevronRight className={cn('h-3 w-3 shrink-0 transition-transform', open && 'rotate-90')} />
          <span className="truncate">{step.toolName.replace(/_/g, ' ')}</span>
          {step.status === 'running' && (
            <span className="h-1 w-10 shrink-0 animate-pulse rounded bg-primary/60" />
          )}
          {step.status === 'error' && <span className="truncate text-destructive">{step.error}</span>}
        </button>
        {panel && (
          <button
            type="button"
            onClick={onOpenPanel}
            aria-label={t('tiAnalyst.canvas.pinOut', { title: panel.title })}
            className="shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:text-primary"
          >
            <ExternalLink className="h-3 w-3" />
          </button>
        )}
      </div>
      {open && (
        <div id={bodyId} className="border-t border-border p-2">
          <ToolResult step={step} onPivot={() => undefined} />
        </div>
      )}
    </div>
  )
}

function Empty({ onAsk }: { onAsk(question: string): void }) {
  const { t } = useTranslation()
  return (
    <div className="mx-auto max-w-lg space-y-4 pt-8 text-center">
      <Sparkles className="mx-auto h-6 w-6 text-primary/60" />
      <p className="font-mono text-sm text-muted-foreground">{t('tiAnalyst.agent.intro')}</p>
      <div className="space-y-1.5">
        {SUGGESTION_KEYS.map((key) => (
          <button
            key={key}
            type="button"
            onClick={() => onAsk(t(key))}
            className="block w-full rounded-md border border-border px-3 py-2 text-start font-mono text-xs text-muted-foreground transition-colors hover:border-primary/40 hover:text-foreground"
          >
            {t(key)}
          </button>
        ))}
      </div>
    </div>
  )
}
