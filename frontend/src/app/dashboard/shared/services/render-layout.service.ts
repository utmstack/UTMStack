import {Injectable} from '@angular/core';
import {GridsterItem} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';


export interface IComponent {
  id: string;
  componentRef: string;
}

@Injectable({
  providedIn: 'root'
})
export class RenderLayoutService {
  public layout: { grid: GridsterItem, visualization: VisualizationType } [] = [];
  public components: IComponent[] = [];

  dropId: string;

  constructor() {
  }

  public static itemChange(item, itemComponent) {
  }

  addItem(vis: VisualizationType, grid?: GridsterItem): void {
    this.layout.push({
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

}
