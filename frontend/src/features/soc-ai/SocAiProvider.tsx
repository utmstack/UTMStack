import { createContext, useCallback, useContext, useMemo, useRef, useState, type ReactNode } from 'react'
import { mockReply } from './lib/mock'

export interface SocAiMessage {
  id: number
  role: 'user' | 'ai'
  text: string
  pending?: boolean
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

export function SocAiProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false)
  const [expanded, setExpanded] = useState(false)
  const [messages, setMessages] = useState<SocAiMessage[]>([])
  const idRef = useRef(0)
  const nextId = () => ++idRef.current

  const openPanel = useCallback(() => setOpen(true), [])
  const closePanel = useCallback(() => setOpen(false), [])
  const togglePanel = useCallback(() => setOpen((v) => !v), [])
  const toggleExpand = useCallback(() => setExpanded((v) => !v), [])
  const clear = useCallback(() => setMessages([]), [])

  const submit = useCallback((raw: string) => {
    const text = raw.trim()
    if (!text) return
    setOpen(true)
    const aiId = nextId()
    setMessages((m) => [
      ...m,
      { id: nextId(), role: 'user', text },
      { id: aiId, role: 'ai', text: '', pending: true },
    ])
    // Mock latency, then swap the pending bubble for a canned rich answer.
    window.setTimeout(() => {
      setMessages((m) => m.map((msg) => (msg.id === aiId ? { ...msg, text: mockReply(text), pending: false } : msg)))
    }, 650)
  }, [])

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
