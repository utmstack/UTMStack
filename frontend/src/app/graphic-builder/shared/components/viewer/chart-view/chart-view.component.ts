import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {ChartFactory} from '../../../../../shared/chart/factories/echart-factory/chart-factory';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ElasticFilterDefaultTime} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {mergeParams, sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {deleteNullValues} from '../../../../../shared/util/object-util';
import {RunVisualizationBehavior} from '../../../behavior/run-visualization.behavior';
import {RunVisualizationService} from '../../../services/run-visualization.service';
import {UtmChartClickActionService} from '../../../services/utm-chart-click-action.service';
import {rebuildVisualizationFilterTime} from '../../../util/chart-filter/chart-filter.util';
import {resolveDefaultVisualizationTime} from '../../../util/visualization/visualization-render.util';
import EChartOption = echarts.EChartOption;
// @ts-ignore
require('echarts-wordcloud');

@Component({
  selector: 'app-chart-view',
  templateUrl: './chart-view.component.html',
  styleUrls: ['./chart-view.component.scss']
})
export class ChartViewComponent implements OnInit {
  @Input() chartId: number;
  @Input() building: boolean;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Input() width: string;
  @Input() height: string;
  @Input() showTime: boolean;
  @Input() timeByDefault: any;
  @Input() exportFormat: boolean;
  @Output() runned = new EventEmitter<string>();
  loadingOption: boolean;
  data: any[] = [];
  chartFactory = new ChartFactory();
  echartOption: EChartOption;
  chartTypeEnum = ChartTypeEnum;
  operatorEnum = ElasticOperatorsEnum;
  runWithError: boolean;
  defaultTime: ElasticFilterDefaultTime;
  echartsIntance: any;

  constructor(private runVisualizationService: RunVisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private toastService: UtmToastService,
              private dashboardBehavior: DashboardBehavior,
              private utmChartClickActionService: UtmChartClickActionService) {
  }

  ngOnInit() {
    this.runVisualizationBehavior.$run.subscribe(id => {
      if (id && this.chartId === id) {
        this.runVisualization();
        this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
      }
    });
    this.dashboardBehavior.$filterDashboard.subscribe(dashboardFilter => {
      if (dashboardFilter && dashboardFilter.indexPattern === this.visualization.pattern.pattern) {
        mergeParams(dashboardFilter.filter, this.visualization.filterType).then(newFilters => {
          this.visualization.filterType = sanitizeFilters(newFilters);
          this.runVisualization();
        });
      }
    });
    this.runVisualization();
    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);

    window.addEventListener('resize', (event) => {
      this.resizeChart();
    });
    window.addEventListener('beforeprint', (event) => {
      this.resizeChart();
    });
    window.addEventListener('afterprint', (event) => {
      this.resizeChart();
    });
    window.addEventListener('print', (event) => {
      this.resizeChart();
    });
  }

    chartEvent($event: any) {
      if (typeof this.visualization.chartAction === 'string') {
        this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
      }
      if (!this.building) {
        if (this.forceSingle()) {
          $event.data.name = $event.seriesName;
        }
        this.utmChartClickActionService.onClickNavigate(this.visualization, $event, this.forceSingle());
      }
    }

  /**
   * This method return if click navigation behavior will treated as single one
   */
  forceSingle(): boolean {
    return (this.visualization.chartType === ChartTypeEnum.BAR_CHART ||
      this.visualization.chartType === ChartTypeEnum.BAR_HORIZONTAL_CHART) &&
      this.data[0].series.length === 1 &&
      (this.visualization.aggregationType.bucket !== null && this.visualization.aggregationType.bucket.subBucket !== null);
  }

  runVisualization() {
    this.loadingOption = true;
    this.runVisualizationService.run(this.visualization).subscribe(data => {
      this.loadingOption = false;
      this.runned.emit('runned');
      this.data = data;
      this.runWithError = false;
      this.onChartChange();
    }, error => {
      this.loadingOption = false;
      this.runWithError = true;
      this.echartOption = null;
      this.runned.emit('runned');
      this.toastService.showError('Error',
        'Error occurred while running visualization' + this.visualization.name);
    });
  }

  /**
   * Build echart object
   */
  onChartChange() {
    if (typeof this.visualization.chartConfig === 'string') {
      this.visualization.chartConfig = JSON.parse(this.visualization.chartConfig);
    }
    this.echartOption = deleteNullValues(this.chartFactory.createChart(
      this.chart,
      this.data,
      this.visualization,
      this.exportFormat
      )
    );
  }

  onTimeFilterChange($event: TimeFilterType) {

    rebuildVisualizationFilterTime($event, this.visualization.filterType).then(filters => {
      this.visualization.filterType = filters;
      this.runVisualization();
    });
  }

  onChartInit(ec) {
    this.echartsIntance = ec;
  }

  resizeChart() {
    if (this.echartsIntance) {
      this.echartsIntance.resize();
    }
  }
}
