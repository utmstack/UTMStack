export class NewIncidentType {
  incidentName: string;
  incidentDescription: string;
  incidentAssignedTo: string;
  alertList: NewIncidentAlert[];
}

export class NewIncidentAlert {
  alertId: string;
  alertName: string;
  alertStatus: number;
  alertSeverity: number;
}
