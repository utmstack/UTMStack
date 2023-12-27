import {GridsterItem} from 'angular-gridster2';
import {UtmDashboardVisualizationType} from '../../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {DashboardFilterType} from '../../../../shared/types/filter/dashboard-filter.type';

export class DashboardDraftType {
  visualization: UtmDashboardVisualizationType[];
  timeEnable: number[];
  filters: DashboardFilterType[];
  layout: { grid: GridsterItem, visualization: VisualizationType } [];
}
