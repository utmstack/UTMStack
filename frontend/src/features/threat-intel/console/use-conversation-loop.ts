import { useCallback, useEffect, useRef, useSyncExternalStore } from 'react'
import { runAgent, explainAgentError } from './agent'
import {
  appendTurn,
  patchTurn,
  replaceHistory,
} from './store'
import { panelForStep } from './step-panel'
import type { AgentStep, Conversation, LoopTurn, PanelRequest } from './types'

interface LiveState {
  busy: boolean
  unread: number
  abort: AbortController | null
}

const live = new Map<string, LiveState>()
const listeners = new Map<string, Set<() => void>>()
let focusedId: string | null = null

function entry(id: string): LiveState {
  let e = live.get(id)
  if (!e) {
    e = { busy: false, unread: 0, abort: null }
    live.set(id, e)
  }
  return e
}

function emit(id: string) {
  listeners.get(id)?.forEach((l) => l())
}

function subscribeTo(id: string, listener: () => void) {
  let set = listeners.get(id)
  if (!set) {
    set = new Set()
    listeners.set(id, set)
  }
  set.add(listener)
  return () => set!.delete(listener)
}

export function runningConversations(): Set<string> {
  const s = new Set<string>()
  live.forEach((e, id) => e.busy && s.add(id))
  return s
}

export function unreadCounts(): Record<string, number> {
  const out: Record<string, number> = {}
  live.forEach((e, id) => e.unread > 0 && (out[id] = e.unread))
  return out
}

export function focusConversation(id: string | null) {
  focusedId = id
  if (!id) return
  const e = live.get(id)
  if (e && e.unread !== 0) {
    e.unread = 0
    emit(id)
  }
}

function useLive(id: string) {
  return useSyncExternalStore(
    useCallback((l) => subscribeTo(id, l), [id]),
    () => entry(id),
  )
}

export function useConversationLoop(conversationId: string | null, conversation: Conversation | undefined) {
  const liveState = useLive(conversationId ?? '__none__')
  const promotedRef = useRef<Set<string>>(new Set())

  const ask = useCallback(
    async (question: string, model: string, onPanel: (p: PanelRequest) => void) => {
      if (!conversationId || !question.trim()) return
      const e = entry(conversationId)
      if (e.busy) return
      const trimmed = question.trim()
      const history = conversation?.history ?? []
      const turnId = appendTurn(conversationId, {
        id: '',
        question: trimmed,
        text: '',
        steps: [],
        status: 'running',
      }, history)

      const controller = new AbortController()
      e.busy = true
      e.abort = controller
      promotedRef.current = new Set()
      emit(conversationId)

      const patch = (fn: (t: LoopTurn) => LoopTurn) => patchTurn(conversationId, turnId, fn)

      try {
        const result = await runAgent(trimmed, {
          model,
          history,
          signal: controller.signal,
          events: {
            onText: (chunk) => patch((t) => ({ ...t, text: t.text + chunk })),
            onStep: (step: AgentStep) => {
              patch((t) => ({
                ...t,
                steps: t.steps.some((s) => s.id === step.id)
                  ? t.steps.map((s) => (s.id === step.id ? step : s))
                  : [...t.steps, step],
              }))
              if (step.status === 'done' && !promotedRef.current.has(step.id)) {
                const panel = panelForStep(step)
                if (panel) {
                  promotedRef.current.add(step.id)
                  onPanel(panel)
                }
              }
            },
          },
        })
        if (result.stoppedReason === 'max-turns') {
          patch((t) => ({
            ...t,
            status: 'error',
            stoppedReason: result.stoppedReason,
            error: 'Reached the reasoning-step limit before finishing.',
          }))
        } else {
          patch((t) => ({ ...t, status: 'done', stoppedReason: result.stoppedReason }))
        }
        replaceHistory(conversationId, result.history)
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          patch((t) => ({ ...t, status: 'done', stoppedReason: 'aborted' }))
        } else {
          patch((t) => ({ ...t, status: 'error', error: explainAgentError(err) }))
        }
      } finally {
        e.busy = false
        e.abort = null
        if (focusedId !== conversationId) e.unread += 1
        emit(conversationId)
      }
    },
    [conversationId, conversation?.history],
  )

  const stop = useCallback(() => {
    if (!conversationId) return
    const e = entry(conversationId)
    e.abort?.abort()
  }, [conversationId])

  useEffect(() => () => {
  }, [])

  return { liveState, ask, stop }
}
