import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import { Observable, of, Subject} from 'rxjs';
import {catchError, filter, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {ChartFactory} from '../../../../../shared/chart/factories/echart-factory/chart-factory';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {
  ElasticFilterDefaultTime
} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {RefreshService, RefreshType} from '../../../../../shared/services/util/refresh.service';
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
  styleUrls: ['./chart-view.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartViewComponent implements OnInit, OnDestroy {

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
  data$: Observable<any[]>;
  chartFactory = new ChartFactory();
  echartOption: EChartOption;
  chartTypeEnum = ChartTypeEnum;
  operatorEnum = ElasticOperatorsEnum;
  runWithError: boolean;
  defaultTime: ElasticFilterDefaultTime;
  echartsIntance: any;
  refreshType: string;
  destroy$ = new Subject<void>();

  constructor(private runVisualizationService: RunVisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private toastService: UtmToastService,
              private dashboardBehavior: DashboardBehavior,
              private utmChartClickActionService: UtmChartClickActionService,
              private refreshService: RefreshService) {
  }

  ngOnInit() {
    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
    this.refreshType = `${this.chartId}`;
    this.data$ = this.refreshService.refresh$
      .pipe(
        takeUntil(this.destroy$),
        filter((refreshType) =>  refreshType === RefreshType.ALL ||
          refreshType === this.refreshType),
        switchMap((value, index) => this.runVisualization()));


    this.runVisualizationBehavior.$run
      .pipe(takeUntil(this.destroy$))
      .subscribe(id => {
      if (id && this.chartId === id) {
        this.refreshService.sendRefresh(this.refreshType);
        this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
      }
    });

    this.dashboardBehavior.$filterDashboard
      .pipe(takeUntil(this.destroy$))
      .subscribe(dashboardFilter => {
      if (dashboardFilter && dashboardFilter.indexPattern === this.visualization.pattern.pattern) {
        mergeParams(dashboardFilter.filter, this.visualization.filterType).then(newFilters => {
          this.visualization.filterType = sanitizeFilters(newFilters);
          this.refreshService.sendRefresh(this.refreshType);
        });
      }
    });

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

    if (!this.defaultTime) {
      this.refreshService.sendRefresh(this.refreshType);
    }
  }

  chartEvent($event: any, data: []) {
    if (typeof this.visualization.chartAction === 'string') {
      this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
    }
    if (!this.building) {
      if (this.forceSingle(data)) {
        $event.data.name = $event.seriesName;
      }
      this.utmChartClickActionService.onClickNavigate(this.visualization, $event, this.forceSingle(data));
    }
  }

  /**
   * This method return if click navigation behavior will treated as single one
   */
  forceSingle(data: any[]): boolean {
    return (this.visualization.chartType === ChartTypeEnum.BAR_CHART ||
        this.visualization.chartType === ChartTypeEnum.BAR_HORIZONTAL_CHART) &&
        data[0].series.length === 1 &&
      (this.visualization.aggregationType.bucket !== null && this.visualization.aggregationType.bucket.subBucket !== null);
  }

  runVisualization() {
    this.loadingOption = true;
    return this.runVisualizationService.run(this.visualization)
      .pipe(
        tap((data) => {
          this.loadingOption = false;
          this.runned.emit('runned');
          this.runWithError = false;
          this.onChartChange(data);
        }),
        catchError(() => {
          this.loadingOption = false;
          this.runWithError = true;
          this.echartOption = null;
          this.runned.emit('runned');
          this.toastService.showError('Error',
            'Error occurred while running visualization' + this.visualization.name);
          return of([]);
        }));
  }

  /**
   * Build echart object
   */
  onChartChange(data: any[]) {
    if (typeof this.visualization.chartConfig === 'string') {
      this.visualization.chartConfig = JSON.parse(this.visualization.chartConfig);
    }
    this.echartOption = deleteNullValues(this.chartFactory.createChart(
        this.chart,
        data,
        this.visualization,
        this.exportFormat
      )
    );
  }

  onTimeFilterChange($event: TimeFilterType) {
    rebuildVisualizationFilterTime($event, this.visualization.filterType).then(filters => {
      this.visualization.filterType = filters;
      this.refreshService.sendRefresh(this.refreshType);
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

  ngOnDestroy(): void {
    this.refreshService.sendRefresh(null);
    this.destroy$.next();
    this.destroy$.complete();
  }
}
