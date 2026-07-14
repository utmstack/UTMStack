import { createContext, useCallback, useContext, useMemo, useRef, useState, type Dispatch, type ReactNode, type SetStateAction } from 'react'
import { useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { extractNavigation, streamChat, type NavAction } from './lib/chat-stream'

export interface ToolStep {
  tool: string
  status: 'running' | 'done' | 'error'
}

export interface SocAiMessage {
  id: number
  role: 'user' | 'ai'
  text: string
  pending?: boolean
  error?: boolean
  steps?: ToolStep[]
  actions?: NavAction[]
}

export type SocAiScope = 'panel' | 'home'

interface SocAiContextValue {
  open: boolean
  expanded: boolean
  messages: SocAiMessage[]
  homeMessages: SocAiMessage[]
  openPanel: () => void
  closePanel: () => void
  togglePanel: () => void
  toggleExpand: () => void
  submit: (text: string, opts?: { openPanel?: boolean; scope?: SocAiScope }) => void
  clear: (scope?: SocAiScope) => void
}

const SocAiContext = createContext<SocAiContextValue | null>(null)

// Friendly labels for the page context sent to the agent so it can pick relevant
// tools and craft navigation.
const PAGE_LABELS: Record<string, string> = {
  '/home': 'Home',
  '/threat-management/alerts': 'Alerts',
  '/threat-management/incidents': 'Incidents',
  '/threat-management/adversaries': 'Adversaries',
  '/log-explorer': 'Log Explorer',
  '/user-auditor': 'User Auditor',
  '/threat-intelligence': 'Threat Intelligence',
  '/compliance': 'Compliance',
  '/datasources': 'Data Sources',
  '/integrations': 'Integrations',
}

function pageContext(pathname: string): string {
  const label = PAGE_LABELS[pathname]
  return label ? `${label} page (route ${pathname})` : `route ${pathname}`
}

export function SocAiProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false)
  const [expanded, setExpanded] = useState(false)
  const [messages, setMessages] = useState<SocAiMessage[]>([])
  const [homeMessages, setHomeMessages] = useState<SocAiMessage[]>([])
  const idRef = useRef(0)
  const nextId = () => ++idRef.current
  const abortRef = useRef<AbortController | null>(null)
  const location = useLocation()
  const { t, i18n } = useTranslation()

  const setters: Record<SocAiScope, Dispatch<SetStateAction<SocAiMessage[]>>> = {
    panel: setMessages,
    home: setHomeMessages,
  }

  const openPanel = useCallback(() => setOpen(true), [])
  const closePanel = useCallback(() => setOpen(false), [])
  const togglePanel = useCallback(() => setOpen((v) => !v), [])
  const toggleExpand = useCallback(() => setExpanded((v) => !v), [])
  const clear = useCallback((scope: SocAiScope = 'panel') => {
    abortRef.current?.abort()
    setters[scope]([])
  }, [])

  const patchMsg = useCallback((scope: SocAiScope, id: number, fn: (m: SocAiMessage) => SocAiMessage) => {
    setters[scope]((list) => list.map((m) => (m.id === id ? fn(m) : m)))
  }, [])

  const submit = useCallback(
    (raw: string, opts?: { openPanel?: boolean; scope?: SocAiScope }) => {
      const text = raw.trim()
      if (!text) return
      const scope: SocAiScope = opts?.scope ?? 'panel'

      // No queueing — ignore a new submit while this scope's last message is
      // still being answered, instead of aborting it out from under itself.
      const current = scope === 'panel' ? messages : homeMessages
      const last = current[current.length - 1]
      if (last?.role === 'ai' && last.pending) return

      if (opts?.openPanel !== false) setOpen(true)
      abortRef.current?.abort()
      const ac = new AbortController()
      abortRef.current = ac

      const aiId = nextId()
      setters[scope]((m) => [
        ...m,
        { id: nextId(), role: 'user', text },
        { id: aiId, role: 'ai', text: '', pending: true, steps: [] },
      ])

      const page = pageContext(location.pathname)
      const lang = (i18n.language || 'en').split('-')[0]

      streamChat(
        { task: text, page, lang },
        (ev) => {
          patchMsg(scope, aiId, (msg) => {
            switch (ev.kind) {
              case 'tool_call':
                return { ...msg, steps: [...(msg.steps ?? []), { tool: ev.tool ?? 'tool', status: 'running' }] }
              case 'tool_result': {
                const steps = (msg.steps ?? []).slice()
                for (let i = steps.length - 1; i >= 0; i--) {
                  if (steps[i].tool === ev.tool && steps[i].status === 'running') {
                    steps[i] = { ...steps[i], status: ev.isError ? 'error' : 'done' }
                    break
                  }
                }
                return { ...msg, steps }
              }
              case 'final': {
                const { text: clean, actions } = extractNavigation(ev.text ?? '')
                return { ...msg, text: clean, actions, pending: false }
              }
              case 'error':
                return { ...msg, text: ev.text || t('socAi.chat.errorGeneric'), error: true, pending: false }
              default:
                return msg
            }
          })
        },
        ac.signal,
      ).catch((err) => {
        if (ac.signal.aborted) return
        patchMsg(scope, aiId, (msg) => ({
          ...msg,
          text: err instanceof Error ? err.message : t('socAi.chat.errorUnreachable'),
          error: true,
          pending: false,
        }))
      })
    },
    [location.pathname, i18n.language, patchMsg, t, messages, homeMessages],
  )

  const value = useMemo(
    () => ({ open, expanded, messages, homeMessages, openPanel, closePanel, togglePanel, toggleExpand, submit, clear }),
    [open, expanded, messages, homeMessages, openPanel, closePanel, togglePanel, toggleExpand, submit, clear],
  )

  return <SocAiContext.Provider value={value}>{children}</SocAiContext.Provider>
}

export function useSocAi(): SocAiContextValue {
  const ctx = useContext(SocAiContext)
  if (!ctx) throw new Error('useSocAi must be used within SocAiProvider')
  return ctx
}
