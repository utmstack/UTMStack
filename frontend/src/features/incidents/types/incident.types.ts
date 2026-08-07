/* Mirrors backend modules/incidents/dto. */

// The words the backend stores, not codes. They are the same vocabulary the
// alerts use, so an incident's status and its alerts' statuses read alike.
export type IncidentStatus = 'Open' | 'In review' | 'Completed' | 'Merged'

/** low | medium | high, or absent on an incident that holds no alerts. */
export type IncidentSeverity = 'low' | 'medium' | 'high'

/** Alert reference linked into an incident (matches backend AlertLinkItem). */
export interface AlertLinkItem {
  alertId: string
  alertName: string
  alertSeverity: string
  alertStatus?: string
}

/** GET /incidents item (IncidentResponse). */
export interface Incident {
  id: string
  incidentName: string
  incidentDescription?: string
  incidentStatus: IncidentStatus
  incidentSeverity?: IncidentSeverity
  /** Free text: a platform user's email, or anyone else the incident was handed to. */
  incidentAssignedTo?: string
  incidentSolution?: string
  incidentCreatedDate: string
  alertCount: number
}

export interface IncidentAlert {
  id: string
  incidentId: string
  alertId: string
  alertName: string
  alertSeverity: IncidentSeverity
  alertStatus?: string
}

export interface IncidentNote {
  id: string
  incidentId: string
  noteText: string
  noteSendDate: string
  noteSendBy?: string
}

export interface IncidentHistory {
  id: string
  incidentId: string
  /** The action code, e.g. INCIDENT_STATUS_CHANGE. The English label and the
   *  detail line that used to sit beside it no longer have columns. */
  action: string
  actionCreatedDate: string
  actionCreatedBy?: string
}

export interface CreateIncidentInput {
  incidentName: string
  incidentDescription?: string
  incidentAssignedTo?: string
  incidentObservation?: string
  alertList: AlertLinkItem[]
}

export interface ChangeStatusInput {
  id: string
  incidentName: string
  incidentDescription?: string
  incidentStatus: IncidentStatus
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
