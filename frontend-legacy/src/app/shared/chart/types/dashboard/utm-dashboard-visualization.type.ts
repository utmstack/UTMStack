import {VisualizationType} from '../visualization.type';
import {UtmDashboardType} from './utm-dashboard.type';

export class UtmDashboardVisualizationType {
  height?: number;
  idDashboard?: number;
  idVisualization?: number;
  left?: number;
  order?: number;
  top?: number;
  width?: number;
  visualization?: VisualizationType;
  dashboard?: UtmDashboardType;
  defaultTimeRange?: any;
  showTimeFilter?: boolean;
  id?: number;
  gridInfo?: any;
}
