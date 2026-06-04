import {EchartClickAction} from '../../types/action/echart-click-action';
import {VisualizationType} from '../../types/visualization.type';

export interface ChartClickInterface {
  buildClickParams(visualization: VisualizationType, chartValue: EchartClickAction): object;
}
