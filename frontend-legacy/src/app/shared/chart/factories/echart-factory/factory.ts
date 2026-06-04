import {VisualizationType} from '../../types/visualization.type';
import {ChartOption} from './chart-option';

export interface Factory {
  createChart(type?: string,
              data?: any[],
              visualization?: VisualizationType): ChartOption;
}
