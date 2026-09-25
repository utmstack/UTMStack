import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type Dispatch, type ReactNode, type SetStateAction } from 'react'
import { useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'
import { DASHBOARDS_QUERY_KEYS } from '@/features/dashboard/hooks/useDashboards'
import { VISUALIZATIONS_QUERY_KEYS } from '@/features/dashboard/hooks/useVisualizations'
import { extractNavigation, streamChat, type ChatHistoryTurn, type NavAction } from './lib/chat-stream'

// How many prior text turns to replay to the backend as chat memory. Server-side
// compaction will still trim if this exceeds the model context window.
const HISTORY_LIMIT = 10

export interface CurrentStep {
  tool: string
  status: 'running' | 'done' | 'error'
}

export interface SocAiMessage {
  id: number
  role: 'user' | 'ai'
  text: string
  pending?: boolean
  error?: boolean
  currentStep?: CurrentStep | null
  actions?: NavAction[]
}

// 'panel', 'dashboard-create' and 'dashboard-edit' all render in the floating
// SocAiPanel (see activeScope) — separate threads, same UI. 'home' has its
// own inline transcript (HomeChatTranscript) and never shows in the panel.
export type SocAiScope = 'panel' | 'home' | 'dashboard-create' | 'dashboard-edit' | 'soar-edit'

export interface SoarEditTarget {
  relPath: string
  name: string
}

/** Which existing dashboard the 'dashboard-edit' thread is currently scoped to. */
export interface DashboardEditTarget {
  id: string
  name: string
}

/**
 * The item the person has open beside the chat (an alert or incident drawer).
 * `context` is what the agent is told about it; `label` is what the chip in the
 * composer shows, so it is visible what is being shared.
 */
export interface SocAiFocus {
  kind: 'alert' | 'incident'
  id: string
  label: string
  context: string
}

/** The page the agent is told the person is on, plus the item they have open. */
export function composePage(page: string, focus: SocAiFocus | null): string {
  return focus ? `${page}\n\n${focus.context}` : page
}

interface SocAiContextValue {
  open: boolean
  expanded: boolean
  // Which scope's thread the floating panel is currently showing.
  activeScope: SocAiScope
  messages: SocAiMessage[]
  homeMessages: SocAiMessage[]
  dashboardCreateMessages: SocAiMessage[]
  dashboardEditMessages: SocAiMessage[]
  soarEditMessages: SocAiMessage[]
  dashboardEditTarget: DashboardEditTarget | null
  // Called right before opening the panel with scope 'dashboard-edit' so every
  // message sent in that thread carries which dashboard is being worked on.
  setDashboardEditTarget: (target: DashboardEditTarget | null) => void
  soarEditTarget: SoarEditTarget | null
  setSoarEditTarget: (target: SoarEditTarget | null) => void
  soarEditVersion: number
  // The open item being shared with the agent, or null when nothing is open or
  // the person removed it from the conversation.
  focus: SocAiFocus | null
  setFocus: (focus: SocAiFocus | null) => void
  detachFocus: () => void
  // A scope argument switches the panel to that thread before opening it.
  openPanel: (scope?: SocAiScope) => void
  closePanel: () => void
  togglePanel: () => void
  toggleExpand: () => void
  submit: (text: string, opts?: { openPanel?: boolean; scope?: SocAiScope }) => void
  // scope is required rather than defaulted: an optional parameter makes this
  // assignable to `() => void`, so passing it straight to an onClick type-checks
  // and then silently receives the click event as its scope.
  clear: (scope: SocAiScope) => void
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
  '/dashboards/new': 'New Dashboard',
}

function pageContext(pathname: string): string {
  const label = PAGE_LABELS[pathname]
  return label ? `${label} page (route ${pathname})` : `route ${pathname}`
}

export function SocAiProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false)
  const [expanded, setExpanded] = useState(false)
  const [activeScope, setActiveScope] = useState<SocAiScope>('panel')
  const [messages, setMessages] = useState<SocAiMessage[]>([])
  const [homeMessages, setHomeMessages] = useState<SocAiMessage[]>([])
  const [dashboardCreateMessages, setDashboardCreateMessages] = useState<SocAiMessage[]>([])
  const [dashboardEditMessages, setDashboardEditMessages] = useState<SocAiMessage[]>([])
  const [soarEditMessages, setSoarEditMessages] = useState<SocAiMessage[]>([])
  const [dashboardEditTarget, setDashboardEditTarget] = useState<DashboardEditTarget | null>(null)
  const [soarEditTarget, setSoarEditTarget] = useState<SoarEditTarget | null>(null)
  const [soarEditVersion, setSoarEditVersion] = useState(0)
  const [openItem, setOpenItem] = useState<SocAiFocus | null>(null)
  // The item the person chose to stop sharing. Reset when nothing is open, so
  // the next time that item opens it is shared again.
  const [detachedKey, setDetachedKey] = useState<string | null>(null)
  const idRef = useRef(0)
  const nextId = () => ++idRef.current
  const abortRef = useRef<AbortController | null>(null)
  const location = useLocation()
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()

  const setters: Record<SocAiScope, Dispatch<SetStateAction<SocAiMessage[]>>> = {
    panel: setMessages,
    home: setHomeMessages,
    'dashboard-create': setDashboardCreateMessages,
    'dashboard-edit': setDashboardEditMessages,
    'soar-edit': setSoarEditMessages,
  }
  const messagesByScope: Record<SocAiScope, SocAiMessage[]> = {
    panel: messages,
    home: homeMessages,
    'dashboard-create': dashboardCreateMessages,
    'dashboard-edit': dashboardEditMessages,
    'soar-edit': soarEditMessages,
  }

  const openKey = openItem ? `${openItem.kind}:${openItem.id}` : null
  const focus = openItem && openKey !== detachedKey ? openItem : null
  const focusRef = useRef(focus)
  focusRef.current = focus
  const setFocus = useCallback((next: SocAiFocus | null) => {
    setOpenItem(next)
    if (!next) setDetachedKey(null)
  }, [])
  const detachFocus = useCallback(() => setDetachedKey(openKey), [openKey])

  const openPanel = useCallback((scope?: SocAiScope) => {
    if (scope) setActiveScope(scope)
    setOpen(true)
  }, [])
  const closePanel = useCallback(() => setOpen(false), [])
  const togglePanel = useCallback(() => setOpen((v) => !v), [])
  const toggleExpand = useCallback(() => setExpanded((v) => !v), [])
  const clear = useCallback((scope: SocAiScope) => {
    abortRef.current?.abort()
    setters[scope]([])
    // A cleared dashboard thread has nothing left to be scoped by — drop the edit
    // target and go back to the general panel so a leftover "Editing: <dashboard>"
    // title can't outlive its history. Only when NOT on a dashboard page — there
    // the scope is still live.
    if ((scope === 'dashboard-edit' || scope === 'dashboard-create') && !location.pathname.startsWith('/dashboards')) {
      setDashboardEditTarget(null)
      setActiveScope('panel')
    }
    if (scope === 'soar-edit' && !location.pathname.startsWith('/soar')) {
      setSoarEditTarget(null)
      setActiveScope('panel')
    }
  }, [location.pathname])

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
      const current = messagesByScope[scope]
      const last = current[current.length - 1]
      if (last?.role === 'ai' && last.pending) return

      if (opts?.openPanel !== false) {
        setActiveScope(scope)
        setOpen(true)
      }
      abortRef.current?.abort()
      const ac = new AbortController()
      abortRef.current = ac

      const aiId = nextId()
      setters[scope]((m) => [
        ...m,
        { id: nextId(), role: 'user', text },
        { id: aiId, role: 'ai', text: '', pending: true, currentStep: null },
      ])

      // 'dashboard-edit' has no fixed route of its own (it's opened from
      // inside whichever dashboard the user is previewing), so its page
      // context comes from the target set by the "Edit with AI" button
      // instead of the current path.
      const page =
        scope === 'dashboard-edit' && dashboardEditTarget
          ? `Dashboard editor — the user is editing dashboard "${dashboardEditTarget.name}" (dashboard id: ${dashboardEditTarget.id}). Use the dashboards/visualizations tools with this id to add, update, or remove its widgets; check what's already there first (dashboards.get / visualizations.list) before changing it.`
          : scope === 'soar-edit' && soarEditTarget
            ? `SOAR flow editor — user is editing flow "${soarEditTarget.name}" at ${soarEditTarget.relPath}. Call soar.rule.get first, then soar.rule.update with the FULL rule JSON (Conditions + Nodes map); preserve all unrelated nodes.`
            : scope === 'panel'
            ? composePage(pageContext(location.pathname), focusRef.current)
            : pageContext(location.pathname)
      const lang = (i18n.language || 'en').split('-')[0]

      const history: ChatHistoryTurn[] = current
        .filter((m) => m.text && !m.error && !m.pending)
        .slice(-HISTORY_LIMIT)
        .map((m) => ({ role: m.role === 'user' ? 'user' : 'assistant', content: m.text }))

      streamChat(
        { task: text, page, lang, history },
        (ev) => {
          patchMsg(scope, aiId, (msg) => {
            switch (ev.kind) {
              case 'tool_call':
                return { ...msg, currentStep: { tool: ev.tool ?? 'tool', status: 'running' } }
              case 'tool_result':
                return { ...msg, currentStep: { tool: ev.tool ?? 'tool', status: ev.isError ? 'error' : 'done' } }
              case 'compaction':
                return { ...msg, currentStep: { tool: t('socAi.chat.compacting'), status: 'running' } }
              case 'final': {
                const { text: clean, actions } = extractNavigation(ev.text ?? '')
                return { ...msg, text: clean, actions, pending: false, currentStep: null }
              }
              case 'error':
                return { ...msg, text: ev.text || t('socAi.chat.errorGeneric'), error: true, pending: false, currentStep: null }
              default:
                return msg
            }
          })
          // The dashboard-edit thread writes straight to the backend via MCP
          // tools — the dashboard grid has no other way to learn its widgets
          // changed, so refetch once the run settles (successfully or not,
          // since a failed run may still have created a few widgets first).
          if (scope === 'dashboard-edit' && (ev.kind === 'final' || ev.kind === 'error')) {
            void queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
            void queryClient.invalidateQueries({ queryKey: VISUALIZATIONS_QUERY_KEYS.all })
          }
          if (scope === 'soar-edit' && (ev.kind === 'final' || ev.kind === 'error')) {
            setSoarEditVersion((v) => v + 1)
          }
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
    [
      location.pathname,
      i18n.language,
      patchMsg,
      t,
      messages,
      homeMessages,
      dashboardCreateMessages,
      dashboardEditMessages,
      soarEditMessages,
      dashboardEditTarget,
      soarEditTarget,
      queryClient,
    ],
  )

  const value = useMemo(
    () => ({
      open,
      expanded,
      activeScope,
      messages,
      homeMessages,
      dashboardCreateMessages,
      dashboardEditMessages,
      soarEditMessages,
      dashboardEditTarget,
      setDashboardEditTarget,
      soarEditTarget,
      setSoarEditTarget,
      soarEditVersion,
      focus,
      setFocus,
      detachFocus,
      openPanel,
      closePanel,
      togglePanel,
      toggleExpand,
      submit,
      clear,
    }),
    [
      open,
      expanded,
      activeScope,
      messages,
      homeMessages,
      dashboardCreateMessages,
      dashboardEditMessages,
      soarEditMessages,
      dashboardEditTarget,
      soarEditTarget,
      soarEditVersion,
      focus,
      setFocus,
      detachFocus,
      openPanel,
      closePanel,
      togglePanel,
      toggleExpand,
      submit,
      clear,
    ],
  )

  return <SocAiContext.Provider value={value}>{children}</SocAiContext.Provider>
}

/**
 * The assistant when it is available, null when it is not. The Topbar lives
 * both inside the dashboard (where the provider exists) and in the federation
 * shell (where it does not), so its entry point has to be able to ask without
 * bringing the page down.
 */
export function useSocAiOptional(): SocAiContextValue | null {
  return useContext(SocAiContext)
}

export function useSocAi(): SocAiContextValue {
  const ctx = useContext(SocAiContext)
  if (!ctx) throw new Error('useSocAi must be used within SocAiProvider')
  return ctx
}

/**
 * Shares the item a drawer shows with the assistant while the drawer is open,
 * so "is this a false positive?" needs no id. A no-op outside the dashboard
 * shell, where there is no assistant.
 */
export function useSocAiFocus(focus: SocAiFocus | null) {
  const setFocus = useContext(SocAiContext)?.setFocus
  const key = focus ? `${focus.kind}|${focus.id}|${focus.label}|${focus.context}` : ''
  useEffect(() => {
    if (!setFocus || !focus) return
    setFocus(focus)
    return () => setFocus(null)
    // `key` covers every field of `focus`; the object itself is rebuilt each render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [setFocus, key])
}
