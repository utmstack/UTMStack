import { proxyFetch } from '../console/proxy'

/**
 * Casework — the analyst's own records: cases, watchlists, alerts, and
 * conversation history. Stored in the UTMStack tenant's own database
 * (not proxied to ThreatWinds — the reverse proxy strips user credentials,
 * so upstream per-user ownership is unexpressible). Every route here is
 * scoped to the authenticated tenant.
 */

export type CaseStatus = 'open' | 'in_progress' | 'closed'

export interface CaseEntity {
  id: string
  case_id: string
  entity_id: string
  entity_type?: string
  entity_value?: string
  note?: string
}

export interface CaseNote {
  id: string
  case_id: string
  body: string
  created_at: string
}

export interface Case {
  id: string
  title: string
  summary?: string
  status: CaseStatus
  severity: number
  notes?: CaseNote[]
  entities?: CaseEntity[]
  created_at: string
  updated_at: string
}

export interface WatchlistItem {
  id: string
  watchlist_id: string
  kind: 'entity' | 'value' | 'tag' | 'type'
  value: string
  max_reputation?: number
}

export interface Watchlist {
  id: string
  name: string
  enabled: boolean
  items?: WatchlistItem[]
  created_at: string
}

export interface Alert {
  id: string
  watchlist_id: string
  item_id: string
  entity_id: string
  entity_type: string
  entity_value: string
  reputation: number | null
  matched_kind: string
  matched_on: string
  created_at: string
}

function post<T>(path: string, body: unknown, signal?: AbortSignal): Promise<T> {
  return proxyFetch<T>(path, { method: 'POST', body: JSON.stringify(body), signal })
}

export function listCases(status?: CaseStatus, signal?: AbortSignal): Promise<Case[]> {
  const q = status ? `?status=${status}` : ''
  return proxyFetch<{ cases: Case[] }>(`/cases${q}`, { signal }).then(r => r.cases ?? [])
}

export function createCase(
  input: { title: string; summary?: string; severity?: number; status?: CaseStatus },
  signal?: AbortSignal,
): Promise<Case> {
  return post<Case>('/cases', input, signal)
}

export function pinEntityToCase(
  caseId: string,
  input: { entityId: string; entityType?: string; entityValue?: string; note?: string },
  signal?: AbortSignal,
): Promise<CaseEntity> {
  return post<CaseEntity>(`/cases/${caseId}/entities`, {
    entity_id: input.entityId,
    entity_type: input.entityType,
    entity_value: input.entityValue,
    note: input.note,
  }, signal)
}

export function addCaseNote(caseId: string, body: string, signal?: AbortSignal): Promise<CaseNote> {
  return post<CaseNote>(`/cases/${caseId}/notes`, { body }, signal)
}

export function listWatchlists(signal?: AbortSignal): Promise<Watchlist[]> {
  return proxyFetch<{ watchlists: Watchlist[] }>('/watchlists', { signal }).then(r => r.watchlists ?? [])
}

export function createWatchlist(input: { name: string; description?: string }, signal?: AbortSignal): Promise<Watchlist> {
  return post<Watchlist>('/watchlists', input, signal)
}

export function addWatchlistItem(
  watchlistId: string,
  input: { kind: 'entity' | 'value' | 'tag' | 'type'; value: string; maxReputation?: number },
  signal?: AbortSignal,
): Promise<WatchlistItem> {
  return post<WatchlistItem>(`/watchlists/${watchlistId}/items`, {
    kind: input.kind,
    value: input.value,
    max_reputation: input.maxReputation,
  }, signal)
}

export function listAlerts(opts: { limit?: number }, signal?: AbortSignal): Promise<{ items: Alert[]; unread: number }> {
  const q = opts.limit ? `?limit=${opts.limit}` : ''
  return proxyFetch<{ items: Alert[]; unread: number }>(`/alerts${q}`, { signal })
    .then(r => ({ items: r.items ?? [], unread: r.unread ?? 0 }))
}

export interface LinkEntitiesResult {
  entityID: string
  relatedEntityID: string
  message: string
}

export function linkEntities(entityID: string, relatedEntityID: string, signal?: AbortSignal): Promise<LinkEntitiesResult> {
  return post<{ message?: string } | null>('/report/association', { entityID, relatedEntityID }, signal)
    .then(ack => ({ entityID, relatedEntityID, message: ack?.message ?? 'accepted' }))
}
