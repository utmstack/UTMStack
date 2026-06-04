import {Component, Input, OnChanges, OnInit, SimpleChanges, ViewChild} from '@angular/core';
import {GridsterComponent, GridsterConfig} from 'angular-gridster2';
import {UtmDashboardVisualizationType} from '../../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';

@Component({
  selector: 'app-utm-dashboard-grid',
  templateUrl: './utm-dashboard-grid.component.html',
  styleUrls: ['./utm-dashboard-grid.component.scss']
})
export class UtmDashboardGridComponent implements OnInit, OnChanges {
  @Input() options: GridsterConfig;
  @Input() visualization: UtmDashboardVisualizationType[];
  @Input() loading = true;
  @Input() UUID: string;
  @Input() mode: 'edit' | 'view';
  activeTimeGridster: number;
  windowHeight = window.innerHeight + 50;
  calculatingHeight = true;
  @ViewChild('gridsterComponent') gridsterComponent: GridsterComponent;

  constructor() {
  }

  ngOnInit() {
    // setTimeout(() => {
    //   this.windowHeight = (this.gridsterComponent.curHeight + this.gridsterComponent.curRowHeight);
    // }, 3000);
    // this.calcGridHeight().then(height => {
    //   console.log(this.gridsterComponent);
    //   // this.windowHeight = height;
    //   this.calculatingHeight = false;
    // });

  }

  ngOnChanges(changes: SimpleChanges) {
    this.calculatingHeight = true;
    this.calcGridHeight().then(height => {
      this.windowHeight = height + 15;
      this.calculatingHeight = false;
    });
  }

  calcGridHeight(): Promise<number> {
    return new Promise<number>(resolve => {
      let sum = 0;
      let prevY = -1;
      let render: UtmDashboardVisualizationType[] = this.visualization.slice();
      render = render.sort((a, b) => {
        const gridA: GridInfo = JSON.parse(a.gridInfo);
        const gridB: GridInfo = JSON.parse(b.gridInfo);
        return gridA.y < gridB.y ? 1 : -1;
      });
      for (const vis of render) {
        const grid: GridInfo = JSON.parse(vis.gridInfo);
        if (grid.y !== prevY) {
          sum += vis.height;
          prevY = grid.y;
        }
      }
      resolve(sum + 30);
    });
  }

  calc(sum: number): number {
    const height = window.innerHeight;
    sum = (height + sum) * 0.75;
    return sum;
  }

  parseGrid(item: UtmDashboardVisualizationType) {
    return JSON.parse(item.gridInfo);
  }

  setHigZindex(item: UtmDashboardVisualizationType) {
    this.activeTimeGridster = item.id;
  }

  deleteVisualization(item: UtmDashboardVisualizationType) {

  }
}

export class GridInfo {
  cols: number;
  id: string;
  rows: number;
  x: number;
  y: number;
}
