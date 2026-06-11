import { getStoredTokens } from '@/shared/lib/api-client'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

/** One streamed step of an agent run (mirrors the plugin's agent.Event). */
export interface ChatEvent {
  kind: 'tool_call' | 'tool_result' | 'final' | 'error'
  step?: number
  tool?: string
  args?: unknown
  output?: string
  isError?: boolean
  text?: string
}

/** A navigation directive emitted by the agent (:::navigate … :::). */
export interface NavAction {
  label: string
  destination: string
  filters?: { field: string; operator: string; value?: unknown }[]
  time?: string
}

/**
 * Streams the SOC-AI chat agent over SSE. The backend (/soc-ai/chat) proxies the
 * plugin's agent and emits tool_call / tool_result / final / error events. Uses
 * fetch + ReadableStream because the shared axios client can't stream.
 */
export async function streamChat(
  body: { task: string; page?: string; lang?: string },
  onEvent: (e: ChatEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const tokens = getStoredTokens()
  const res = await fetch(`${API_URL}/soc-ai/chat`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(tokens?.access_token ? { Authorization: `Bearer ${tokens.access_token}` } : {}),
    },
    body: JSON.stringify(body),
    signal,
  })

  if (!res.ok || !res.body) {
    let msg = `Request failed (${res.status})`
    try {
      const j = (await res.json()) as { error?: string }
      if (j?.error) msg = j.error
    } catch {
      /* non-JSON error body */
    }
    throw new Error(msg)
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    let idx: number
    // SSE frames are separated by a blank line.
    while ((idx = buf.indexOf('\n\n')) >= 0) {
      const frame = buf.slice(0, idx)
      buf = buf.slice(idx + 2)
      const ev = parseFrame(frame)
      if (ev) onEvent(ev)
    }
  }
}

function parseFrame(frame: string): ChatEvent | null {
  let data = ''
  for (const line of frame.split('\n')) {
    if (line.startsWith('data:')) data += line.slice(5).trim()
  }
  if (!data) return null
  try {
    return JSON.parse(data) as ChatEvent
  } catch {
    return null
  }
}

const NAV_RE = /:::navigate\s*\n([\s\S]*?)\n:::/g

/**
 * Pulls navigation directives out of the final answer and returns the cleaned
 * text plus the parsed actions (rendered as buttons by the panel).
 */
export function extractNavigation(text: string): { text: string; actions: NavAction[] } {
  const actions: NavAction[] = []
  const cleaned = text.replace(NAV_RE, (_m, json: string) => {
    try {
      const a = JSON.parse(json.trim()) as NavAction
      if (a && a.destination) actions.push(a)
    } catch {
      /* ignore malformed directive */
    }
    return ''
  })
  return { text: cleaned.replace(/\n{3,}/g, '\n\n').trim(), actions }
}
