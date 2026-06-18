import { createContext, useCallback, useContext, useMemo, useRef, useState, type ReactNode } from 'react'
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

interface SocAiContextValue {
  open: boolean
  expanded: boolean
  messages: SocAiMessage[]
  openPanel: () => void
  closePanel: () => void
  togglePanel: () => void
  toggleExpand: () => void
  submit: (text: string) => void
  clear: () => void
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
  const idRef = useRef(0)
  const nextId = () => ++idRef.current
  const abortRef = useRef<AbortController | null>(null)
  const location = useLocation()
  const { t, i18n } = useTranslation()

  const openPanel = useCallback(() => setOpen(true), [])
  const closePanel = useCallback(() => setOpen(false), [])
  const togglePanel = useCallback(() => setOpen((v) => !v), [])
  const toggleExpand = useCallback(() => setExpanded((v) => !v), [])
  const clear = useCallback(() => {
    abortRef.current?.abort()
    setMessages([])
  }, [])

  const patchMsg = useCallback((id: number, fn: (m: SocAiMessage) => SocAiMessage) => {
    setMessages((list) => list.map((m) => (m.id === id ? fn(m) : m)))
  }, [])

  const submit = useCallback(
    (raw: string) => {
      const text = raw.trim()
      if (!text) return
      setOpen(true)
      abortRef.current?.abort()
      const ac = new AbortController()
      abortRef.current = ac

      const aiId = nextId()
      setMessages((m) => [
        ...m,
        { id: nextId(), role: 'user', text },
        { id: aiId, role: 'ai', text: '', pending: true, steps: [] },
      ])

      const page = pageContext(location.pathname)
      const lang = (i18n.language || 'en').split('-')[0]

      streamChat(
        { task: text, page, lang },
        (ev) => {
          patchMsg(aiId, (msg) => {
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
        patchMsg(aiId, (msg) => ({
          ...msg,
          text: err instanceof Error ? err.message : t('socAi.chat.errorUnreachable'),
          error: true,
          pending: false,
        }))
      })
    },
    [location.pathname, i18n.language, patchMsg, t],
  )

  const value = useMemo(
    () => ({ open, expanded, messages, openPanel, closePanel, togglePanel, toggleExpand, submit, clear }),
    [open, expanded, messages, openPanel, closePanel, togglePanel, toggleExpand, submit, clear],
  )

  return <SocAiContext.Provider value={value}>{children}</SocAiContext.Provider>
}

export function useSocAi(): SocAiContextValue {
  const ctx = useContext(SocAiContext)
  if (!ctx) throw new Error('useSocAi must be used within SocAiProvider')
  return ctx
}
