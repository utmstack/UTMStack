import i18n from '@/shared/i18n'
import { getStoredTokens } from '@/shared/lib/api-client'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'

/** A message streamed back from the agent command WebSocket. */
interface ConsoleMessage {
  type: 'output' | 'error' | 'done'
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

export interface ConsoleHandlers {
  onOutput: (data: string) => void
  onError: (message: string) => void
  onDone: () => void
}

/**
 * Opens a WebSocket to the backend agent command stream
 * (GET /soar/ws/command/:agentId), sends ONE command, and streams its output.
 * The protocol is one command per connection — the server emits output chunks,
 * then a "done" message, then closes. Returns a cancel fn that closes the socket.
 *
 * The JWT travels as a query param because browsers can't set WS headers; the
 * backend handler accepts `?token=` and verifies it.
 */
export function runAgentCommand(agentId: string, payload: CommandPayload, h: ConsoleHandlers): () => void {
  const token = getStoredTokens()?.access_token ?? ''
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${window.location.host}${API_URL}/soar/ws/command/${encodeURIComponent(agentId)}?token=${encodeURIComponent(token)}`

  const ws = new WebSocket(url)
  let finished = false
  const finish = () => {
    if (!finished) {
      finished = true
      h.onDone()
    }
  }

  ws.onopen = () => {
    ws.send(
      JSON.stringify({
        originType: 'datasource',
        originId: agentId,
        reason: 'Interactive console',
        ...payload,
      }),
    )
  }
  ws.onmessage = (e) => {
    let msg: ConsoleMessage
    try {
      msg = JSON.parse(e.data as string) as ConsoleMessage
    } catch {
      return
    }
    if (msg.type === 'output') h.onOutput(msg.data ?? '')
    else if (msg.type === 'error') {
      h.onError(msg.message ?? i18n.t('datasources.console.cmdError'))
      finish()
    } else if (msg.type === 'done') finish()
  }
  ws.onerror = () => {
    h.onError(i18n.t('datasources.console.connError'))
    finish()
  }
  ws.onclose = () => finish()

  return () => {
    try {
      ws.close()
    } catch {
      /* noop */
    }
  }
}
