import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-render-visualization-print',
  templateUrl: './render-visualization-print.component.html',
  styleUrls: ['./render-visualization-print.component.scss']
})
export class RenderVisualizationPrintComponent implements OnInit {
  @Input() visualizationRender: UtmDashboardVisualizationType[];
  @Input() loadingVisualizations = true;
  @Output() visualizationLoaded = new EventEmitter<boolean>();
  runList = 0;
  chartTypeEnum = ChartTypeEnum;

  constructor() {
  }

  ngOnInit() {
  }

  setGridColumn(vis: UtmDashboardVisualizationType): string {
    // return 'col-lg-12 col-md-12 col-sm-12 col-xs-12';
    if (vis.visualization.chartType === ChartTypeEnum.TEXT_CHART) {
      return 'col-print-6 order-0';
    } else if (vis.visualization.chartType === ChartTypeEnum.PIE_CHART ||
      vis.visualization.chartType === ChartTypeEnum.METRIC_CHART ||
      vis.visualization.chartType === ChartTypeEnum.GAUGE_CHART ||
      vis.visualization.chartType === ChartTypeEnum.GOAL_CHART) {
      return 'col-print-6 order-1';
    } else {
      return 'col-print-12 order-2';
    }
  }

  onRun($event: string) {
    this.runList += 1;
    if (this.runList === this.visualizationRender.length) {
      console.log('All the visualizations data has loaded, waiting for rendering');
      setTimeout(() => this.visualizationLoaded.emit(true), 5000);
      console.log('All the visualizations now has rendered');
    }
  }
}
