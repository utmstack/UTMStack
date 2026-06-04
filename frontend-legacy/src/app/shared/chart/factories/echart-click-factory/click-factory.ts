import {MULTIPLE_BUCKETS_CHART} from '../../../constants/visualization-bucket-metric.constant';
import {EchartClickAction} from '../../types/action/echart-click-action';
import {VisualizationType} from '../../types/visualization.type';
import {MultipleBucketChartClick} from './bucket-type/multiple-bucket-chart-click';
import {SingleBucketChartClick} from './bucket-type/single-bucket-chart-click';

export class ClickFactory {
  public createParams(visualization: VisualizationType, chartValue: EchartClickAction, forceSingle?: boolean): object {
    if (!MULTIPLE_BUCKETS_CHART.includes(visualization.chartType) || forceSingle) {
      return new SingleBucketChartClick().buildClickParams(visualization, chartValue);
    } else {
      return new MultipleBucketChartClick().buildClickParams(visualization, chartValue);
    }
  }
}
