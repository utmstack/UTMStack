import { useSyncExternalStore } from 'react'
import type { CanvasPanel, Conversation, LoopTurn, PanelGeometry, PanelRequest } from './types'
import { proxyFetch } from './proxy'

const CANVAS_KEY = 'utmstack.ti-analyst.canvases'
// Legacy localStorage for conversations, used only as a fallback when the
// backend is unavailable or as a one-time migration source.
const CONV_KEY = 'utmstack.ti-analyst.conversations'

interface Store {
  conversations: Conversation[]
  canvases: Record<string, CanvasPanel[]>
}

function loadCanvases(): Record<string, CanvasPanel[]> {
  try {
    const raw = localStorage.getItem(CANVAS_KEY)
    return raw ? (JSON.parse(raw) as Record<string, CanvasPanel[]>) : {}
  } catch {
    return {}
  }
}

function readLegacyConversations(): Conversation[] {
  try {
    const raw = localStorage.getItem(CONV_KEY)
    return raw ? (JSON.parse(raw) as Conversation[]) : []
  } catch {
    return []
  }
}

function writeLegacyConversations(conversations: Conversation[]) {
  try {
    localStorage.setItem(CONV_KEY, JSON.stringify(conversations))
  } catch {
    // quota — ignore, the backend is the real store
  }
}

let state: Store = { conversations: [], canvases: loadCanvases() }
const listeners = new Set<() => void>()
let backendReady = false

/**
 * Whether the store has finished hydrating from the backend (or falling back
 * to localStorage). Consumers that auto-create must wait for this — creating
 * into an unhydrated store orphans the conversation behind the backend's list.
 */
export function isConsoleStoreReady(): boolean {
  return backendReady
}

function emit() {
  listeners.forEach((l) => l())
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}
export { subscribe }
function getSnapshot() {
  return state
}

export function useConsoleStore(): Store {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot)
}

function makeId(prefix: string): string {
  return `${prefix}_${Date.now().toString(36)}${Math.random().toString(36).slice(2, 7)}`
}

// ---- backend persistence (best-effort; falls back to localStorage) ----

function toDTO(conv: Conversation) {
  return { id: conv.id, title: conv.title, turns: conv.turns, history: conv.history }
}

interface ConversationDTO {
  id: string
  title: string
  turns: LoopTurn[]
  history: Conversation['history']
  updatedAt: string
}

function mapDTO(dto: ConversationDTO): Conversation {
  const updatedAt = Date.parse(dto.updatedAt)
  return { id: dto.id, title: dto.title, turns: dto.turns ?? [], history: dto.history ?? [], updatedAt: Number.isNaN(updatedAt) ? Date.now() : updatedAt }
}

async function createOnBackend(conv: Conversation) {
  await proxyFetch('/analyst/conversations', { method: 'POST', body: JSON.stringify(toDTO(conv)) })
}

async function putOnBackend(conv: Conversation) {
  await proxyFetch(`/analyst/conversations/${encodeURIComponent(conv.id)}`, { method: 'PUT', body: JSON.stringify(toDTO(conv)) })
}

async function deleteOnBackend(id: string) {
  await proxyFetch(`/analyst/conversations/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

/**
 * Hydrates the store from the backend once. Seeds from legacy localStorage if
 * the backend returns nothing (one-time migration); if the backend is
 * unreachable, loads conversations from localStorage and stays local.
 */
let initPromise: Promise<void> | null = null

export async function initConsoleStore(): Promise<void> {
  if (backendReady) return
  if (initPromise) return initPromise
  initPromise = (async () => {
  try {
    const res = await proxyFetch<{ conversations: ConversationDTO[] }>('/analyst/conversations')
    const dtos = res.conversations ?? []
    let conversations = dtos.map(mapDTO)
    if (!conversations.length) {
      // Empty on the backend — migrate anything the analyst had locally.
      const legacy = readLegacyConversations()
      conversations = legacy
      for (const conv of legacy) await createOnBackend(conv).catch(() => undefined)
    }
    backendReady = true
    state = { ...state, conversations }
    emit()
  } catch {
    // Backend unavailable — degrade to localStorage so the console still works.
    backendReady = true
    state = { ...state, conversations: readLegacyConversations() }
    writeLegacyConversations(state.conversations)
    emit()
  }
  })()
  return initPromise
}

function persist(conv: Conversation, op: 'create' | 'put' | 'delete') {
  if (!backendReady) return
  const p = op === 'create' ? createOnBackend(conv) : op === 'put' ? putOnBackend(conv) : deleteOnBackend(conv.id)
  p.catch(() => {
    // Backend write failed — keep the local store authoritative and mirror to
    // localStorage so a refresh does not lose it.
    if (op !== 'delete') writeLegacyConversations(state.conversations)
  })
}

// ---- conversations ----

function commit(next: Store, changedConv?: Conversation, op?: 'create' | 'put' | 'delete') {
  state = next
  try {
    localStorage.setItem(CANVAS_KEY, JSON.stringify(next.canvases))
  } catch {
    // Canvas layout is best-effort local state; a quota or storage error must
    // not break the conversation store.
  }
  if (changedConv) persist(changedConv, op ?? 'put')
  listeners.forEach((l) => l())
}

export function createConversation(title: string): Conversation {
  const conversation: Conversation = {
    id: makeId('conv'),
    title: title.trim() || 'New conversation',
    updatedAt: Date.now(),
    turns: [],
    history: [],
  }
  commit({ ...state, conversations: [conversation, ...state.conversations] }, conversation, 'create')
  return conversation
}

export function deleteConversation(id: string) {
  const canvases = { ...state.canvases }
  delete canvases[id]
  const removed = state.conversations.find((c) => c.id === id)
  commit(
    {
      conversations: state.conversations.filter((c) => c.id !== id),
      canvases,
    },
    removed,
    'delete',
  )
}

export function patchTurn(conversationId: string, turnId: string, fn: (t: LoopTurn) => LoopTurn) {
  let conv: Conversation | undefined
  commit({
    ...state,
    conversations: state.conversations.map((c) => {
      if (c.id !== conversationId) return c
      conv = { ...c, turns: c.turns.map((t) => (t.id === turnId ? fn(t) : t)) }
      return conv
    }),
  }, conv)
}

export function appendTurn(conversationId: string, turn: LoopTurn, history: Conversation['history']): string {
  const id = makeId('turn')
  let conv: Conversation | undefined
  commit({
    ...state,
    conversations: state.conversations.map((c) => {
      if (c.id !== conversationId) return c
      conv = { ...c, updatedAt: Date.now(), turns: [...c.turns, { ...turn, id }], history }
      return conv
    }),
  }, conv)
  return id
}

export function replaceHistory(conversationId: string, history: Conversation['history']) {
  let conv: Conversation | undefined
  commit({
    ...state,
    conversations: state.conversations.map((c) => {
      if (c.id === conversationId) {
        conv = { ...c, history }
        return conv
      }
      return c
    }),
  }, conv)
}

// ---- canvas panels (localStorage — layout, not worth a backend table) ----

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

export function movePanel(conversationId: string, panelId: string, geo: Partial<PanelGeometry>) {
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

export const noopListener = () => {}
