import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { AlertCircle, ArrowUpRight, Check, Copy, Loader2, Wrench, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useSocAi, type SocAiMessage, type ToolStep } from '../SocAiProvider'
import type { NavAction } from '../lib/chat-stream'
import { MarkdownMessage } from './MarkdownMessage'

// Logical navigation destinations the agent may emit → app routes. Filters/time
// travel in router state for the target page to apply (where supported).
const DEST_ROUTE: Record<string, string> = {
  'log-explorer': '/log-explorer',
  alerts: '/threat-management/alerts',
  incidents: '/threat-management/incidents',
  adversaries: '/threat-management/adversaries',
  compliance: '/compliance',
  datasources: '/datasources',
  dashboards: '/home',
}

export function MessageRow({ message }: { message: SocAiMessage }) {
  if (message.role === 'user') {
    return (
      <div className="max-w-[80%] self-end rounded-2xl rounded-br-md bg-foreground/[0.07] px-3.5 py-1.5 text-foreground [overflow-wrap:anywhere]">
        {message.text}
      </div>
    )
  }

  const hasSteps = (message.steps?.length ?? 0) > 0
  return (
    <div className="self-stretch">
      {hasSteps && <Steps steps={message.steps!} />}
      {message.pending && !message.text ? (
        <TypingDots />
      ) : message.error ? (
        <div className="flex items-start gap-2 rounded-md border border-red-500/30 bg-red-500/5 px-3 py-2 text-xs text-red-600 dark:text-red-300">
          <AlertCircle size={14} className="mt-0.5 shrink-0" />
          <span>{message.text}</span>
        </div>
      ) : (
        <>
          <MarkdownMessage text={message.text} />
          {message.actions && message.actions.length > 0 && <Actions actions={message.actions} />}
          <CopyButton text={message.text} />
        </>
      )}
    </div>
  )
}

function Steps({ steps }: { steps: ToolStep[] }) {
  return (
    <div className="mb-2 space-y-1 rounded-md border border-border bg-muted/30 px-2.5 py-2">
      {steps.map((s, i) => (
        <div key={i} className="flex items-center gap-2 text-[11px] text-muted-foreground">
          {s.status === 'running' ? (
            <Loader2 size={12} className="shrink-0 animate-spin" />
          ) : s.status === 'error' ? (
            <X size={12} className="shrink-0 text-red-500" />
          ) : (
            <Check size={12} className="shrink-0 text-emerald-500" />
          )}
          <Wrench size={11} className="shrink-0 opacity-60" />
          <span className="truncate font-mono">{s.tool}</span>
        </div>
      ))}
    </div>
  )
}

function Actions({ actions }: { actions: NavAction[] }) {
  const navigate = useNavigate()
  const { closePanel } = useSocAi()

  const go = (a: NavAction) => {
    const route = DEST_ROUTE[a.destination]
    if (!route) return
    closePanel()
    navigate(route, { state: { socaiFilters: a.filters ?? [], socaiTime: a.time } })
  }

  return (
    <div className="mt-3 flex flex-wrap gap-2">
      {actions
        .filter((a) => DEST_ROUTE[a.destination])
        .map((a, i) => (
          <button
            key={i}
            type="button"
            onClick={() => go(a)}
            className="inline-flex items-center gap-1.5 rounded-md border border-primary/40 bg-primary/5 px-2.5 py-1.5 text-xs font-medium text-primary transition-colors hover:bg-primary/10"
          >
            {a.label || 'Open'}
            <ArrowUpRight size={13} />
          </button>
        ))}
    </div>
  )
}

function TypingDots() {
  return (
    <div className="flex items-center gap-1 py-1 text-muted-foreground">
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          className="h-1.5 w-1.5 animate-bounce rounded-full bg-current"
          style={{ animationDelay: `${i * 120}ms` }}
        />
      ))}
    </div>
  )
}

function CopyButton({ text }: { text: string }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const copy = () => {
    void navigator.clipboard?.writeText(text)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1500)
  }
  return (
    <button
      type="button"
      onClick={copy}
      className={cn(
        'mt-2.5 inline-flex h-6 items-center gap-1.5 rounded-md border border-transparent px-2 text-xs font-medium text-muted-foreground transition-colors hover:border-border hover:bg-muted hover:text-foreground',
        copied && 'border-primary/35 bg-primary/10 text-primary',
      )}
    >
      {copied ? <Check size={13} /> : <Copy size={13} />}
      {copied ? t('socAi.chat.copied') : t('socAi.chat.copy')}
    </button>
  )
}
