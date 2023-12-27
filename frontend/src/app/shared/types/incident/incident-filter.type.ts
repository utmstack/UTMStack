export class IncidentFilterType {
  'incidentName.contains'?: string;
  page: number;
  size: number;
  sort?: string;
  'id.equals'?: number;
  'incidentStatus.in'?: string[];
  'incidentAssignedTo.contains'?: string;
  'incidentCreatedDate.greaterThanOrEqual'?: string;
  'incidentCreatedDate.lessThanOrEqual'?: string;
  'incidentSeverity.in'?: number[];
  'incidentAssignedTo.in'?: string[];
}
