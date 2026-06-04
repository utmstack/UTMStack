export class AlertLocationType {
  alertName: string;
  locationType: 'source' | 'destination';
  ip: string;
  location: [number, number];
  accuracy: number;
}
