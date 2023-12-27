import {extractMetricLabel} from '../../../../../graphic-builder/chart-builder/chart-property-builder/shared/functions/visualization-util';
import {ChartTypeEnum} from '../../../../enums/chart-type.enum';
import {HeatMapPropertiesType} from '../../../types/charts/heatmap/heat-map-properties.type';
import {HeatMapResponseType} from '../../../types/response/heat-map-response.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartBuildInterface} from '../chart-build.interface';
import {ChartOption} from '../chart-option';

export class Heatmap implements ChartBuildInterface {
  chartEnumType = ChartTypeEnum;

  constructor() {
  }

  buildChart(data?: HeatMapResponseType[], visualization?: VisualizationType): ChartOption {
    if (data) {
      const heatMapOptions: HeatMapPropertiesType = visualization.chartConfig;
      // reset visualmap max and min to recalculate
      heatMapOptions.visualMap.max = undefined;
      heatMapOptions.visualMap.min = undefined;
      heatMapOptions.series = [{
        name: extractMetricLabel(0, visualization),
        type: 'heatmap',
        data: data[0].data,
      }];
      heatMapOptions.xAxis.data = data[0].xAxis;
      heatMapOptions.yAxis.data = data[0].yAxis;
      const values = data[0].data.map((row) => {
        return row[2];
      });
      // if (!heatMapOptions.visualMap.max) {
      heatMapOptions.visualMap.max = Math.max(...values);
      // }
      // if (!heatMapOptions.visualMap.min) {
      heatMapOptions.visualMap.min = values.length > 1 ? Math.min(...values) : 0;
      // }
      return heatMapOptions;
    } else {
      return null;
    }
  }

}
