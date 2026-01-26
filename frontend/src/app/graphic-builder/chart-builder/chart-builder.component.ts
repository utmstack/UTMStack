import {Location} from '@angular/common';
import {AfterViewChecked, ChangeDetectorRef, Component, OnInit, ViewChild} from '@angular/core';
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
import {CodeEditorComponent, ConsoleOptions} from '../../shared/components/code-editor/code-editor.component';
import {
  ElasticFilterDefaultTime
} from '../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {UTM_CHART_ICONS} from '../../shared/constants/icons-chart.const';
import {ALERT_INDEX_PATTERN, LOG_INDEX_PATTERN} from '../../shared/constants/main-index-pattern.constant';
import {MULTIPLE_METRIC_CHART} from '../../shared/constants/visualization-bucket-metric.constant';
import {ChartBuilderQueryLanguageEnum} from '../../shared/enums/chart-builder-query-language.enum';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../shared/enums/nature-data.enum';
import {RouteCallbackEnum} from '../../shared/enums/route-callback.enum';
import {SqlValidationService} from '../../shared/services/code-editor/sql-validation.service';
import {FieldDataService} from '../../shared/services/elasticsearch/field-data.service';
import {LocalFieldService} from '../../shared/services/elasticsearch/local-field.service';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {UtmIndexPattern} from '../../shared/types/index-pattern/utm-index-pattern';
import {RunVisualizationBehavior} from '../shared/behavior/run-visualization.behavior';
import {VisualizationQueryParamsEnum} from '../shared/enums/visualization-query-params.enum';
import {VisualizationService} from '../visualization/shared/services/visualization.service';
import {VisualizationSaveComponent} from '../visualization/visualization-save/visualization-save.component';

import {DashboardStatusEnum} from '../dashboard-builder/shared/enums/dashboard-status.enum';
import {HttpResponse} from "@angular/common/http";
import {IndexPatternService} from "../../shared/services/elasticsearch/index-pattern.service";
import {map} from "rxjs/operators";


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

  @ViewChild(CodeEditorComponent) codeEditor: CodeEditorComponent;
  isSqlMode = false;
  errorMessage = '';
  sqlQuery = '';
  indexPatternNames: string[] = [];
  codeEditorOptions: ConsoleOptions = {lineNumbers: 'off'};
  loading = true;
  indexPattern: UtmIndexPattern[];

  constructor(private spinner: NgxSpinnerService,
              private route: ActivatedRoute,
              private modalService: NgbModal,
              private fieldDataBehavior: FieldDataService,
              private cdr: ChangeDetectorRef,
              private visualizationService: VisualizationService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private location: Location,
              private router: Router,
              private localFieldService: LocalFieldService,
              private sqlValidationService: SqlValidationService,
              private indexPatternService: IndexPatternService) {
    route.queryParams.subscribe(params => {
      this.chart = params[VisualizationQueryParamsEnum.CHART];
      this.mode = params[VisualizationQueryParamsEnum.MODE];
      if (params[VisualizationQueryParamsEnum.CALLBACK]) {
        this.callback = params[VisualizationQueryParamsEnum.CALLBACK];
      }
      this.visualizationId = params[VisualizationQueryParamsEnum.VISUALIZATION_ID];
    });

  }

  ngOnInit() {
    this.tempId = Math.floor(Math.random() * (1000000 - 20000 + 1) + 20000);
    if (this.mode === 'edit') {
      this.visualizationService.find(this.visualizationId).subscribe(vis => {
        this.visualization = vis.body;
        if (this.visualization.sqlQuery) {
          this.visualization.queryLanguage = ChartBuilderQueryLanguageEnum.SQL;
          this.sqlQuery = this.visualization.sqlQuery;
          this.isSqlMode = true;
        } else {
          this.visualization.queryLanguage = ChartBuilderQueryLanguageEnum.DSL;
          this.pattern = this.visualization.pattern.pattern;
          this.patternId = this.visualization.pattern.id;
          this.isSqlMode = false;
        }
        const defaultFilterTime = this.getDefaultFilterTimeFromVisualization(this.visualization.filterType);
        this.defaultTime = defaultFilterTime ? defaultFilterTime : new ElasticFilterDefaultTime('now-24h', 'now');
        this.loading = false;
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
        idPattern: null,
        chartType: this.chart,
        eventType: this.type,
        userCreated: null,
        name: '',
        pattern: null,
        queryLanguage: ChartBuilderQueryLanguageEnum.DSL
      };
      this.loading = false;
    }

    this.getIndexPatterns();
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
    if (this.isSqlMode) {
      this.errorMessage = this.sqlValidationService.validateSqlQuery(this.sqlQuery);
      if (this.errorMessage) {
        this.running = false;
        return;
      }
      this.visualization.sqlQuery = this.sqlQuery;
      this.visualization.queryLanguage = ChartBuilderQueryLanguageEnum.SQL;
    } else {
      this.visualization.queryLanguage = ChartBuilderQueryLanguageEnum.DSL;
    }
    this.runVisualizationBehavior.$run.next(this.tempId);
  }

  saveVisualization() {
    if (this.isSqlMode && this.sqlQuery === '') {
      this.errorMessage = 'SQL Query cannot be empty';
      return;
    }
    const modal = this.modalService.open(VisualizationSaveComponent, {centered: true});
    if (this.isSqlMode) {
      this.nullifyUnusedFields();
      this.visualization.queryLanguage = ChartBuilderQueryLanguageEnum.SQL;
      this.visualization.sqlQuery = this.sqlQuery;
    } else {
      this.visualization.queryLanguage = ChartBuilderQueryLanguageEnum.DSL;
      this.visualization.sqlQuery = '';
    }
    modal.componentInstance.visualization = this.visualization;
    modal.componentInstance.callback = this.callback;
    modal.componentInstance.mode = this.mode;
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

  toggleSqlMode($event: boolean) {
    this.visualization.sqlQuery = '';
    this.isSqlMode = $event;
  }

  indexPatternSelected(pattern: UtmIndexPattern) {
    this.pattern = pattern.pattern;
    this.visualization.pattern = pattern;
    this.visualization.idPattern = pattern.id;
    this.patternId = pattern.id;
    this.getFields();
  }

  loadFieldNames() {
    return [
      ...this.localFieldService.getPatternStoredFields(ALERT_INDEX_PATTERN).map(f => f.name),
      ...this.localFieldService.getPatternStoredFields(LOG_INDEX_PATTERN).map(f => f.name)
    ];
  }

  indexPatternLoaded(indexPatternNames: string[]) {
    this.indexPatternNames = indexPatternNames;
  }

  nullifyUnusedFields() {
    this.visualization.aggregationType = null;
  }

  clearMessages(): void {
    this.errorMessage = '';
  }

  getIndexPatterns() {
    const req = {
      page: 0,
      size: 1000,
      sort: 'id,asc',
      'isActive.equals': true,
    };
    this.indexPatternService.query(req)
      .pipe(
        map((res: HttpResponse<any>) => res.body || [])
      )
      .subscribe({
        next: data => {
          this.indexPatternNames = data.map((pattern: UtmIndexPattern) => pattern.pattern);
          this.indexPattern = data;
        },
        error: () => this.indexPatternNames = [],
      });
  }

   onIndexPatternChange($event: string) {
    const indexPattern = this.indexPattern.find(p => p.pattern === $event);

    if (indexPattern) {
      this.indexPatternSelected(indexPattern);
    }
  }
}
