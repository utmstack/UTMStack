import {IncidentStatusEnum} from '../../enums/incident/incident-status.enum';

export class UtmIncidentType {
  id?: number;
  incidentName: string;
  incidentDescription: string;
  incidentStatus: IncidentStatusEnum;
  incidentSeverity: number;
  incidentAssignedTo: string;
  incidentCreatedDate?: Date;
  incidentSolution: string;
}
