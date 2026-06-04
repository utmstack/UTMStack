import {MetricAggregationType} from '../../../../../shared/chart/types/metric/metric-aggregation.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';

/**
 * Return label from metric
 * @param metricId ID of metric
 * @param visualization Visualization object
 */
export function extractMetricLabel(metricId: number | string, visualization: VisualizationType): string {
  const index = visualization.aggregationType.metrics.findIndex(v =>
    Number(v.id) === Number(metricId));
  if (visualization.aggregationType.metrics[index]) {
    return getMetricLabelByIndex(index, visualization);
  }
}


export function extractLabelFromMetric(metricId: number | string, metrics: MetricAggregationType[]): string {
  const index = metrics.findIndex(v =>
    Number(v.id) === Number(metricId));
  if (metrics[index]) {
    return getMetricLabelByIndexFromMetric(index, metrics);
  }
}

export function getMetricLabelByIndexFromMetric(index: number, metrics: MetricAggregationType[]): string {
  return metrics[index].customLabel === '' ?
    (metrics[index].aggregation +
      ' ' + (metrics[index].aggregation === 'COUNT' ?
        '(' + metrics[index].id + ')' :
        '(' + metrics[index].field + ')')) :
    metrics[index].customLabel;
}

/**
 * Return label for a bucket, if there any bucket and custom label is not empty return custom label
 * else return metric name on field
 * @param index Index of bucket
 * @param visualization Visualization object
 */
export function getBucketLabel(index: number, visualization: VisualizationType): string {
  if (visualization.aggregationType.bucket) {
    return visualization.aggregationType.bucket.customLabel === '' ?
      (visualization.aggregationType.bucket.aggregation +
        '(' + visualization.aggregationType.bucket.field + ')') :
      visualization.aggregationType.bucket.customLabel;
  } else {
    return getMetricLabelByIndex(index, visualization);
  }
}

/**
 * Return metric or bucket label based in index and visualization object depending of custom label
 * of metric or bucket
 * @param index Index of bucket or metric
 * @param visualization Visualization object
 */
export function getMetricLabelByIndex(index: number, visualization: VisualizationType): string {
  return visualization.aggregationType.metrics[index].customLabel === '' ?
    visualization.aggregationType.metrics[index].aggregation : visualization.aggregationType.metrics[index].customLabel;
  // return (visualization.aggregationType.metrics[index].aggregation +
  //   ' ' + (visualization.aggregationType.metrics[index].aggregation === 'COUNT' ?
  //     '(' + visualization.aggregationType.metrics[index].id + ')' :
  //     '(' + visualization.aggregationType.metrics[index].field + ')'));
}
