export class MetricAggregationType {
  id: number;
  aggregation: string;
  field?: string;
  customLabel?: string;
  status?: boolean;
  value?: number;
  jsonQuery?: object;
  icon?: string;
  fontSize?: string;
  color?: string;
}
