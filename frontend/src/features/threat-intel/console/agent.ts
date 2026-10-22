import { getStoredTokens } from '@/shared/lib/api-client'
import { IS_FEDERATION } from '@/shared/config/mode'
import { getCurrentInstanceId } from '@/shared/lib/current-instance'
import type {
  AgentStep,
  ChatMessage,
  IntelEntity,
  RelationsResult,
  EntityDetails,
  ModelToolCall,
  ToolPayload,
} from './types'
import {
  countEntitiesByType,
  detectIndicatorTypes,
  reportIoc,
} from '../lib/report'
import {
  addCaseNote,
  addWatchlistItem,
  createCase,
  createWatchlist,
  linkEntities,
  listAlerts,
  listCases,
  listWatchlists,
  pinEntityToCase,
  type Case,
  type CaseStatus,
  type Watchlist,
} from '../lib/casework'
import {
  compactMessages,
  estimateTokens,
  renderForSummary,
  SUMMARY_INPUT_RATIO,
  DEFAULT_COMPACTION,
  type CompactionConfig,
  type Summarizer,
} from './compaction'

const API_URL = import.meta.env.VITE_API_URL || '/api/v1'
const BASE = `${API_URL}/threat-intel`
const MAX_TOOL_ROUNDS = 25

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
- Never invent entity ids. Only use ids returned by tools.

Reporting is different from investigating. report_ioc and link_entities WRITE to the shared ThreatWinds corpus where other instances see it, and it is not easy to retract:
- Only call them when the analyst explicitly asks to report, submit, flag or link.
- Never call them to "record" something you found while investigating, and never call them speculatively.
- State the type, value and reputation you are about to submit before calling it. Reputation uses the same -3..3 scale you read.

The workspace is the analyst's own: cases, watchlists and the alerts they raise. Everything in it can be undone, but it belongs to the tenant.
- list_cases, list_watchlists and list_alerts are ordinary reads. Use them freely — checking whether an indicator is already in a case is often the most useful thing you can say about it.
- create_case, pin_to_case, add_case_note, create_watchlist and add_watchlist_item change that workspace. Only on explicit request, never as a side effect of investigating. Finding something alarming is not permission to open a case about it; offer, and wait.
- Prefer adding to what already exists. Call list_cases before create_case and list_watchlists before create_watchlist, and say which you are adding to.
- Pin entities by the id a tool returned, never by a value you retyped.
- Case severity runs 0 to 4: 0 informational, 1 low, 2 medium, 3 high, 4 critical. Case status is open, in_progress or closed. A watchlist rule kind is entity, value, tag or type.`

export interface AgentEvents {
  onText?(chunk: string): void
  onStep?(step: AgentStep): void
  /** Older turns were folded into a summary to stay within the input budget. */
  onCompaction?(info: { droppedGroups: number; tokensBefore: number; tokensAfter: number }): void
}

export interface AgentRunResult {
  answer: string
  history: ChatMessage[]
  stoppedReason: 'complete' | 'max-turns' | 'aborted'
}

let stepCounter = 0
const nextStepId = () => `step_${++stepCounter}`

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

async function proxyFetch<T>(path: string, init: RequestInit = {}, signal?: AbortSignal): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { ...init, headers: await authHeaders(), signal })
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = JSON.parse(await res.text())
      const err = body?.error
      msg = (typeof err === 'string' ? err : err?.message) ?? msg
    } catch {
      // non-JSON error body — keep the HTTP status message
    }
    throw new ProxyError(msg, res.status)
  }
  return (await res.json()) as T
}

/** Distinguished from generic failures so lookup can separate "the corpus said no" from "we never got an answer". */
export class ProxyError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ProxyError'
    this.status = status
  }
}

/** Turns transport failures into something an analyst can act on. */
export function explainAgentError(err: unknown): string {
  if (err instanceof ProxyError) {
    if (err.status === 503) {
      return 'The ThreatWinds proxy is not configured on this instance (or is down). Ask the platform admin to set it up; search still works locally.'
    }
    if (err.status === 401 || err.status === 403) {
      return 'Your session was rejected by the UTMStack backend. Sign in again.'
    }
    if (err.status === 504) {
      return 'ThreatWinds did not answer in time. Try again; very deep relation graphs can exceed the proxy timeout.'
    }
    return err.message
  }
  return err instanceof Error ? err.message : String(err)
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

/** Timestamps, trimmed to the second — nanosecision precision is noise and costs tokens. */
function at(iso: string | null | undefined): string | null {
  if (!iso) return null
  const ms = Date.parse(iso)
  return Number.isNaN(ms) ? null : new Date(ms).toISOString().replace(/\.\d+Z$/, 'Z')
}

function seenText(firstSeen: string | null | undefined, lastSeen: string | null | undefined): string | null {
  const first = at(firstSeen)
  const last = at(lastSeen)
  if (first && last) return first === last ? `seen=${first}` : `first_seen=${first} last_seen=${last}`
  if (last) return `last_seen=${last}`
  if (first) return `first_seen=${first}`
  return null
}

function describeEntity(e: IntelEntity): string {
  const parts = [`${e.type}:${e.value}`, `id=${e.id}`, `reputation=${e.reputation}`]
  if (e.tags.length) parts.push(`tags=${e.tags.join('|')}`)
  const seen = seenText(null, e.lastSeen)
  if (seen) parts.push(seen)
  return parts.join(' ')
}

const str = (v: unknown, fallback = ''): string => (typeof v === 'string' ? v : fallback)
const int = (v: unknown, fallback: number): number =>
  typeof v === 'number' && Number.isFinite(v) ? Math.trunc(v) : fallback

/** Models emit nested associations as loosely-typed objects, so each level is validated. */
function parseAssociations(raw: unknown): import('../lib/report').ReportAssociationInput[] | undefined {
  if (!Array.isArray(raw) || raw.length === 0) return undefined
  const parsed = raw
    .map((entry): import('../lib/report').ReportAssociationInput | null => {
      if (!entry || typeof entry !== 'object') return null
      const e = entry as Record<string, unknown>
      const value = str(e.value)
      if (!value) return null
      return {
        value,
        type: str(e.type) || undefined,
        mode: e.mode === 'aggregation' ? 'aggregation' : 'association',
        associations: parseAssociations(e.associations),
      }
    })
    .filter((a): a is import('../lib/report').ReportAssociationInput => a !== null)
  return parsed.length ? parsed : undefined
}

interface ToolDef {
  schema: {
    name: string
    description: string
    parameters: Record<string, unknown>
  }
  run(args: Record<string, unknown>, signal?: AbortSignal): Promise<{ payload: ToolPayload; summarize: () => string }>
}

const TOOLS: Record<string, ToolDef> = {
  lookup_indicator: {
    schema: {
      name: 'lookup_indicator',
      description:
        'Resolve ONE exact indicator (IP, CIDR, domain, hostname, URL, or file hash) to its entity. This is the correct first step whenever the analyst names a specific indicator — full-text search tokenizes the value and returns thousands of unrelated entities instead.',
      parameters: {
        type: 'object',
        properties: {
          value: { type: 'string', description: 'The exact indicator, e.g. "8.8.8.8".' },
          type: { type: 'string', description: 'Optional entity type (ip, domain, hostname, url, md5, sha256...). Inferred from the value when omitted.' },
        },
        required: ['value'],
      },
    },
    async run(args, signal) {
      const value = str(args.value)
      const types = str(args.type) ? [str(args.type)] : detectIndicatorTypes(value)
      if (!types.length) {
        return {
          payload: { kind: 'lookup', value, entity: null, rejected: [] },
          summarize: () =>
            `"${value}" has no recognisable indicator shape. If it is a name (malware, campaign, filename) use search_entities.`,
        }
      }
      // Try each candidate type; remember which were refused so a miss can
      // be distinguished from "the corpus was asked and had no answer".
      const rejected: string[] = []
      let found: SearchItem | null = null
      for (const type of types) {
        try {
          found = await proxyFetch<SearchItem>('/entity/lookup', {
            method: 'POST',
            body: JSON.stringify({ type, value }),
          }, signal)
          break
        } catch (err) {
          if (err instanceof ProxyError && (err.status === 404 || err.status === 400)) {
            rejected.push(type)
            continue
          }
          if (err instanceof ProxyError && err.status === 404) continue
          throw err
        }
      }
      if (found) {
        const entity = toItem(found)
        return {
          payload: { kind: 'lookup', value, entity, rejected: [] },
          summarize: () => `Resolved ${value} -> ${describeEntity(entity)}. Use this id with get_entity_details or get_relations.`,
        }
      }
      const allRejected = rejected.length === types.length
      return {
        payload: { kind: 'lookup', value, entity: null, rejected: allRejected ? rejected : [] },
        summarize: () =>
          allRejected
            ? `"${value}" was not looked up: the API rejected it as ${rejected.join(' and ')}, ` +
              `which are the only types its shape suggests. This is NOT evidence of absence — nothing was searched. ` +
              `If you know the type (malware, filename, campaign names have no recognisable shape), ` +
              `call lookup_indicator again with it set. Otherwise use search_entities to find it by name.`
            : `"${value}" is not present in the corpus. It has no recorded threat intelligence.`,
      }
    },
  },

  search_entities: {
    schema: {
      name: 'search_entities',
      description:
        'Full-text search for entities by NAME or KEYWORD — malware families, campaign names, filenames, free text. Do NOT use it to look up a specific IP, domain or hash: the query is tokenized, so "8.8.8.8" returns thousands of unrelated addresses. Use lookup_indicator for exact indicators.',
      parameters: {
        type: 'object',
        properties: {
          query: { type: 'string', description: 'Search terms, e.g. "emotet".' },
          limit: { type: 'integer', description: 'Max results, 1-50. Default 15.' },
        },
        required: ['query'],
      },
    },
    async run(args, signal) {
      const query = str(args.query)
      const limit = Math.min(Math.max(int(args.limit, 15), 1), 50)
      const res = await proxyFetch<{ items: number; results: SearchItem[] }>('/ai/search', {
        method: 'POST',
        body: JSON.stringify({ query, page: 0, limit }),
      }, signal)
      const results = (res.results ?? []).map(toItem)
      return {
        payload: { kind: 'search', query, items: res.items ?? results.length, results },
        summarize: () => {
          if (!results.length) return `No entities matched "${query}".`
          const lines = results.slice(0, 15).map((e) => `- ${describeEntity(e)}`)
          return `${res.items ?? results.length} total matches for "${query}". Top ${lines.length}:\n${lines.join('\n')}`
        },
      }
    },
  },

  get_entity_details: {
    schema: {
      name: 'get_entity_details',
      description:
        'Full dossier for one entity id: reputation and accuracy scores with their labels, when it was first and last reported, description, metadata (ASN, city, country), associated entities, and geolocations. This is where recency questions are answered — "when was this last seen" is first_seen/last_seen here. Requires an exact entity id from search_entities, lookup_indicator or get_relations.',
      parameters: {
        type: 'object',
        properties: {
          id: { type: 'string', description: 'Exact entity id, e.g. "ip-4f24afca…".' },
        },
        required: ['id'],
      },
    },
    async run(args, signal) {
      const id = str(args.id).replace(/[.\//\\]/g, '')
      const raw = await proxyFetch<RawDetails>(`/entity/${encodeURIComponent(id)}`, {}, signal)
      const details = normalizeDetails(raw)
      return {
        payload: { kind: 'details', id, details },
        summarize: () => {
          const a = details.attributes
          const lines = [
            `${a.type}:${a.value}`,
            `reputation=${a.reputation}(${a.reputation_score}) accuracy=${a.accuracy}(${a.accuracy_score})`,
            `worst=${a.worst_reputation} best=${a.best_reputation}`,
          ]
          const seen = seenText(a.first_seen, a.last_seen)
          if (seen) lines.push(seen)
          if (details.metadata) {
            const meta = Object.entries(details.metadata)
              .map(([k, v]) => `${k}=${v}`)
              .join(' ')
            if (meta) lines.push(`metadata: ${meta}`)
          }
          const assoc = details.latest_associations ?? []
          if (assoc.length) {
            lines.push(
              `associations (${assoc.length}): ` +
                assoc
                  .slice(0, 10)
                  .map((x) => `${x.type}:${x.value} id=${x.id}`)
                  .join('; '),
            )
          }
          const geo = details.geolocations ?? []
          if (geo.length) {
            lines.push(
              `geolocations (${geo.length}): ` +
                geo.slice(0, 5).map((g) => `${g.lat},${g.lng} r=${g.radius}m`).join('; '),
            )
          }
          return lines.join('\n')
        },
      }
    },
  },

  get_relations: {
    schema: {
      name: 'get_relations',
      description: 'Relationship graph around an entity: connected entities and the edges between them. Use to pivot from one indicator to infrastructure it touches.',
      parameters: {
        type: 'object',
        properties: {
          id: { type: 'string', description: 'Exact entity id.' },
        },
        required: ['id'],
      },
    },
    async run(args, signal) {
      const id = str(args.id).replace(/[.\//\\]/g, '')
      const rels = await proxyFetch<RawRelation[]>(`/entity/${encodeURIComponent(id)}/relations`, {}, signal)
      const graph = relationsToGraph(id, '', Array.isArray(rels) ? rels : [])
      return {
        payload: { kind: 'relations', id, graph },
        summarize: () => {
          const { nodes, edges } = graph
          if (nodes.length <= 1) return 'No related entities found.'
          const listed = nodes
            .slice(1, 21)
            .map((n) => `- ${n.type}:${n.value} id=${n.id} reputation=${n.reputation ?? 'n/a'}`)
          return `${nodes.length} nodes, ${edges.length} edges.\n${listed.join('\n')}`
        },
      }
    },
  },

  get_geolocations: {
    schema: {
      name: 'get_geolocations',
      description: 'Geographic locations attributed to an entity, with accuracy radius. Only IP-like entities usually have these.',
      parameters: {
        type: 'object',
        properties: {
          id: { type: 'string', description: 'Exact entity id.' },
        },
        required: ['id'],
      },
    },
    async run(args, signal) {
      const id = str(args.id).replace(/[.\//\\]/g, '')
      const raw = await proxyFetch<RawDetails>(`/entity/${encodeURIComponent(id)}`, {}, signal)
      const details = normalizeDetails(raw)
      const locations = detailsToGeo(details)
      return {
        payload: { kind: 'geo', id, locations },
        summarize: () => {
          if (!locations.length) return 'No geolocation data for this entity.'
          return locations.map((g) => `- ${g.lat}, ${g.lng} radius=${g.radius}m`).join('\n')
        },
      }
    },
  },

  count_entities_by_type: {
    schema: {
      name: 'count_entities_by_type',
      description: 'Corpus-wide counts of how many entities exist per type. Use for scale and coverage questions, not for investigating a specific indicator.',
      parameters: {
        type: 'object',
        properties: {
          types: { type: 'array', items: { type: 'string' }, description: 'Entity types to count. Omit for all known types.' },
        },
      },
    },
    async run(args, _signal) {
      const requested = Array.isArray(args.types)
        ? args.types.filter((x): x is string => typeof x === 'string')
        : []
      const buckets = await countEntitiesByType(requested)
      return {
        payload: { kind: 'counts', buckets },
        summarize: () => {
          if (!buckets.length) return 'No counts returned.'
          return buckets.map((b) => `${b.key}: ${b.count.toLocaleString()}`).join('\n')
        },
      }
    },
  },

  report_ioc: {
    schema: {
      name: 'report_ioc',
      description:
        'Report a NEW indicator of compromise back into the ThreatWinds corpus. This WRITES data visible to all instances sharing this proxy, so only call it when the user explicitly asks to report, submit or flag an indicator — never as part of an investigation.',
      parameters: {
        type: 'object',
        properties: {
          value: { type: 'string', description: 'The indicator, e.g. "203.0.113.1".' },
          type: {
            type: 'string',
            description: 'Entity type (ip, domain, hostname, url, md5, sha256...). Inferred when omitted. MUST be set explicitly for names — malware, filenames, campaigns — because those have no recognisable shape and are otherwise rejected.',
          },
          reputation: {
            type: 'integer',
            description:
              'Reputation from -3 to 3, the same scale used when reading: -3 Alarming, -2 Worrying, -1 Concerning, 0 Indefinable, 1 Adequate, 2 Good, 3 Exceptional. 0 means no verdict was reached, not that the indicator is neutral or safe — omit reputation entirely rather than submitting 0 to mean "probably fine".',
          },
          tags: { type: 'array', items: { type: 'string' }, description: 'Short labels, e.g. ["scanner", "bruteforce"]' },
          description: { type: 'string', description: 'Why this indicator is being reported.' },
          associations: {
            type: 'array',
            description: 'Related entities to create and link in the same submission, e.g. the domain an IP resolves to or the malware a file contains.',
            items: {
              type: 'object',
              properties: {
                value: { type: 'string', description: 'The related indicator.' },
                type: { type: 'string', description: 'Entity type. Inferred only for network indicators and hashes; set explicitly for names.' },
                mode: { type: 'string', enum: ['aggregation', 'association'], description: '"aggregation" when the related entity is a component OF the parent (a file contains malware); "association" for a looser link.' },
              },
              required: ['value'],
            },
          },
        },
        required: ['value'],
      },
    },
    async run(args, _signal) {
      const tags = Array.isArray(args.tags)
        ? args.tags.filter((t): t is string => typeof t === 'string')
        : undefined
      const result = await reportIoc({
          value: str(args.value),
          type: str(args.type) || undefined,
          reputation: typeof args.reputation === 'number' ? args.reputation : undefined,
          tags,
          description: str(args.description) || undefined,
          associations: parseAssociations(args.associations),
        })
      const skipped = result.skippedAssociations
      return {
        payload: { kind: 'report', value: result.value, type: result.type, reputation: result.reputation, skippedAssociations: skipped },
        summarize: () => {
          const lines = [
            `Reported ${result.type}:${result.value} with reputation ${result.reputation ?? 'not stated'}${result.tags.length ? ` and tags ${result.tags.join(', ')}` : ''}.`,
          ]
          if (skipped.length) {
            lines.push(
              `WARNING: these associations were NOT recorded because no entity type was given and none could be inferred: ${skipped.join(', ')}. Retry them with an explicit "type" (e.g. "malware" or "filename"), and tell the analyst they were not saved.`,
            )
          }
          lines.push('Ingestion is asynchronous, so it will not be immediately searchable.')
          return lines.join(' ')
        },
      }
    },
  },

  link_entities: {
    schema: {
      name: 'link_entities',
      description:
        'Associate two entities that ALREADY exist in the corpus, by their entity ids. Use this to record a relationship between things already in the corpus; use report_ioc with associations to create and link new entities in one go. WRITES data — only on explicit request.',
      parameters: {
        type: 'object',
        properties: {
          entityID: { type: 'string', description: 'Source entity id, e.g. "ip-4f24afca…".' },
          relatedEntityID: { type: 'string', description: 'Target entity id.' },
        },
        required: ['entityID', 'relatedEntityID'],
      },
    },
    async run(args, signal) {
      const result = await linkEntities(str(args.entityID), str(args.relatedEntityID), signal)
      return {
        payload: { kind: 'link', entityID: result.entityID, relatedEntityID: result.relatedEntityID },
        summarize: () => `Linked ${result.entityID} to ${result.relatedEntityID}.`,
      }
    },
  },

  /* ------------------------------------------------------------------ */
  /* Casework — the tenant's own workspace: private, reversible.         */
  /* ------------------------------------------------------------------ */

  list_cases: {
    schema: {
      name: 'list_cases',
      description:
        "The tenant's investigation cases, with the entities pinned to each. Use it to find the case an indicator belongs in before pinning, to answer whether something is already being tracked, or to pick up where an investigation left off.",
      parameters: {
        type: 'object',
        properties: {
          status: { type: 'string', enum: ['open', 'in_progress', 'closed'], description: 'Filter by status. Omit for all cases.' },
        },
      },
    },
    async run(args, signal) {
      const status = parseCaseStatus(args.status)
      const cases = await listCases(status, signal)
      return {
        payload: { kind: 'cases', cases },
        summarize: () => {
          if (!cases.length) return 'There are no cases.'
          return `${cases.length} case(s):\n${cases.map(describeCase).join('\n')}`
        },
      }
    },
  },

  create_case: {
    schema: {
      name: 'create_case',
      description:
        'Open a NEW investigation case. Only on explicit request. Cases count against a standing tenant quota, so prefer pinning to an existing case — check list_cases first.',
      parameters: {
        type: 'object',
        properties: {
          title: { type: 'string', description: 'Short name, e.g. "Emotet resurgence, Sept".' },
          summary: { type: 'string', description: 'What this case is about.' },
          severity: { type: 'integer', description: 'Severity 0-4: 0 informational, 1 low, 2 medium, 3 high, 4 critical.' },
          status: { type: 'string', enum: ['open', 'in_progress', 'closed'], description: 'Defaults to open.' },
        },
        required: ['title'],
      },
    },
    async run(args, signal) {
      const severity = parseSeverity(args.severity)
      const created = await createCase(
        {
          title: str(args.title),
          summary: str(args.summary) || undefined,
          severity,
          status: parseCaseStatus(args.status) ?? 'open',
        },
        signal,
      )
      return {
        payload: { kind: 'casework', action: 'case-created', label: created.title, ref: created.id },
        summarize: () => `Created case "${created.title}" (id ${created.id}). Use this id with pin_to_case and add_case_note.`,
      }
    },
  },

  pin_to_case: {
    schema: {
      name: 'pin_to_case',
      description:
        "Attach an entity already in the corpus to one of the tenant's cases, so it is kept with that investigation. Needs a case id from list_cases or create_case, and an entity id from lookup_indicator, search_entities or get_relations.",
      parameters: {
        type: 'object',
        properties: {
          caseId: { type: 'string', description: 'Target case id.' },
          entityId: { type: 'string', description: 'Entity id, e.g. "ip-4f24afca…".' },
          entityType: { type: 'string', description: 'Entity type, so the case reads without a lookup.' },
          entityValue: { type: 'string', description: 'The indicator itself, for the same reason.' },
          note: { type: 'string', description: 'Why this entity matters to the case.' },
        },
        required: ['caseId', 'entityId'],
      },
    },
    async run(args, signal) {
      const pinned = await pinEntityToCase(
        str(args.caseId),
        {
          entityId: str(args.entityId),
          entityType: str(args.entityType) || undefined,
          entityValue: str(args.entityValue) || undefined,
          note: str(args.note) || undefined,
        },
        signal,
      )
      return {
        payload: { kind: 'casework', action: 'entity-pinned', label: pinned.entity_value || pinned.entity_id, ref: pinned.case_id },
        summarize: () => `Pinned ${pinned.entity_value || pinned.entity_id} to case ${pinned.case_id}.`,
      }
    },
  },

  add_case_note: {
    schema: {
      name: 'add_case_note',
      description: "Append a note to one of the tenant's cases — findings, reasoning, what to do next. This is how an investigation gets written up. Only on explicit request.",
      parameters: {
        type: 'object',
        properties: {
          caseId: { type: 'string', description: 'Target case id.' },
          body: { type: 'string', description: 'The note. Plain text.' },
        },
        required: ['caseId', 'body'],
      },
    },
    async run(args, signal) {
      const note = await addCaseNote(str(args.caseId), str(args.body), signal)
      return {
        payload: { kind: 'casework', action: 'note-added', label: note.id, ref: note.case_id },
        summarize: () => `Added a note to case ${note.case_id}.`,
      }
    },
  },

  list_watchlists: {
    schema: {
      name: 'list_watchlists',
      description:
        "The tenant's watchlists and the rules in each. A watchlist fires an alert when a reported entity matches a rule. Use it to find which list a new rule belongs in, or to answer what is currently being watched.",
      parameters: { type: 'object', properties: {} },
    },
    async run(_args, signal) {
      const watchlists = await listWatchlists(signal)
      return {
        payload: { kind: 'watchlists', watchlists },
        summarize: () => {
          if (!watchlists.length) return 'There are no watchlists.'
          return `${watchlists.length} watchlist(s):\n${watchlists.map(describeWatchlist).join('\n')}`
        },
      }
    },
  },

  create_watchlist: {
    schema: {
      name: 'create_watchlist',
      description:
        'Create a NEW watchlist. Only on explicit request. Watchlists count against a standing tenant quota, so check list_watchlists and add to an existing one where it fits.',
      parameters: {
        type: 'object',
        properties: {
          name: { type: 'string', description: 'Short name, e.g. "Ransomware infrastructure".' },
          description: { type: 'string', description: 'What this list is for.' },
        },
        required: ['name'],
      },
    },
    async run(args, signal) {
      const created = await createWatchlist({ name: str(args.name), description: str(args.description) || undefined }, signal)
      return {
        payload: { kind: 'casework', action: 'watchlist-created', label: created.name, ref: created.id },
        summarize: () => `Created watchlist "${created.name}" (id ${created.id}). Use this id with add_watchlist_item.`,
      }
    },
  },

  add_watchlist_item: {
    schema: {
      name: 'add_watchlist_item',
      description:
        'Add one matching rule to a watchlist, so future reported entities matching it raise an alert. Needs a watchlist id from list_watchlists or create_watchlist.',
      parameters: {
        type: 'object',
        properties: {
          watchlistId: { type: 'string', description: 'Target watchlist id.' },
          kind: {
            type: 'string',
            enum: ['entity', 'value', 'tag', 'type'],
            description: 'What to match on. "entity" is one exact entity id; "value" an exact indicator string; "tag" any entity carrying that tag; "type" every entity of a type, which is broad and will be noisy.',
          },
          value: { type: 'string', description: 'The id, indicator, tag or type name to match, according to kind.' },
          maxReputation: {
            type: 'integer',
            description: 'Only alert on entities at or below this reputation, on the -3..+3 scale. Omit to alert on any match.',
          },
        },
        required: ['watchlistId', 'kind', 'value'],
      },
    },
    async run(args, signal) {
      const kind = parseMatchKind(args.kind)
      if (!kind) {
        throw new Error(`kind must be one of ${MATCH_KINDS.join(', ')}`)
      }
      const item = await addWatchlistItem(
        str(args.watchlistId),
        {
          kind,
          value: str(args.value),
          maxReputation: typeof args.maxReputation === 'number' ? Math.trunc(args.maxReputation) : undefined,
        },
        signal,
      )
      return {
        payload: { kind: 'casework', action: 'watchlist-item-added', label: `${item.kind}:${item.value}`, ref: item.watchlist_id },
        summarize: () => `Added rule ${item.kind}:${item.value} to watchlist ${item.watchlist_id}.`,
      }
    },
  },

  list_alerts: {
    schema: {
      name: 'list_alerts',
      description:
        "Alerts raised for the tenant when a reported entity matched a watchlist rule — what matched, which rule caught it, and the entity's reputation at the time. Use it to answer what has fired recently.",
      parameters: {
        type: 'object',
        properties: {
          limit: { type: 'integer', description: 'Max alerts, 1-100. Default 20.' },
        },
      },
    },
    async run(args, signal) {
      const page = await listAlerts({ limit: Math.max(1, Math.min(100, int(args.limit, 20))) }, signal)
      return {
        payload: { kind: 'alerts', alerts: page.items, unread: page.unread },
        summarize: () => {
          if (!page.items.length) return 'No alerts.'
          const rows = page.items.map(
            (a) =>
              `${a.created_at} ${a.entity_type}:${a.entity_value} reputation=${a.reputation ?? 'n/a'} matched ${a.matched_kind}:${a.matched_on} (watchlist ${a.watchlist_id})`,
          )
          return `${page.items.length} alert(s):\n${rows.join('\n')}`
        },
      }
    },
  },
}

const CASE_STATUSES: readonly CaseStatus[] = ['open', 'in_progress', 'closed']
const MATCH_KINDS = ['entity', 'value', 'tag', 'type'] as const

function parseCaseStatus(v: unknown): CaseStatus | undefined {
  return CASE_STATUSES.find((s) => s === v)
}

/** 0 informational to 4 critical. Anything else is dropped, not clamped: a
 *  model that emitted 7 did not mean 4, and guessing its intent is worse than
 *  letting the case carry no severity. */
function parseSeverity(v: unknown): number | undefined {
  if (typeof v !== 'number' || !Number.isInteger(v)) return undefined
  return v >= 0 && v <= 4 ? v : undefined
}

function parseMatchKind(v: unknown): 'entity' | 'value' | 'tag' | 'type' | undefined {
  return MATCH_KINDS.find((k) => k === v)
}

function describeCase(c: Case): string {
  return `${c.id} "${c.title}" status=${c.status} severity=${c.severity} pinned=${c.entities?.length ?? 0}`
}

function describeWatchlist(w: Watchlist): string {
  const rules = (w.items ?? []).map((i) => `${i.kind}:${i.value}`).join(', ')
  return `${w.id} "${w.name}"${w.enabled ? '' : ' (disabled)'} rules=[${rules}]`
}

function getToolSchemas() {
  return Object.values(TOOLS).map((tool) => ({
    type: 'function' as const,
    function: tool.schema,
  }))
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
      id: c.id,
      type: 'function' as const,
      function: {
        name: c.function?.name ?? '',
        arguments: typeof c.function?.arguments === 'string' ? c.function.arguments : '{}',
      },
    }))
    return { ...m, tool_calls: calls }
  })
}

/**
 * A non-streaming completion. Used by the summarizer — compaction must not
 * interleave with the main stream.
 */
async function plainCompletion(
  model: string,
  messages: ChatMessage[],
  signal: AbortSignal | undefined,
): Promise<string> {
  const res = await fetch(`${BASE}/ai/chat`, {
    method: 'POST',
    headers: await authHeaders(),
    body: JSON.stringify({ model, messages, stream: false }),
    signal,
  })
  if (!res.ok) {
    let msg = `Request failed (${res.status})`
    try {
      const j = JSON.parse(await res.text())
      const err = j?.error
      msg = (typeof err === 'string' ? err : err?.message) ?? msg
    } catch {
      // keep status message
    }
    throw new Error(msg)
  }
  const j = (await res.json()) as { choices?: { message?: { content?: string } }[] }
  const content = j.choices?.[0]?.message?.content?.trim()
  if (!content) throw new Error('summarizer returned nothing')
  return `[Earlier conversation, condensed]\n${content}`
}

function createSummarizer(model: string, config: CompactionConfig, signal?: AbortSignal): Summarizer {
  const budget = Math.floor(config.maxInputTokens * SUMMARY_INPUT_RATIO)
  return async (groups) => {
    const transcript = renderForSummary(groups, budget)
    return plainCompletion(
      model,
      [
        {
          role: 'system',
          content:
            'Condense this threat-intelligence conversation for use as context in an ongoing investigation. Keep every concrete finding: indicators, entity IDs, reputations, verdicts and which tools were already run with which arguments. Drop pleasantries and repetition. Be terse and factual.',
        },
        { role: 'user', content: transcript },
      ],
      signal,
    )
  }
}

async function handleToolCall(
  call: ModelToolCall,
  signal: AbortSignal | undefined,
  emit: AgentEvents,
  convo: ChatMessage[],
): Promise<void> {
  let args: Record<string, unknown> = {}
  try {
    const parsed: unknown = JSON.parse(call.function.arguments || '{}')
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      args = parsed as Record<string, unknown>
    }
  } catch {
    // malformed tool args — call with {} rather than crashing the turn
  }
  const step: AgentStep = {
    id: nextStepId(),
    toolName: call.function.name,
    args,
    status: 'running',
  }
  emit.onStep?.({ ...step })
  const t0 = performance.now()
  const tool = TOOLS[call.function.name]
  if (!tool) {
    step.status = 'error'
    step.error = `Unknown tool "${call.function.name}"`
    step.durationMs = performance.now() - t0
    step.result = step.error
    emit.onStep?.({ ...step })
    convo.push({ role: 'tool', tool_call_id: call.id, content: step.error })
    return
  }
  try {
    const { payload, summarize: toSummary } = await tool.run(args, signal)
    step.status = 'done'
    step.payload = payload
    step.durationMs = performance.now() - t0
    step.result = toSummary()
    emit.onStep?.({ ...step })
    convo.push({ role: 'tool', tool_call_id: call.id, content: step.result })
  } catch (e) {
    step.status = 'error'
    step.error = e instanceof Error ? e.message : String(e)
    step.durationMs = performance.now() - t0
    step.result = `Tool failed: ${step.error}`
    emit.onStep?.({ ...step })
    // Feed the failure back so the model can adapt rather than stall.
    convo.push({ role: 'tool', tool_call_id: call.id, content: step.result })
  }
}

/**
 * Runs one investigation to completion.
 *
 * The loop is: stream a completion → if the model emitted tool calls, execute
 * them all concurrently, append the results as `tool` messages, and go again →
 * otherwise the text is the final answer.
 */
export async function runAgent(
  question: string,
  opts: {
    model: string
    history: ChatMessage[]
    signal?: AbortSignal
    events?: AgentEvents
    compaction?: Partial<CompactionConfig>
  },
): Promise<AgentRunResult> {
  const { model, signal, events } = opts
  const emit = events ?? {}
  const config: CompactionConfig = { ...DEFAULT_COMPACTION, ...opts.compaction }
  const summarize = createSummarizer(model, config, signal)
  const toolSchemas = getToolSchemas()

  const convo: ChatMessage[] = [...opts.history, { role: 'user', content: question }]
  let lastText = ''

  for (let round = 0; round < MAX_TOOL_ROUNDS; round++) {
    if (signal?.aborted) {
      return { answer: lastText, history: convo, stoppedReason: 'aborted' }
    }

    // Compact before the call, not after: the point is to keep THIS request
    // under the model's input limit.
    const full = [{ role: 'system' as const, content: SYSTEM_PROMPT }, ...sanitizeHistory(convo)]
    const compacted = await compactMessages(full, config, summarize)
    if (compacted.compacted) {
      emit.onCompaction?.({
        droppedGroups: compacted.droppedGroups,
        tokensBefore: compacted.tokensBefore,
        tokensAfter: compacted.tokensAfter,
      })
      // The first entry is always the system prompt (never summarized); the
      // second is the summary when compaction happened. The rest is the new
      // convo: old turns replaced by their summary, recent turns verbatim.
      const next = compacted.messages.slice(2)
      convo.length = 0
      convo.push(...next)
    }

    const requestMessages: ChatMessage[] = [
      { role: 'system', content: SYSTEM_PROMPT },
      ...sanitizeHistory(convo),
    ]
    const res = await fetch(`${BASE}/ai/chat`, {
      method: 'POST',
      headers: await authHeaders(),
      body: JSON.stringify({ model, messages: requestMessages, tools: toolSchemas, stream: true }),
      signal,
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
      throw new ProxyError(msg, res.status)
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
          if (!tc.function?.name && !tc.function?.arguments) continue
          toolCalls[tc.index] = {
            id: tc.id ?? slot?.id ?? `call_${round}_${tc.index}`,
            type: 'function',
            function: {
              name: tc.function?.name ?? slot?.function.name ?? '',
              arguments: (slot?.function.arguments ?? '') + (tc.function.arguments ?? ''),
            },
          }
        }
      }
    }

    const calls = toolCalls.filter((c): c is ModelToolCall => Boolean(c))
    if (calls.length === 0) {
      lastText = text
      convo.push({ role: 'assistant', content: text })
      return { answer: text, history: convo, stoppedReason: 'complete' }
    }

    convo.push({
      role: 'assistant',
      content: text || null,
      tool_calls: calls,
    })

    // Tool calls within a turn are independent, so run them concurrently.
    await Promise.all(
      calls.map((call) => handleToolCall(call, signal, emit, convo)),
    )
  }

  // Soft limit: return what we have rather than throwing — the analyst keeps
  // the partial answer and the thread stays continuable.
  return { answer: lastText, history: convo, stoppedReason: 'max-turns' }
}

export { detailsToGeo, estimateTokens }
