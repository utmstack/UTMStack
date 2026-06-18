/* Mirrors backend modules/incidents/dto. */

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
