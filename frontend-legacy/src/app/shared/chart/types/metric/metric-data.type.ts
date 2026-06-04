import {MetricAggregationType} from './metric-aggregation.type';
import {MetricBucketsType} from './metric-buckets.type';

export class MetricDataType {
  metrics?: MetricAggregationType[];
  bucket?: MetricBucketsType;
}
