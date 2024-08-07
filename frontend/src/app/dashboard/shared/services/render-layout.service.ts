import {Injectable} from '@angular/core';
import {GridsterItem} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';


export interface IComponent {
  id: string;
  componentRef: string;
}

@Injectable({
  providedIn: 'root'
})
export class RenderLayoutService {
  private _layout: { grid: GridsterItem, visualization: VisualizationType } [] = [];
  public components: IComponent[] = [];
  private _dashboard: UtmDashboardType;

  dropId: string;

  constructor() {
  }

  public static itemChange(item, itemComponent) {
  }

  addItem(vis: VisualizationType, grid?: GridsterItem): void {
    this._layout.push({
      grid: grid ? grid : {
        cols: 10,
        id: UUID.UUID(),
        rows: 8,
        x: 0,
        y: 0
      },
      visualization: vis
    });
  }

  get layout() {
    return [...this._layout];
  }

  get dashboard() {
    return this._dashboard;
  }

  set dashboard(dashboard: UtmDashboardType) {
    this._dashboard = {...dashboard };
  }

  clearLayout() {
    this._layout = [];
    this._dashboard = null;
  }
}
