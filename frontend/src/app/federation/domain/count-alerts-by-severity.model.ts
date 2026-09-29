export type AlertSeveritySerie = 'Low' | 'Medium' | 'High';

export interface CountAlertsBySeverityBucket {
  name: AlertSeveritySerie;
  value: number;
}

export interface CountAlertsBySeverityEntry {
  instanceId: number;
  instanceName: string;
  data: {value:CountAlertsBySeverityBucket[]};
}
