import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {EchartClickAction} from '../../../../../shared/chart/types/action/echart-click-action';
import {MetricResponse} from '../../../../../shared/chart/types/metric/metric-response';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {
  ElasticFilterDefaultTime
} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {mergeParams, sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {extractMetricLabel} from '../../../../chart-builder/chart-property-builder/shared/functions/visualization-util';
import {RunVisualizationBehavior} from '../../../behavior/run-visualization.behavior';
import {RunVisualizationService} from '../../../services/run-visualization.service';
import {UtmChartClickActionService} from '../../../services/utm-chart-click-action.service';
import {rebuildVisualizationFilterTime} from '../../../util/chart-filter/chart-filter.util';
import {resolveDefaultVisualizationTime} from '../../../util/visualization/visualization-render.util';
import {Observable, of, Subject} from "rxjs";
import {RefreshService, RefreshType} from "../../../../../shared/services/util/refresh.service";
import {catchError, filter, switchMap, takeUntil, tap} from 'rxjs/operators';

@Component({
  selector: 'app-metric-view',
  templateUrl: './metric-view.component.html',
  styleUrls: ['./metric-view.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class MetricViewComponent implements OnInit, OnDestroy {
  data: MetricResponse[];
  data$: Observable<MetricResponse[]>;
  @Input() visualization: VisualizationType;
  @Input() building: boolean;
  @Output() runned = new EventEmitter<string>();
  @Input() showTime: boolean;
  @Input() height: string;
  @Input() width: string;
  @Input() timeByDefault: any;
  @Input() chartId: number;
  runningChart: boolean;
  error = false;
  kpis: KpiValues[] = [];
  METRIC_CHART = ChartTypeEnum.METRIC_CHART;
  defaultTime: ElasticFilterDefaultTime;
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
        filter((refreshType) => refreshType === RefreshType.ALL ||
          refreshType === this.refreshType),
        switchMap((value, index) => this.runVisualization()));

    this.runVisualizationBehavior.$run
      .pipe(takeUntil(this.destroy$))
      .subscribe((id) => {
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

    if (!this.defaultTime) {
      this.refreshService.sendRefresh(this.refreshType);
    }
  }

  runVisualization() {
    this.runningChart = true;
    /*this.runVisualizationService.run(this.visualization).subscribe(data => {
      this.runningChart = false;
      this.runned.emit('runned');
      this.data = data;
      this.error = false;
      this.extractKpiValues().then(kpis => {
        this.kpis = kpis;
      });
    }, error => {
      this.runningChart = false;
      this.runned.emit('runned');
      this.error = true;
      this.toastService.showError('Error',
        'Error occurred while running visualization');
    });*/

    return this.runVisualizationService.run(this.visualization)
      .pipe(
        tap((data) => {
          this.runningChart = false;
          this.runned.emit('runned');
          this.data = data;
          this.error = false;
          this.extractKpiValues().then(kpis => {
            this.kpis = kpis;
          });
        }),
        catchError(() => {
          this.runningChart = false;
          this.runned.emit('runned');
          this.error = true;
          this.toastService.showError('Error',
            'Error occurred while running visualization');
          return of([]);
        })
      );
  }

  chartEvent($event: KpiValues) {
    if (typeof this.visualization.chartAction === 'string') {
      this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
    }
    const echartClickAction: EchartClickAction = {
      data: {name: $event.bucketKey}
    };
    if (!this.building) {
      this.utmChartClickActionService.onClickNavigate(this.visualization, echartClickAction);
    }
  }

  extractKpiValues(): Promise<KpiValues[]> {
    return new Promise<KpiValues[]>(resolve => {
      const kpi: KpiValues[] = [];
      if (this.data) {
        const dat = this.data;
        for (const d of dat) {
          if (typeof this.visualization.chartConfig === 'string') {
            JSON.parse(this.visualization.chartConfig);
          }
          if (typeof this.visualization.chartConfig !== 'string') {
            const metricIndex = this.visualization.aggregationType.metrics.findIndex(value => Number(value.id) === Number(d.metricId));
            const optionIndex = this.visualization.chartConfig.findIndex(value => Number(value.metricId) === Number(d.metricId));
            kpi.push({
              value: d.value,
              label: extractMetricLabel(this.visualization.aggregationType.metrics[metricIndex].id, this.visualization),
              group: this.extractGroupName(d.bucketKey),
              bucketKey: d.bucketKey,
              color: optionIndex > -1 ? this.visualization.chartConfig[optionIndex].color : null,
              icon: optionIndex > -1 ? this.visualization.chartConfig[optionIndex].icon : null,
              decimal: optionIndex > -1 ? this.visualization.chartConfig[optionIndex].decimal : null
            });
          }
        }
      }
      resolve(kpi);
    });
  }

  extractGroupName(bucketKey: string): string {
    return this.visualization.aggregationType.bucket ?
      (this.visualization.aggregationType.bucket.customLabel === '' ? bucketKey :
        this.visualization.aggregationType.bucket.customLabel + ' - ' + bucketKey) : null;
  }

  onTimeFilterChange($event: TimeFilterType) {
    rebuildVisualizationFilterTime($event, this.visualization.filterType).then(filters => {
      this.visualization.filterType = filters;
      this.refreshService.sendRefresh(this.refreshType);
    });
  }

  ngOnDestroy(): void {
    this.refreshService.stopInterval();
    this.destroy$.next();
    this.destroy$.complete();
  }
}

export class KpiValues {
  icon?: string;
  label?: string;
  color?: string;
  value?: number;
  group?: string;
  decimal?: number;
  bucketKey?: any;
}
