import {ChartType} from '../../../shared/chart/types/chart.type';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';


export const CHART_TYPES: ChartType[] = [
  {
    name: 'Line',
    type: ChartTypeEnum.LINE_CHART,
    description: 'Emphasize trends'
  },
  {
    name: 'Area LineBar',
    type: ChartTypeEnum.AREA_LINE_CHART,
    description: 'Emphasize the quantity beneath a line chart'
  },
  {
    name: 'Pie',
    type: ChartTypeEnum.PIE_CHART,
    description: 'Compare parts of a whole'
  },
  {
    name: 'Bar',
    type: ChartTypeEnum.BAR_CHART,
    description: 'Assign continuous variable to each axis'
  },
  {
    name: 'Bar horizontal',
    type: ChartTypeEnum.BAR_HORIZONTAL_CHART,
    description: 'Assign continuous variable to each axis'
  },
  {
    name: 'Tag cloud',
    type: ChartTypeEnum.TAG_CLOUD_CHART,
    description: 'A group of word, sized according to their importance'
  },
  {
    name: 'Table',
    type: ChartTypeEnum.TABLE_CHART,
    description: 'Display values in a aggregated table'
  },
  {
    name: 'List',
    type: ChartTypeEnum.LIST_CHART,
    description: 'Display values in a table without aggregation'
  },
  {
    name: 'Gauge',
    type: ChartTypeEnum.GAUGE_CHART,
    description: 'Gauge indicate status of the metrics. Use it to show how a metric\'s relates to reference threshold values '
  },
  {
    name: 'Goal',
    type: ChartTypeEnum.GOAL_CHART,
    description: 'A Goal chart indicates how close your are to your final goal'
  },
  {
    name: 'Metric',
    type: ChartTypeEnum.METRIC_CHART,
    description: 'Display a calculation as a single number'
  },
  {
    name: 'Region map',
    type: ChartTypeEnum.MARKER_CHART,
    description: 'Plot latitude and longitudes coordinates on a map'
  },
  {
    name: 'Heat map',
    type: ChartTypeEnum.HEATMAP_CHART,
    description: 'Shade cells whit in a matrix'
  },
  {
    name: 'Text',
    type: ChartTypeEnum.TEXT_CHART,
    description: 'Write text to improve other chart understanding'
  },
];
