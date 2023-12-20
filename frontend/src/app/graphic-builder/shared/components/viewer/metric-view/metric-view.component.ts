import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {EchartClickAction} from '../../../../../shared/chart/types/action/echart-click-action';
import {MetricResponse} from '../../../../../shared/chart/types/metric/metric-response';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ElasticFilterDefaultTime} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {mergeParams, sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {extractMetricLabel} from '../../../../chart-builder/chart-property-builder/shared/functions/visualization-util';
import {RunVisualizationBehavior} from '../../../behavior/run-visualization.behavior';
import {RunVisualizationService} from '../../../services/run-visualization.service';
import {UtmChartClickActionService} from '../../../services/utm-chart-click-action.service';
import {rebuildVisualizationFilterTime} from '../../../util/chart-filter/chart-filter.util';
import {resolveDefaultVisualizationTime} from '../../../util/visualization/visualization-render.util';

@Component({
  selector: 'app-metric-view',
  templateUrl: './metric-view.component.html',
  styleUrls: ['./metric-view.component.scss']
})
export class MetricViewComponent implements OnInit {
  data: MetricResponse[];
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

  constructor(private runVisualizationService: RunVisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private toastService: UtmToastService,
              private dashboardBehavior: DashboardBehavior,
              private utmChartClickActionService: UtmChartClickActionService) {
  }


  ngOnInit() {
    this.runVisualization();
    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
    this.runVisualizationBehavior.$run.subscribe((id) => {
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
  }

  runVisualization() {
    this.runningChart = true;
    this.runVisualizationService.run(this.visualization).subscribe(data => {
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
    });
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
      this.runVisualization();
    });
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
