import {ChartTypeEnum} from '../../../enums/chart-type.enum';
import {VisualizationType} from '../../types/visualization.type';
import {ChartOption} from './chart-option';
import {Gauge} from './charts/gauge';
import {Heatmap} from './charts/heatmap';
import {LineBar} from './charts/line-bar';
import {Pie} from './charts/pie';
import {ScatterMap} from './charts/scatter-map';
import {TagCloud} from './charts/tag-cloud';
import {Factory} from './factory';

export class ChartFactory implements Factory {
  createChart(type?: string, data?: any[], visualization?: VisualizationType, toExport?: boolean):
    ChartOption {
    switch (type) {
      case ChartTypeEnum.PIE_CHART:
        return new Pie().buildChart(data, visualization, toExport);
      case  ChartTypeEnum.GAUGE_CHART:
        return new Gauge().buildChart(data, visualization);
      case  ChartTypeEnum.TAG_CLOUD_CHART:
        return new TagCloud().buildChart(data, visualization);
      case  ChartTypeEnum.LINE_CHART:
        return new LineBar().buildChart(data, visualization);
      case  ChartTypeEnum.AREA_LINE_CHART:
        return new LineBar().buildChart(data, visualization);
      case  ChartTypeEnum.BAR_CHART:
        return new LineBar().buildChart(data, visualization, toExport);
      case  ChartTypeEnum.BAR_HORIZONTAL_CHART:
        return new LineBar().buildChart(data, visualization, toExport);
      case  ChartTypeEnum.MARKER_CHART:
        return new ScatterMap().buildChart(data, visualization);
      case  ChartTypeEnum.HEATMAP_CHART:
        return new Heatmap().buildChart(data, visualization);
    }
  }
}
