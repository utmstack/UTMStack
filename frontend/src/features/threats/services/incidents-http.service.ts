import { ApiError, createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

export { ApiError as IncidentsHttpError }

/** Alert reference linked into an incident (matches backend AlertLinkItem). */
export interface AlertLinkItem {
  alertId: string
  alertName: string
  alertSeverity: number
  alertStatus?: number
}

/** GET /incidents list item (subset). */
export interface Incident {
  id: number
  incidentName: string
  incidentDescription?: string
  status?: number
  statusLabel?: string
  createdDate?: string
  assignedTo?: string
}

export interface CreateIncidentInput {
  incidentName: string
  incidentDescription?: string
  incidentAssignedTo?: string
  alertList: AlertLinkItem[]
}

export const incidentsHttpService = {
  // List existing incidents (for the "add to existing" picker).
  list: (search = '') => {
    const q = new URLSearchParams({ page: '1', size: '500' })
    if (search) q.set('incidentName', search)
    return api.getPaged<Incident[]>(`/incidents?${q.toString()}`)
  },
  // Create a new incident with its linked alerts. Returns the created incident (with id).
  create: (input: CreateIncidentInput) => api.post<Incident>('/incidents', input),
  // Link more alerts into an existing incident.
  addAlerts: (incidentId: number, alertList: AlertLinkItem[]) =>
    api.post<Incident>('/incidents/add-alerts', { incidentId, alertList }),
}
