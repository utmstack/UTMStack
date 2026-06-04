import {BucketAggregationEnum} from '../enums/bucket-aggregation.enum';
import {MetricAggregationEnum} from '../enums/metric-aggregation.enum';

export const AGGREGATIONS = [
  {agg: 'Average', group: 'Metric Aggregations', value: MetricAggregationEnum.AVERAGE},
  {agg: 'Count', group: 'Metric Aggregations', value: MetricAggregationEnum.COUNT},
  {agg: 'Max', group: 'Metric Aggregations', value: MetricAggregationEnum.MAX},
  {agg: 'Median', group: 'Metric Aggregations', value: MetricAggregationEnum.MEDIAN},
  {agg: 'Min', group: 'Metric Aggregations', value: MetricAggregationEnum.MIN},
  // {agg: 'Percentile Ranks', group: 'Metric Aggregations', value:  MetricAggregationEnum.PERCENTILE_RANK},
  // {agg: 'Percentiles', group: 'Metric Aggregations', value:  MetricAggregationEnum.PERCENTIL},
  {agg: 'Sum', group: 'Metric Aggregations', value: MetricAggregationEnum.SUM},
  // {agg: 'Top Hit', group: 'Metric Aggregations', value:  MetricAggregationEnum.TOP_HIT},
  {agg: 'Unique Count', group: 'Metric Aggregations', value: MetricAggregationEnum.UNIQUE_COUNT},
  // {agg: 'Average Bucket', group: 'Sibling Pipeline Aggregations', value:  MetricAggregationEnum.AVG'},
  // {agg: 'Max Bucket', group: 'Sibling Pipeline Aggregations', value: 'AVG'},
  // {agg: 'Min Bucket', group: 'Sibling Pipeline Aggregations', value: 'AVG'},
  // {agg: 'Sum Bucket', group: 'Sibling Pipeline Aggregations', value: 'AVG'},
];

export const PIE_AGGREGATIONS = [
  {agg: 'Count', group: 'Metric Aggregations', value: MetricAggregationEnum.COUNT},
  {agg: 'Sum', group: 'Metric Aggregations', value: MetricAggregationEnum.SUM},
  {agg: 'Top Hit', group: 'Metric Aggregations', value: MetricAggregationEnum.TOP_HIT},
  {agg: 'Unique Count', group: 'Metric Aggregations', value: MetricAggregationEnum.UNIQUE_COUNT}
];

export const BUCKETS_AGGREGATIONS = [
  // {agg: 'Date Range', value: BucketAggregationEnum.DATE_RANGE},
  // {agg: 'Filters', value: BucketAggregationEnum.FILTERS},
  // {agg: 'Histogram', value: BucketAggregationEnum.HISTOGRAM},
  // {agg: 'IPv4 Range', value: BucketAggregationEnum.IPV4_RANGE},
  // {agg: 'Range', value: BucketAggregationEnum.RANGE},
  // {agg: 'Significant Terms', value: BucketAggregationEnum.SIGNIFICANT_TERMS},
  {agg: 'Terms', value: BucketAggregationEnum.TERMS},
  {agg: 'Date Histogram', value: BucketAggregationEnum.DATE_HISTOGRAM},
];
