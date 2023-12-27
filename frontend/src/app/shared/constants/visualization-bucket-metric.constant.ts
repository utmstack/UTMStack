import {ChartTypeEnum} from '../enums/chart-type.enum';

export const MULTIPLE_BUCKETS_CHART: ChartTypeEnum[] = [
  ChartTypeEnum.TABLE_CHART,
  ChartTypeEnum.LINE_CHART,
  ChartTypeEnum.AREA_LINE_CHART,
  ChartTypeEnum.BAR_HORIZONTAL_CHART,
  ChartTypeEnum.LIST_CHART,
  ChartTypeEnum.BAR_CHART,
  ChartTypeEnum.HEATMAP_CHART];

export const MULTIPLE_METRIC_CHART: ChartTypeEnum[] =
  [ChartTypeEnum.METRIC_CHART,
    ChartTypeEnum.GOAL_CHART,
    ChartTypeEnum.TABLE_CHART,
    ChartTypeEnum.LINE_CHART,
    ChartTypeEnum.AREA_LINE_CHART,
    ChartTypeEnum.BAR_HORIZONTAL_CHART,
    ChartTypeEnum.BAR_CHART
  ];
