import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Observable, of, Subject} from 'rxjs';
import {catchError, filter, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {EchartClickAction} from '../../../../../shared/chart/types/action/echart-click-action';
import {UtmGoalOption} from '../../../../../shared/chart/types/charts/goal/utm-goal-option';
import {MetricResponse} from '../../../../../shared/chart/types/metric/metric-response';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ElasticFilterDefaultTime} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {RefreshService, RefreshType} from '../../../../../shared/services/util/refresh.service';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {mergeParams, sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {extractMetricLabel} from '../../../../chart-builder/chart-property-builder/shared/functions/visualization-util';
import {RunVisualizationBehavior} from '../../../behavior/run-visualization.behavior';
import {RunVisualizationService} from '../../../services/run-visualization.service';
import {UtmChartClickActionService} from '../../../services/utm-chart-click-action.service';
import {rebuildVisualizationFilterTime} from '../../../util/chart-filter/chart-filter.util';
import {resolveDefaultVisualizationTime} from '../../../util/visualization/visualization-render.util';

@Component({
  selector: 'app-goal-view',
  templateUrl: './goal-view.component.html',
  styleUrls: ['./goal-view.component.scss']
})
export class GoalViewComponent implements OnInit, OnDestroy {
  data$: Observable<MetricResponse[]>;
  @Input() chartId: number;
  @Input() visualization: VisualizationType;
  @Input() building: boolean;
  @Input() showTime: boolean;
  @Input() timeByDefault: any;
  @Output() runned = new EventEmitter<string>();
  @Input() width: string;
  @Input() height: string;
  runningChart: boolean;
  goals: UtmGoalOption[] = [];
  GOAL_CHART = ChartTypeEnum.GOAL_CHART;
  error: boolean;
  defaultTime: ElasticFilterDefaultTime;
  destroy$: Subject<any> = new Subject<any>();
  refreshType: string;

  constructor(private runVisualizationService: RunVisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private toastService: UtmToastService,
              private dashboardBehavior: DashboardBehavior,
              private utmChartClickActionService: UtmChartClickActionService,
              private refreshService: RefreshService) {
  }

  ngOnInit() {
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

    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
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
      this.extractGoals();
      this.error = false;
    }, error => {
      this.runningChart = false;
      this.error = true;
      this.runned.emit('runned');
      this.toastService.showError('Error',
        'Error occurred while running visualization');
    });*/

    return this.runVisualizationService.run(this.visualization)
      .pipe(
        tap((data) => {
          this.runningChart = false;
          this.runned.emit('runned');
          this.extractGoals(data);
          this.error = false;
        }),
        catchError(() => {
          this.runningChart = false;
          this.error = true;
          this.runned.emit('runned');
          this.toastService.showError('Error',
            'Error occurred while running visualization');
          return of([]);
        })
      );
  }

  chartEvent($event: UtmGoalOption) {
    if (typeof this.visualization.chartAction === 'string') {
      this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
    }
    const echartClickAction: EchartClickAction = {
      data: {name: $event.label}
    };
    if (!this.building) {
      this.utmChartClickActionService.onClickNavigate(this.visualization, echartClickAction);
    }
  }

  extractGoals(data: MetricResponse[]): UtmGoalOption[] {
    this.goals = [];
    const config: UtmGoalOption[] = this.visualization.chartConfig;
    if (data) {
      for (const d of data) {
        const metricIndex = this.visualization.aggregationType.metrics.findIndex(value => Number(value.id) === Number(d.metricId));
        const optionIndex = config.findIndex(value => Number(value.metricId) === Number(d.metricId));
        const max = (config[optionIndex].max ? config[optionIndex].max : this.calcTotal(data));
        const goal = new UtmGoalOption(Number(d.metricId),
          this.calcPercent(max, d.value, config[optionIndex].decimal),
          config[optionIndex].append,
          max,
          config[optionIndex].animate,
          config[optionIndex].min,
          config[optionIndex].thick,
          config[optionIndex].cap,
          config[optionIndex].type,
          config[optionIndex].thresholds,
          d.bucketKey ? d.bucketKey : extractMetricLabel(d.metricId, this.visualization),
          config[optionIndex].foregroundColor
        );
        this.goals.push(goal);
      }
    }
    return this.goals;
  }

  calcTotal(data: MetricResponse[]): number {
    return data.reduce((accum, item) => accum + item.value, 0);
  }

  calcPercent(goal, value, decimal): number {
    return Number.parseFloat(((value / goal) * 100).toFixed(decimal));
  }

  processThresholds(thresholds: object): object {
    const keys = Object.keys(thresholds);
    // tslint:disable-next-line:prefer-for-of
    for (let i = 0; i < keys.length; i++) {
      const key = '\'' + keys[i] + '\'';
      thresholds[key] = thresholds[keys[i]];
    }
    return thresholds;
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

