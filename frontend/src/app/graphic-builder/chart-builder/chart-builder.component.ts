import {AfterViewChecked, ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {ChartActionType} from '../../shared/chart/types/action/chart-action.type';
import {Color} from '../../shared/chart/types/charts/chart-properties/color/color';
import {Grid} from '../../shared/chart/types/charts/chart-properties/grid/grid';
import {Legend} from '../../shared/chart/types/charts/chart-properties/legend/legend';
import {Toolbox} from '../../shared/chart/types/charts/chart-properties/toolbox/toolbox';
import {UtmGaugeOptionType} from '../../shared/chart/types/charts/gauge/utm-gauge-option.type';
import {UtmGoalOption} from '../../shared/chart/types/charts/goal/utm-goal-option';
import {HeatMapPropertiesType} from '../../shared/chart/types/charts/heatmap/heat-map-properties.type';
import {UtmMetricOptions} from '../../shared/chart/types/charts/metric/utm-metric-options';
import {UtmScatterMapOptionType} from '../../shared/chart/types/charts/scatter/utm-scatter-map-option.type';
import {UtmTagCloudOptionType} from '../../shared/chart/types/charts/tag-cloud/utm-tag-cloud-option.type';
import {MetricAggregationType} from '../../shared/chart/types/metric/metric-aggregation.type';
import {MetricBucketsType} from '../../shared/chart/types/metric/metric-buckets.type';
import {VisualizationType} from '../../shared/chart/types/visualization.type';
import {
  ElasticFilterDefaultTime
} from '../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {UTM_CHART_ICONS} from '../../shared/constants/icons-chart.const';
import {MULTIPLE_METRIC_CHART} from '../../shared/constants/visualization-bucket-metric.constant';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../shared/enums/nature-data.enum';
import {RouteCallbackEnum} from '../../shared/enums/route-callback.enum';
import {ElasticSearchIndexService} from '../../shared/services/elasticsearch/elasticsearch-index.service';
import {FieldDataService} from '../../shared/services/elasticsearch/field-data.service';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {RunVisualizationBehavior} from '../shared/behavior/run-visualization.behavior';
import {VisualizationQueryParamsEnum} from '../shared/enums/visualization-query-params.enum';
import {VisualizationService} from '../visualization/shared/services/visualization.service';
import {VisualizationSaveComponent} from '../visualization/visualization-save/visualization-save.component';
import {VisualizationBehavior} from './chart-property-builder/shared/behaviors/visualization.behavior';
import {Location} from "@angular/common";
import {DashboardStatusEnum} from "../dashboard-builder/shared/enums/dashboard-status.enum";

@Component({
  selector: 'app-chart-builder',
  templateUrl: './chart-builder.component.html',
  styleUrls: ['./chart-builder.component.scss']
})
export class ChartBuilderComponent implements OnInit, AfterViewChecked {
  chart: ChartTypeEnum;
  type: DataNatureTypeEnum;
  chartTypeEnum = ChartTypeEnum;
  visualization: VisualizationType = null;
  mode: string;
  visualizationId: number;
  display: 'data' | 'options';
  headers = [{name: 'data', label: 'Data'}, {name: 'options', label: 'Options'}, {name: 'click', label: 'On click'}];
  property = this.headers[0].name;
  metricNoErrors = true;
  bucketNoErrors = true;
  configChange = true;
  running = false;
  multipleMetrics = MULTIPLE_METRIC_CHART;
  pattern: string;
  tempId: number;
  callback: RouteCallbackEnum;
  private patternId: number;
  defaultTime = new ElasticFilterDefaultTime('now-24h', 'now');


  constructor(private spinner: NgxSpinnerService,
              private route: ActivatedRoute,
              private modalService: NgbModal,
              private fieldDataBehavior: FieldDataService,
              private visualizationBehavior: VisualizationBehavior,
              private cdr: ChangeDetectorRef,
              private indexPatternFieldService: ElasticSearchIndexService,
              private visualizationService: VisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private location: Location,
              private router: Router) {
    route.queryParams.subscribe(params => {
      this.chart = params[VisualizationQueryParamsEnum.CHART];
      this.mode = params[VisualizationQueryParamsEnum.MODE];
      this.pattern = params[VisualizationQueryParamsEnum.PATTERN_NAME];
      this.patternId = Number(params[VisualizationQueryParamsEnum.PATTERN_ID]);
      if (params[VisualizationQueryParamsEnum.CALLBACK]) {
        this.callback = params[VisualizationQueryParamsEnum.CALLBACK];
      }
      this.visualizationId = params[VisualizationQueryParamsEnum.VISUALIZATION_ID];
    });

  }

  ngOnInit() {
    this.tempId = Math.floor(Math.random() * (1000000 - 20000 + 1) + 20000);
    this.getFields();
    if (this.mode === 'edit') {
      this.visualizationService.find(this.visualizationId).subscribe(vis => {
        this.visualization = vis.body;
        const defaultFilterTime = this.getDefaultFilterTimeFromVisualization(this.visualization.filterType);
        this.defaultTime = defaultFilterTime ? defaultFilterTime : new ElasticFilterDefaultTime('now-24h', 'now');
      });
    } else {
      this.visualization = {
        aggregationType: {metrics: [], bucket: null},
        chartConfig: this.chart === this.chartTypeEnum.METRIC_CHART ? [] : {
          legend: new Legend(true, []),
          colors: new Color(),
          toolbox: new Toolbox(),
          grid: new Grid()
        },
        chartAction: new ChartActionType(false),
        filterType: [{field: '@timestamp', operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-24h', 'now']}],
        idPattern: this.patternId,
        chartType: this.chart,
        eventType: this.type,
        userCreated: null,
        name: '',
        pattern: {
          id: this.patternId,
          pattern: this.pattern
        },
      };
    }
  }

  getFields() {
    this.fieldDataBehavior.getFields(this.pattern)
      .subscribe(fields => {
        this.fieldDataBehavior.$fields.next(fields);
      });
  }

  ngAfterViewChecked(): void {
    this.cdr.detectChanges();
  }

  viewProperty($event: string) {
    this.property = $event;
  }

  runVisualization() {
    this.running = true;
    this.runVisualizationBehavior.$run.next(this.tempId);
  }

  saveVisualization() {
    const modal = this.modalService.open(VisualizationSaveComponent, {centered: true});
    modal.componentInstance.visualization = this.visualization;
    modal.componentInstance.callback = this.callback;
    modal.componentInstance.mode = this.mode;
  }

  chartIconResolver(): string {
    return UTM_CHART_ICONS[this.chart];
  }

  onFilterChange($event: ElasticFilterType[]) {
    this.configChange = true;
    this.visualization.filterType = $event;
  }

  onMetricChange($event: MetricAggregationType[]) {
    this.configChange = true;
    this.visualization.aggregationType.metrics = $event;
  }

  onMetricErrors($event: boolean) {
    this.metricNoErrors = $event;
  }

  onBucketChange($event: MetricBucketsType) {
    this.configChange = true;
    this.visualization.aggregationType.bucket = $event;
  }

  onBucketErrors($event: boolean) {
    this.bucketNoErrors = $event;
  }

  onMetricOptionChange($event: UtmMetricOptions[]) {
    this.visualization.chartConfig = $event;
  }

  onRunVisualization($event: string) {
    this.configChange = false;
    this.running = false;
  }

  resolveConfStatus(): { icon: string, label: string, class: string } {
    let conf: { icon: string, label: string, class: string } = null;
    if (!this.metricNoErrors) {
      conf = {
        icon: 'icon-warning22',
        label: 'Metric has error please check in order to run',
        class: 'text-danger-800'
      };
    } else if (!this.bucketNoErrors) {
      conf = {
        icon: 'icon-warning22',
        label: 'Bucket has error please check in order to run',
        class: 'text-danger-800'
      };
    } else if (this.metricNoErrors && this.bucketNoErrors && this.configChange) {
      conf = {
        icon: 'icon-checkmark-circle',
        label: 'Configuration change, may need run visualization to view result',
        class: 'text-success-800'
      };
    }
    return conf;
  }

  onPieOptionsChange($event) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onGaugeOptionChange($event: UtmGaugeOptionType) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onGoalOptionChange($event: UtmGoalOption[]) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onTableOptionChange($event: any) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onTagOptionsChange($event: UtmTagCloudOptionType) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onLineOptionsChange($event: any) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onScatterMapOptions($event: UtmScatterMapOptionType) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onHeatMapOptionChange($event: HeatMapPropertiesType) {
    this.configChange = true;
    this.visualization.chartConfig = $event;
  }

  onChartActionChange($event: ChartActionType) {
    this.visualization.chartAction = $event;
  }

  getDefaultFilterTimeFromVisualization(filters: ElasticFilterType[] | null): ElasticFilterDefaultTime {
    if (filters) {
      const indexTime = filters.findIndex(value => value.field === '@timestamp');
      if (indexTime !== -1) {
        return new ElasticFilterDefaultTime(filters[indexTime].value[0], filters[indexTime].value[1]);
      } else {
        return null;
      }
    } else {
      return null;
    }
  }

  cancel() {
    if (this.callback === RouteCallbackEnum.DASHBOARD) {
      this.router.navigate(['/creator/dashboard/builder'],
          {
            queryParams:
                {
                  mode: DashboardStatusEnum.DRAFT
                }
          });
    } else {
      this.location.back();
    }
  }
}
