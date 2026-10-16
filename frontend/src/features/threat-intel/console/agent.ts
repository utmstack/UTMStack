import { getStoredTokens } from '@/shared/lib/api-client'
import { IS_FEDERATION } from '@/shared/config/mode'
import { getCurrentInstanceId } from '@/shared/lib/current-instance'
import type {
  AgentStep,  ChatMessage,
  IntelEntity,
  RelationsResult,
  EntityDetails,
  ModelToolCall,
  ToolPayload,
} from './types'
import { detectIndicatorTypes } from '../lib/report'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'
const BASE = `${API_URL}/threat-intel`
const MAX_TOOL_ROUNDS = 6

const SYSTEM_PROMPT = `You are the ThreatWinds analyst agent, embedded in the UTMStack AI analyst console.

Your job is to investigate indicators using the tools provided, then report findings like a SOC analyst would.

Rules:
- ALWAYS use tools to get facts. Never guess reputation, geolocation or relationships from memory.
- For a specific indicator (IP, domain, hostname, URL, hash) use lookup_indicator FIRST. It resolves the exact value. Do not use search_entities for these — it tokenizes the query and returns thousands of unrelated entities.
- Use search_entities only for names and keywords: malware families, campaigns, filenames.
- Once you have an entity id, use get_entity_details for the dossier and get_relations to pivot to connected infrastructure.
- If lookup_indicator returns nothing, the indicator simply has no intelligence recorded. Say so and stop — do not fall back to full-text search and present unrelated results as if they were matches.
- Chain tools when it helps: a suspicious IP is worth expanding via relations, and its associations are worth checking.
- Reputation runs -3 to +3: -3 Alarming, -2 Worrying, -1 Concerning, 0 Indefinable, 1 Adequate, 2 Good, 3 Exceptional. 0 means no verdict was reached, not that the indicator is safe. -3 is the most severe score there is — treat it as such.
- Accuracy runs 0 to 3: 0 Inaccurate, 1 Imprecise, 2 Fairly precise, 3 Precise. It never goes negative, and it measures confidence in the scoring, not severity.
- Be precise and concise. Lead with the verdict, then the evidence that supports it.
- If the data is thin or ambiguous, say so plainly rather than inflating a conclusion.
- Never invent entity ids. Only use ids returned by tools.`

interface ToolSchema {
  type: 'function'
  function: { name: string; description: string; parameters: Record<string, unknown> }
}

const TOOL_SCHEMAS: ToolSchema[] = [
  {
    type: 'function',
    function: {
      name: 'lookup_indicator',
      description:
        'Exact lookup of one indicator in the ThreatWinds corpus. Returns its record (id, tags, reputation, lastSeen) or reports it is absent.',
      parameters: {
        type: 'object',
        properties: {
          value: { type: 'string', description: 'The exact indicator, e.g. "8.8.8.8"' },
          type: { type: 'string', description: 'Optional indicator type (ip, domain, hostname, url, hash, cve, email-address). Inferred from the value when omitted.' },
        },
        required: ['value'],
      },
    },
  },
  {
    type: 'function',
    function: {
      name: 'search_entities',
      description:
        'Free-text search over the indicator corpus. Returns up to 10 matches with id, type, value, tags, reputation.',
      parameters: {
        type: 'object',
        properties: { query: { type: 'string', description: 'Search terms' } },
        required: ['query'],
      },
    },
  },
  {
    type: 'function',
    function: {
      name: 'get_entity_details',
      description:
        'Full dossier for an entity id returned by search/lookup: attributes, geolocation metadata, associations.',
      parameters: {
        type: 'object',
        properties: { id: { type: 'string', description: 'Entity id' } },
        required: ['id'],
      },
    },
  },
  {
    type: 'function',
    function: {
      name: 'get_relations',
      description: 'Relation graph for an entity id: the connected entities and links.',
      parameters: {
        type: 'object',
        properties: { id: { type: 'string', description: 'Entity id' } },
        required: ['id'],
      },
    },
  },
]

export interface AgentEvents {
  onText?(chunk: string): void
  onStep?(step: AgentStep): void
}

async function authHeaders(): Promise<Record<string, string>> {
  const h: Record<string, string> = { 'Content-Type': 'application/json' }
  const tokens = getStoredTokens()
  if (tokens?.access_token) h.Authorization = `Bearer ${tokens.access_token}`
  if (IS_FEDERATION) {
    const id = getCurrentInstanceId()
    if (id != null) h['X-UTM-Instance'] = String(id)
  }
  return h
}

async function proxyFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { ...init, headers: await authHeaders() })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return (await res.json()) as T
}

interface SearchItem {
  id: string
  type: string
  attributes: Record<string, string>
  tags?: string[]
  lastSeen?: string
  reputation: number
  accuracy: number
}

function itemValue(item: { type: string; attributes: Record<string, string> }): string {
  return item.attributes[item.type] ?? Object.values(item.attributes)[0] ?? ''
}

function toItem(item: SearchItem): IntelEntity {
  return {
    id: item.id,
    type: item.type,
    value: itemValue(item),
    reputation: item.reputation ?? 0,
    accuracy: item.accuracy ?? 0,
    tags: item.tags ?? [],
    lastSeen: item.lastSeen,
  }
}

function detailsToGeo(d: EntityDetails): { lat: number; lng: number; radius: number; ip?: string }[] {
  return (d.geolocations ?? []).map((g) => ({ lat: g.lat, lng: g.lng, radius: g.radius, ip: g.ip }))
}

export interface RawDetails {
  attributes: {
    id: string
    type: string
    label: string
    description?: string
    tags: string[]
    first_seen: string
    last_seen: string
    reputation_score: number
    reputation: string
    best_reputation_score: number
    best_reputation: string
    worst_reputation_score: number
    worst_reputation: string
    accuracy_score: number
    accuracy: string
  }
  metadata: Record<string, string> | null
  extended_metadata?: unknown[]
  latest_associations: { id: string; type: string; value?: string; label?: string; tags?: string[]; reputation?: number; accuracy?: number }[]
  geolocations: { latitude: number; longitude: number; accuracy_radius: number }[]
}

export function normalizeDetails(raw: RawDetails): EntityDetails {
  const a = raw.attributes
  return {
    attributes: {
      id: a.id,
      type: a.type,
      value: a.label,
      description: a.description,
      tags: a.tags ?? [],
      first_seen: a.first_seen,
      last_seen: a.last_seen,
      reputation_score: a.reputation_score,
      reputation: a.reputation,
      best_reputation_score: a.best_reputation_score,
      best_reputation: a.best_reputation,
      worst_reputation_score: a.worst_reputation_score,
      worst_reputation: a.worst_reputation,
      accuracy_score: a.accuracy_score,
      accuracy: a.accuracy,
    },
    metadata: raw.metadata ?? null,
    latest_associations: (raw.latest_associations ?? []).map((x) => ({
      id: x.id,
      type: x.type,
      value: x.value ?? x.label ?? x.id,
      reputation: x.reputation ?? 0,
      accuracy: x.accuracy ?? 0,
      tags: x.tags ?? [],
    })),
    geolocations: (raw.geolocations ?? []).map((g) => ({
      lat: g.latitude,
      lng: g.longitude,
      radius: g.accuracy_radius,
    })),
  }
}

export interface RawRelation {
  id: string
  type: string
  label: string
  reputation_score: number
  reputation: string
}
export function relationsToGraph(id: string, self: string, rels: RawRelation[]): RelationsResult {
  const nodes: { id: string; type: string; value: string; reputation: number | null }[] = [
    { id, type: 'root', value: self || id, reputation: null },
  ]
  const edges: { source: string; target: string }[] = []
  for (const r of rels) {
    nodes.push({ id: r.id, type: r.type, value: r.label, reputation: r.reputation_score })
    edges.push({ source: id, target: r.id })
  }
  return { nodes, edges }
}

async function runTool(name: string, args: Record<string, unknown>, step: AgentStep): Promise<string> {
  switch (name) {
    case 'lookup_indicator': {
      const value = String(args.value ?? '')
      const type = String(args.type || '') || detectIndicatorTypes(value)[0]
      if (!type) {
        step.status = 'error'
        step.error = 'could not infer an indicator type'
        return `Tool failed: could not infer a type for "${value}".`
      }
      try {
        const item = await proxyFetch<SearchItem>('/entity/lookup', {
          method: 'POST',
          body: JSON.stringify({ type, value }),
        })
        step.payload = { kind: 'lookup', value, entity: toItem(item) }
        return JSON.stringify(toItem(item))
      } catch {
        step.payload = { kind: 'lookup', value, entity: null }
        return `No threat intelligence recorded for "${value}".`
      }
    }
    case 'search_entities': {
      const res = await proxyFetch<{ items: number; results: SearchItem[] }>('/ai/search', {
        method: 'POST',
        body: JSON.stringify({ query: String(args.query), page: 0, limit: 10 }),
      })
      const results = (res.results ?? []).map(toItem)
      step.payload = { kind: 'search', query: String(args.query), items: res.items ?? results.length, results }
      return JSON.stringify({ items: res.items, results })
    }
    case 'get_entity_details': {
      const id = String(args.id).replace(/[.\//\\]/g, '')
      const raw = await proxyFetch<RawDetails>(`/entity/${encodeURIComponent(id)}`)
      const details = normalizeDetails(raw)
      step.payload = { kind: 'details', id, details }
      return JSON.stringify(details)
    }
    case 'get_relations': {
      const id = String(args.id).replace(/[.\//\\]/g, '')
      const rels = await proxyFetch<RawRelation[]>(`/entity/${encodeURIComponent(id)}/relations`)
      const graph = relationsToGraph(id, '', Array.isArray(rels) ? rels : [])
      step.payload = { kind: 'relations', id, graph }
      return JSON.stringify(graph)
    }
    default:
      return `Tool failed: unknown tool ${name}`
  }
}

interface ChatChunk {
  choices?: {
    delta?: {
      content?: string
      tool_calls?: { index: number; id?: string; function?: { name?: string; arguments?: string } }[]
    }
  }[]
}

function sanitizeHistory(messages: ChatMessage[]): ChatMessage[] {
  return messages.map((m) => {
    if (m.role !== 'assistant' || !m.tool_calls?.length) return m
    const calls = m.tool_calls.map((c) => ({
      ...c,
      function: { ...c.function, arguments: typeof c.function.arguments === 'string' ? c.function.arguments : '{}' },
    }))
    return { ...m, tool_calls: calls }
  })
}

export async function runAgent(
  question: string,
  opts: {
    model: string
    history: ChatMessage[]
    signal?: AbortSignal
    events?: AgentEvents
  },
): Promise<{ answer: string; history: ChatMessage[] }> {
  const convo: ChatMessage[] = [...opts.history, { role: 'user', content: question }]
  const emit = opts.events ?? {}

  for (let round = 0; round < MAX_TOOL_ROUNDS; round++) {
    const requestMessages: ChatMessage[] = [
      { role: 'system', content: SYSTEM_PROMPT },
      ...sanitizeHistory(convo),
    ]
    const res = await fetch(`${BASE}/ai/chat`, {
      method: 'POST',
      headers: await authHeaders(),
      body: JSON.stringify({ model: opts.model, messages: requestMessages, tools: TOOL_SCHEMAS, stream: true }),
      signal: opts.signal,
    })
    if (!res.ok || !res.body) {
      let msg = `Request failed (${res.status})`
      try {
        const j = JSON.parse(await res.text())
        const err = j?.error
        msg = (typeof err === 'string' ? err : err?.message) ?? msg
      } catch {
        // non-JSON error body — keep the HTTP status message
      }
      throw new Error(msg)
    }

    let text = ''
    const toolCalls: (ModelToolCall | undefined)[] = []
    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buf = ''
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buf += decoder.decode(value, { stream: true })
      let idx: number
      while ((idx = buf.indexOf('\n\n')) >= 0) {
        const frame = buf.slice(0, idx)
        buf = buf.slice(idx + 2)
        let data = ''
        for (const line of frame.split('\n')) if (line.startsWith('data:')) data += line.slice(5).trim()
        if (!data || data === '[DONE]') continue
        let chunk: ChatChunk
        try {
          chunk = JSON.parse(data)
        } catch {
          continue
        }
        const delta = chunk.choices?.[0]?.delta
        if (!delta) continue
        if (delta.content) {
          text += delta.content
          emit.onText?.(delta.content)
        }
        for (const tc of delta.tool_calls ?? []) {
          const slot = toolCalls[tc.index]
          if (!tc.function?.name) continue
          toolCalls[tc.index] = {
            id: tc.id ?? slot?.id ?? `call_${round}_${tc.index}`,
            type: 'function',
            function: {
              name: tc.function.name,
              arguments: (slot?.function.arguments ?? '') + (tc.function.arguments ?? ''),
            },
          }
        }
      }
    }

    const calls = toolCalls.filter((c): c is ModelToolCall => Boolean(c))
    convo.push({ role: 'assistant', content: text, ...(calls.length ? { tool_calls: calls } : {}) })

    if (calls.length === 0) {
      return { answer: text, history: convo }
    }

    for (const call of calls) {
      let args: Record<string, unknown> = {}
      try {
        args = JSON.parse(call.function.arguments || '{}')
      } catch {
        // malformed tool args — call with {} rather than crashing the turn
      }
      const step: AgentStep = {
        id: call.id,
        toolName: call.function.name,
        args,
        status: 'running',
      }
      emit.onStep?.({ ...step })
      const t0 = performance.now()
      let result: string
      try {
        result = await runTool(call.function.name, args, step)
      } catch (e) {
        step.status = 'error'
        step.error = e instanceof Error ? e.message : String(e)
        result = `Tool failed: ${step.error}`
      }
      step.durationMs = performance.now() - t0
      if (step.status !== 'error') step.status = 'done'
      step.result = result
      emit.onStep?.({ ...step })
      convo.push({ role: 'tool', tool_call_id: call.id, content: result })
    }
  }
  throw new Error('Reached the reasoning-step limit before finishing.')
}

export { detailsToGeo, type ToolPayload }
