import { ApiError, createApiClient } from '@/shared/lib/api-client'
import type {
  AlertLinkItem,
  ChangeStatusInput,
  CreateIncidentInput,
  Incident,
  IncidentAlert,
  IncidentHistory,
  IncidentListQuery,
  IncidentNote,
} from '../types/incident.types'

const api = createApiClient()

export { ApiError as IncidentsHttpError }

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
  getById: (id: string) => api.get<Incident>(`/incidents/${id}`),
  create: (input: CreateIncidentInput) => api.post<Incident>('/incidents', input),
  addAlerts: (incidentId: string, alertList: AlertLinkItem[]) =>
    api.post<Incident>('/incidents/add-alerts', { incidentId, alertList }),
  changeStatus: (input: ChangeStatusInput) => api.put<Incident>('/incidents/change-status', input),
  // Assign / reassign, or unassign with an empty string. Free text: the picker
  // suggests platform users, but an incident can be handed to anyone.
  assign: (incidentId: string, assignedTo: string) =>
    api.put<Incident>('/incidents/assign', { incidentId, assignedTo }),
  // The assignees currently in use, for the filter to offer.
  assignees: () => api.get<string[]>('/incidents/assignees'),

  // ── Linked alerts ──
  alerts: (incidentId: string) =>
    api.getPaged<IncidentAlert[]>(`/incident-alerts?incidentId=${incidentId}&page=1&size=500`),
  updateAlertStatus: (incidentId: string, alertIds: string[], status: string) =>
    api.post<void>('/incident-alerts/update-status', { incidentId, alertIds, status }),
  removeAlert: (id: string) => api.delete<{ message: string }>(`/incident-alerts/${id}`),

  // ── Notes ──
  notes: (incidentId: string) =>
    api.getPaged<IncidentNote[]>(`/incident-notes?incidentId=${incidentId}&page=1&size=500`),
  addNote: (incidentId: string, noteText: string) =>
    api.post<IncidentNote>('/incident-notes', { incidentId, noteText }),

  // ── History ──
  history: (incidentId: string) =>
    api.getPaged<IncidentHistory[]>(`/incident-histories?incidentId=${incidentId}&page=1&size=500`),
}
