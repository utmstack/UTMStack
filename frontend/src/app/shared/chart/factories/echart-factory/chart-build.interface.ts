import {ChartOption} from './chart-option';

export interface ChartBuildInterface {
  buildChart(data?: any[], options?: any, toExport?: boolean): ChartOption;
}
