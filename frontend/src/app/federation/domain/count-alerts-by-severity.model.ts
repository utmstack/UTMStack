export type AlertSeveritySerie = 'Low' | 'Medium' | 'High';

export interface CountAlertsBySeverityBucket {
  serie: AlertSeveritySerie;
  value: number;
}

export interface CountAlertsBySeverityEntry {
  instanceId: number;
  instanceName: string;
  data: CountAlertsBySeverityBucket[];
}
