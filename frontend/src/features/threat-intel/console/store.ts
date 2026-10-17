import { useSyncExternalStore } from 'react'
import type { CanvasPanel, Conversation, LoopTurn, PanelGeometry, PanelRequest } from './types'

const CONV_KEY = 'utmstack.ti-analyst.conversations'
const CANVAS_KEY = 'utmstack.ti-analyst.canvases'

interface Store {
  conversations: Conversation[]
  canvases: Record<string, CanvasPanel[]>
}

function load(): Store {
  const read = <T>(key: string, fallback: T): T => {
    try {
      const raw = localStorage.getItem(key)
      return raw ? (JSON.parse(raw) as T) : fallback
    } catch {
      return fallback
    }
  }
  return {
    conversations: read<Conversation[]>(CONV_KEY, []),
    canvases: read<Record<string, CanvasPanel[]>>(CANVAS_KEY, {}),
  }
}

let state: Store = load()
const listeners = new Set<() => void>()

function commit(next: Store) {
  state = next
  try {
    localStorage.setItem(CONV_KEY, JSON.stringify(next.conversations))
    localStorage.setItem(CANVAS_KEY, JSON.stringify(next.canvases))
  } catch {
  }
  listeners.forEach((l) => l())
}

const noop = () => {}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}
function getSnapshot() {
  return state
}

export function useConsoleStore(): Store {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot)
}

function makeId(prefix: string): string {
  return `${prefix}_${Date.now().toString(36)}${Math.random().toString(36).slice(2, 7)}`
}

export function createConversation(title: string): Conversation {
  const conversation: Conversation = {
    id: makeId('conv'),
    title: title.trim() || 'New conversation',
    updatedAt: Date.now(),
    turns: [],
    history: [],
  }
  commit({ ...state, conversations: [conversation, ...state.conversations] })
  return conversation
}

export function deleteConversation(id: string) {
  const canvases = { ...state.canvases }
  delete canvases[id]
  commit({
    conversations: state.conversations.filter((c) => c.id !== id),
    canvases,
  })
}

export function patchTurn(conversationId: string, turnId: string, fn: (t: LoopTurn) => LoopTurn) {
  commit({
    ...state,
    conversations: state.conversations.map((c) =>
      c.id === conversationId
        ? { ...c, turns: c.turns.map((t) => (t.id === turnId ? fn(t) : t)) }
        : c,
    ),
  })
}

export function appendTurn(conversationId: string, turn: LoopTurn, history: Conversation['history']): string {
  const id = makeId('turn')
  commit({
    ...state,
    conversations: state.conversations.map((c) =>
      c.id === conversationId
        ? { ...c, updatedAt: Date.now(), turns: [...c.turns, { ...turn, id }], history }
        : c,
    ),
  })
  return id
}

export function replaceHistory(conversationId: string, history: Conversation['history']) {
  commit({
    ...state,
    conversations: state.conversations.map((c) => (c.id === conversationId ? { ...c, history } : c)),
  })
}

const CASCADE_STEP = 28
const DEFAULT_SIZE = { width: 520, height: 400 }
const MAX_PANELS = 12

export function setPanels(conversationId: string, panels: CanvasPanel[]) {
  commit({ ...state, canvases: { ...state.canvases, [conversationId]: panels } })
}

export function openPanel(conversationId: string, input: PanelRequest): void {
  const prev = state.canvases[conversationId] ?? []
  const existing = prev.find((p) => p.kind === input.kind && p.ref === input.ref)
  let next: CanvasPanel[]
  if (existing) {
    const z = prev.reduce((m, p) => Math.max(m, p.z), 0) + 1
    next = prev.map((p) => (p.id === existing.id ? { ...p, z, minimised: false } : p))
  } else {
    const kept = prev.length >= MAX_PANELS ? prev.slice(1) : prev
    const z = kept.reduce((m, p) => Math.max(m, p.z), 0) + 1
    const index = (kept.length % 6) + 1
    next = [
      ...kept,
      {
        id: makeId('p'),
        kind: input.kind,
        title: input.title,
        ref: input.ref,
        x: CASCADE_STEP * index,
        y: CASCADE_STEP * index,
        z,
        ...DEFAULT_SIZE,
      },
    ]
  }
  setPanels(conversationId, next)
}

export function closePanel(conversationId: string, panelId: string) {
  setPanels(
    conversationId,
    (state.canvases[conversationId] ?? []).filter((p) => p.id !== panelId),
  )
}

export function focusPanel(conversationId: string, panelId: string) {
  const prev = state.canvases[conversationId] ?? []
  const max = prev.reduce((m, p) => Math.max(m, p.z), 0)
  setPanels(
    conversationId,
    prev.map((p) => (p.id === panelId ? { ...p, z: max + 1, minimised: false } : p)),
  )
}

export function movePanel(
  conversationId: string,
  panelId: string,
  geo: Partial<PanelGeometry>,
) {
  const prev = state.canvases[conversationId] ?? []
  setPanels(
    conversationId,
    prev.map((p) => {
      if (p.id !== panelId) return p
      const width = Math.max(300, geo.width ?? p.width)
      const height = Math.max(200, geo.height ?? p.height)
      return { ...p, x: Math.max(0, geo.x ?? p.x), y: Math.max(0, geo.y ?? p.y), width, height }
    }),
  )
}

export function toggleMinimised(conversationId: string, panelId: string) {
  const prev = state.canvases[conversationId] ?? []
  setPanels(
    conversationId,
    prev.map((p) => (p.id === panelId ? { ...p, minimised: !p.minimised } : p)),
  )
}

export function tilePanels(conversationId: string, canvas: { width: number; height: number }) {
  const prev = state.canvases[conversationId] ?? []
  const visible = prev.filter((p) => !p.minimised)
  if (!visible.length) return
  const columns = Math.ceil(Math.sqrt(visible.length))
  const rows = Math.ceil(visible.length / columns)
  const width = Math.max(300, Math.floor(canvas.width / columns) - 8)
  const height = Math.max(200, Math.floor(canvas.height / rows) - 8)
  const placed = new Map<string, { x: number; y: number; width: number; height: number }>()
  visible.forEach((panel, i) =>
    placed.set(panel.id, {
      x: (i % columns) * (width + 8),
      y: Math.floor(i / columns) * (height + 8),
      width,
      height,
    }),
  )
  setPanels(
    conversationId,
    prev.map((p) => (placed.has(p.id) ? { ...p, ...placed.get(p.id)! } : p)),
  )
}

export function confinePanels(conversationId: string, canvas: { width: number; height: number }) {
  const prev = state.canvases[conversationId] ?? []
  let changed = false
  const next = prev.map((p) => {
    const width = Math.max(300, Math.min(p.width, canvas.width))
    const height = p.minimised ? p.height : Math.max(200, Math.min(p.height, canvas.height))
    const x = Math.max(0, Math.min(p.x, canvas.width - width))
    const y = Math.max(0, Math.min(p.y, canvas.height - height))
    if (x !== p.x || y !== p.y || width !== p.width || height !== p.height) changed = true
    return { ...p, width, height, x, y }
  })
  if (changed) setPanels(conversationId, next)
}

export function closeAllPanels(conversationId: string) {
  setPanels(conversationId, [])
}

export function byRecency(a: Conversation, b: Conversation): number {
  return b.updatedAt - a.updatedAt
}

export const noopListener = noop
