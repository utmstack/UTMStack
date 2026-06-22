export type AlertStatusSerie = 'Open' | 'In review' | 'Completed';

export interface CountAlertsByStatusBucket {
  serie: AlertStatusSerie;
  value: number;
}

export interface CountAlertsByStatusEntry {
  instanceId: number;
  instanceName: string;
  data: CountAlertsByStatusBucket[];
}
