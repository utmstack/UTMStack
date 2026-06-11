import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as IncidentsHttpError }

/* ─── Types (mirror backend modules/incidents/dto) ─────────────────────── */

export type IncidentStatus = 'OPEN' | 'IN_REVIEW' | 'COMPLETED' | 'MERGED'

/** Alert reference linked into an incident (matches backend AlertLinkItem). */
export interface AlertLinkItem {
  alertId: string
  alertName: string
  alertSeverity: number
  alertStatus?: number
}

/** GET /incidents item (IncidentResponse). */
export interface Incident {
  id: number
  incidentName: string
  incidentDescription?: string
  incidentStatus: IncidentStatus
  incidentSeverity?: number
  incidentAssignedTo?: string
  incidentSolution?: string
  incidentCreatedDate: string
  alertCount: number
}

export interface IncidentAlert {
  id: number
  incidentId: number
  alertId: string
  alertName: string
  alertSeverity: number
  alertStatus?: number
}

export interface IncidentNote {
  id: number
  incidentId: number
  noteText: string
  noteSendDate: string
  noteSendBy?: string
}

export interface IncidentHistory {
  id: number
  incidentId: number
  action: string
  actionType: string
  actionDetail?: string
  actionCreatedDate: string
  actionCreatedBy?: string
}

export interface UserAssigned {
  id: number
  login: string
}

export interface CreateIncidentInput {
  incidentName: string
  incidentDescription?: string
  incidentAssignedTo?: string
  incidentObservation?: string
  alertList: AlertLinkItem[]
}

export interface ChangeStatusInput {
  id: number
  incidentName: string
  incidentDescription?: string
  incidentStatus: IncidentStatus
  incidentSeverity?: number
  incidentSolution?: string
  incidentCreatedDate?: string
}

export interface IncidentListQuery {
  incidentName?: string
  incidentStatus?: IncidentStatus
  incidentAssignedTo?: string
  createdDateStart?: string
  createdDateEnd?: string
  page?: number // 1-based
  size?: number
  sort?: string
}

function listQuery(q: IncidentListQuery): string {
  const p = new URLSearchParams()
  if (q.incidentName) p.set('incidentName', q.incidentName)
  if (q.incidentStatus) p.set('incidentStatus', q.incidentStatus)
  if (q.incidentAssignedTo) p.set('incidentAssignedTo', q.incidentAssignedTo)
  if (q.createdDateStart) p.set('createdDateStart', q.createdDateStart)
  if (q.createdDateEnd) p.set('createdDateEnd', q.createdDateEnd)
  p.set('page', String(q.page ?? 1))
  p.set('size', String(q.size ?? 20))
  if (q.sort) p.set('sort', q.sort)
  return p.toString()
}

export const incidentsHttpService = {
  // ── Incidents ──
  list: (q: IncidentListQuery = {}) => api.getPaged<Incident[]>(`/incidents?${listQuery({ size: 500, ...q })}`),
  getById: (id: number) => api.get<Incident>(`/incidents/${id}`),
  create: (input: CreateIncidentInput) => api.post<Incident>('/incidents', input),
  addAlerts: (incidentId: number, alertList: AlertLinkItem[]) =>
    api.post<Incident>('/incidents/add-alerts', { incidentId, alertList }),
  changeStatus: (input: ChangeStatusInput) => api.put<Incident>('/incidents/change-status', input),
  // Assign / reassign (assignedTo = login) or unassign (assignedTo = null).
  assign: (incidentId: number, assignedTo: string | null) =>
    api.put<Incident>('/incidents/assign', { incidentId, assignedTo }),
  usersAssigned: () => api.get<UserAssigned[]>('/incidents/users-assigned'),

  // ── Linked alerts ──
  alerts: (incidentId: number) =>
    api.getPaged<IncidentAlert[]>(`/incident-alerts?incidentId=${incidentId}&page=1&size=500`),
  updateAlertStatus: (incidentId: number, alertIds: string[], status: number) =>
    api.post<void>('/incident-alerts/update-status', { incidentId, alertIds, status }),
  removeAlert: (id: number) => api.delete<{ message: string }>(`/incident-alerts/${id}`),

  // ── Notes ──
  notes: (incidentId: number) =>
    api.getPaged<IncidentNote[]>(`/incident-notes?incidentId=${incidentId}&page=1&size=500`),
  addNote: (incidentId: number, noteText: string) =>
    api.post<IncidentNote>('/incident-notes', { incidentId, noteText }),

  // ── History ──
  history: (incidentId: number) =>
    api.getPaged<IncidentHistory[]>(`/incident-histories?incidentId=${incidentId}&page=1&size=500`),
}
