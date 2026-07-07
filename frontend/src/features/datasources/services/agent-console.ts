import i18n from '@/shared/i18n'
import { getStoredTokens } from '@/shared/lib/api-client'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

/** A message streamed back from the agent command WebSocket. */
interface ConsoleMessage {
  type: 'output' | 'error' | 'ready'
  data?: string
  message?: string
}

export interface CommandPayload {
  command: string
  shell: string
  originType?: string
  originId?: string
  reason?: string
}

export interface SessionHandlers {
  onOutput: (data: string) => void
  onError: (message: string) => void
  /** Server finished the current command and is ready for the next. */
  onReady: () => void
  /** Socket closed (network drop, server restart, or explicit close). */
  onClose?: () => void
}

export interface ConsoleSession {
  /** Send a command over the open socket. No-op if the socket isn't open. */
  send: (payload: CommandPayload) => void
  /** Close the socket. Idempotent. */
  close: () => void
}

/**
 * Opens a persistent WebSocket to the backend agent command stream
 * (GET /soar/ws/command/:agentId). One socket, many commands: each `send()`
 * writes a JSON command frame, the server emits output frames, then a "ready"
 * frame to unblock the next send. The socket stays open until close() or
 * network failure.
 *
 * The JWT travels as a query param because browsers can't set WS headers; the
 * backend handler accepts `?token=` and verifies it.
 */
export function openAgentConsole(agentId: string, h: SessionHandlers): ConsoleSession {
  const token = getStoredTokens()?.access_token ?? ''
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${window.location.host}${API_URL}/soar/ws/command/${encodeURIComponent(agentId)}?token=${encodeURIComponent(token)}`

  const ws = new WebSocket(url)
  let closed = false

  ws.onmessage = (e) => {
    let msg: ConsoleMessage
    try {
      msg = JSON.parse(e.data as string) as ConsoleMessage
    } catch {
      return
    }
    if (msg.type === 'output') h.onOutput(msg.data ?? '')
    else if (msg.type === 'error') h.onError(msg.message ?? i18n.t('datasources.console.cmdError'))
    else if (msg.type === 'ready') h.onReady()
  }
  ws.onerror = () => h.onError(i18n.t('datasources.console.connError'))
  ws.onclose = () => {
    closed = true
    h.onClose?.()
  }

  return {
    send: (payload) => {
      if (ws.readyState !== WebSocket.OPEN) return
      ws.send(
        JSON.stringify({
          originType: 'datasource',
          originId: agentId,
          reason: 'Interactive console',
          ...payload,
        }),
      )
    },
    close: () => {
      if (closed) return
      closed = true
      try {
        ws.close()
      } catch {
        /* noop */
      }
    },
  }
}
