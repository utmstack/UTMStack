import {Injectable} from '@angular/core';
import {GridsterItem} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';

export interface IComponent {
  id: string;
  componentRef: string;
}

@Injectable({
  providedIn: 'root'
})
export class LayoutService {


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

  setItem(vis: VisualizationType) {
    const indexVis = this.layout.findIndex(value => value.visualization.id === vis.id);
    if (indexVis !== -1) {
      this.layout[indexVis].visualization = vis;
    }
  }

  deleteItem(gridsterItem: GridsterItem): void {
    const index = this.layout.findIndex(d => d.grid.id === gridsterItem.id);
    this.layout.splice(index, 1);
    // @ts-ignore
    const indexComp = this.components.findIndex(c => c.id === gridsterItem.id);
    this.components.splice(indexComp, 1);
  }

  setDropId(dropId: string): void {
    this.dropId = dropId;
  }

  dropItem(dragId: string): void {
    const {components} = this;
    const comp: IComponent = components.find(c => c.id === this.dropId);
    const updateIdx: number = comp ? components.indexOf(comp) : components.length;
    const componentItem: IComponent = {
      id: this.dropId,
      componentRef: dragId
    };
    this.components = Object.assign([], this.components, {[updateIdx]: componentItem});
  }

  getComponentRef(id: string): string {
    const comp = this.components.find(c => c.id === id);
    return comp ? comp.componentRef : null;
  }

}
