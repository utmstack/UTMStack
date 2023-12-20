import {ChartTypeEnum} from '../../../../enums/chart-type.enum';
import {ChartValueSeparator} from '../../../../enums/chart-value-separator';
import {IndexPatternSystemEnumName} from '../../../../enums/index-pattern-system.enum';
import {EchartClickAction} from '../../../types/action/echart-click-action';
import {MetricBucketsType} from '../../../types/metric/metric-buckets.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartClickInterface} from '../chart-click.interface';

export class MultipleBucketChartClick implements ChartClickInterface {
  buildClickParams(visualization: VisualizationType, chartValue: EchartClickAction): object {
    const queryParams: any = {};
    visualization.filterType.forEach(value => {
      queryParams[value.field] = value.operator + '->' + value.value;
    });
    if (visualization.chartType !== ChartTypeEnum.TABLE_CHART) {
      const axisField = visualization.aggregationType.bucket.field;
      queryParams[axisField] = chartValue.name;
    }
    const buckets = chartValue.seriesName.split(ChartValueSeparator.BUCKET_SEPARATOR);
    for (const bucket of buckets) {
      const bucketSplit = bucket.split(ChartValueSeparator.VALUE_SEPARATOR);
      queryParams[bucketSplit[0]] = bucketSplit[1];
    }

    if (visualization.pattern.pattern === IndexPatternSystemEnumName.ALERT) {
      queryParams.alertType = 'ALERT';
    } else {
      queryParams.patternId = visualization.pattern.id;
      queryParams.indexPattern = visualization.pattern.pattern;
    }
    return Object.keys(queryParams).length === 0 ? null : queryParams;
  }

  extractVisualizationBuckets(metricBucketsType: MetricBucketsType): MetricBucketsType[] {
    const bucketsTypes: MetricBucketsType[] = [];
    while (metricBucketsType.subBucket) {
      bucketsTypes.push(metricBucketsType.subBucket);
    }
    return bucketsTypes;
  }
}
