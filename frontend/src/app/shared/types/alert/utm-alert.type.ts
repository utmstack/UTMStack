export class UtmAlertType {
  timestamp: Date;
  id: string;
  status: AlertStatusEnum;
  statusLabel: AlertStatusLabelEnum;
  statusObservation: string;
  isIncident: boolean;
  incidentDetail: IncidentDetail;
  name: string;
  category: string;
  severity: AlertSeverityEnum;
  severityLabel: AlertSeverityLabelEnum;
  protocol: string;
  description: string;
  solution: string;
  tactic: string;
  reference: string[];
  dataType: string;
  dataSource: string;
  source: AlertHost;
  destination: AlertHost;
  logs: string[];
  tags: string[];
  notes: string;
  tagRulesApplied: number[];
}

export enum AlertStatusLabelEnum {
  AUTOMATIC_REVIEW_LABEL = 'Automatic review',
  OPEN_LABEL = 'Open',
  IN_REVIEW_LABEL = 'In review',
  IGNORED_LABEL = 'Ignored',
  COMPLETED_LABEL = 'Completed'
}

export enum AlertStatusEnum {
  AUTOMATIC_REVIEW = 1,
  OPEN = 2,
  IN_REVIEW = 3,
  IGNORED = 4,
  COMPLETED = 5,

}

export class IncidentDetail {
  createdBy: string;
  creationDate: Date;
  incidentId: string;
  incidentName: string;
  observation: string;
  source: string;
}

export enum AlertSeverityEnum {
  LOW = 1,
  MEDIUM = 2,
  HIGH = 3
}

export enum AlertSeverityLabelEnum {
  LOW_LABEL = 'LOW',
  MEDIUM_LABEL = 'MEDIUM',
  HIGH_LABEL = 'HIGH'
}

export class AlertHost {
  user: string;
  host: string;
  String;
  ip: string;
  port: number;
  country: string;
  countryCode: string;
  city: string;
  coordinates: number[];
  accuracyRadius: number;
  asn: number;
  aso: string;
  isSatelliteProvider: boolean;
  isAnonymousProxy: boolean;
}
