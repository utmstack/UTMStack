import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {NgxSpinnerService} from 'ngx-spinner';
import {Subject} from 'rxjs';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {rebuildVisualizationFilterTime} from '../../graphic-builder/shared/util/chart-filter/chart-filter.util';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {ComplianceParamsEnum} from '../shared/enums/compliance-params.enum';
import {CpControlConfigService} from '../shared/services/cp-control-config.service';
import {ComplianceControlEvaluationsType} from '../shared/type/compliance-control-evaluations.type';

@Component({
  selector: 'app-compliance-evaluations-view',
  templateUrl: './compliance-evaluations-view.component.html',
  styleUrls: ['./compliance-evaluations-view.component.scss']
})
export class ComplianceEvaluationsViewComponent implements OnInit, OnDestroy {
  @Input() showExport = true;
  @Input() template: 'default' | 'compliance' = 'default';
  controlId: number;
  evaluations: ComplianceControlEvaluationsType[];
  UUID = UUID.UUID();
  interval: any;
  dashboard: UtmDashboardType;
  pdfExport = false;
  public options: GridsterConfig = {
    gridType: GridType.ScrollVertical,
    setGridSize: true,
    compactType: CompactType.None,
    minCols: 30,
    minRows: 1,
    minItemRows: 1,
    fixedRowHeight: 430,
    fixedColWidth: 500,
    defaultItemCols: 1,
    defaultItemRows: 1,
    draggable: {
      enabled: false,
    },
    resizable: {
      enabled: false,
    },
    swap: false,
    disableWindowResize: false,
  };
  standardId: number;
  sectionId: number;
  configSolution: string;
  filtersValues: ElasticFilterType[] = [];
  destroy$: Subject<void> = new Subject<void>();
  showBack = false;

  option = {
    tooltip: {
      trigger: 'item',
      formatter: params => {
        const e = params.data;
        return `
        <b>${e.dateFormatted}</b><br/>
        Status: <b>${e.status}</b>
      `;
      }
    },
    xAxis: {
      type: 'time',
      name: 'Evaluations over time'
    },
    yAxis: {
      show: false
    },
    dataZoom: [
      {
        type: 'slider',
        show: true,
        xAxisIndex: 0,
        filterMode: 'none',
        start: 0,
        end: 10
      },
      {
        type: 'inside',
        xAxisIndex: 0,
        filterMode: 'none',
        start: 0,
        end: 10
      }
    ],
    series: [
      {
        type: 'scatter',
        symbolSize: 20, // ← importante para que se vea como punto
        data: [] // ← se llena después
      }
    ]
  };

  constructor(private activeRoute: ActivatedRoute,
              // private cpReportsService: CpReportsService,
              private cpControlConfigService: CpControlConfigService,
              private timeFilterBehavior: TimeFilterBehavior,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService) {
  }

  ngOnInit() {
    this.activeRoute.queryParams
      .pipe(
          takeUntil(this.destroy$),
          filter((params) => Object.keys(params).length > 0),
          tap(() => {
            this.showBack = true;
          }))
      .subscribe((params) => {
        this.initializeReportParams(params);
    });

    this.cpControlConfigService.onLoadControl$
      .pipe(takeUntil(this.destroy$),
            filter(params => !!params),
      ).subscribe(params => {
        this.loadReport(params);
      });

    this.timeFilterBehavior.$time
      .pipe(takeUntil(this.destroy$))
      .subscribe(time => {
        if (time) {
          rebuildVisualizationFilterTime({timeFrom: time.from, timeTo: time.to}, this.filtersValues).then(filters => {
            this.filtersValues = filters;
          });
        }
      });
  }

  loadReport(params: any) {
    this.controlId = params.template && params.template.id ? params.template.id : null;
    this.standardId = params[ComplianceParamsEnum.STANDARD_ID];
    this.sectionId = params[ComplianceParamsEnum.SECTION_ID];
    this.evaluations = params.template;
    this.getEvaluations();
  }

  initializeReportParams(params: any) {
    this.controlId = params[ComplianceParamsEnum.TEMPLATE];
    this.standardId = params[ComplianceParamsEnum.STANDARD_ID];
    this.sectionId = params[ComplianceParamsEnum.SECTION_ID];

    this.getEvaluations();
  }

  getEvaluations() {
    if (this.controlId) {
      this.cpControlConfigService.evaluationsByControl(this.controlId)
        .subscribe(response => {
          this.evaluations = response.body;
          console.log('Evaluations: ', this.evaluations);
          this.option.series[0].data = this.buildChartData();
          this.option = { ...this.option };
        });
    }
  }

  private buildChartData() {
    return this.evaluations.map(e => ({
      value: [e.timestamp, 0],
      controlId: e.controlId,
      status: e.status,
      timestamp: e.timestamp,
      dateFormatted: new Date(e.timestamp).toLocaleString(),
      itemStyle: {
        color: e.status === 'COMPLIANT' ? '#4CAF50' : '#F44336'
      }
    }));
  }

  loadVisualizations(dashboardId) {
    /*this.dashboardId = dashboardId;
    if (this.dashboardId) {
      const request = {
        page: 0,
        size: 10000,
        'idDashboard.equals': this.dashboardId,
        sort: 'order,asc'
      };
      this.utmRenderVisualization.query(request).subscribe(vis => {
        this.visualizationRender = vis.body;
        this.loadingVisualizations = false;
      });
    }*/
  }
  exportToPdf() {
    /*filtersToStringParam(this.filtersValues).then(queryParams => {
      this.spinner.show('buildPrintPDF');
      const params = queryParams !== '' ? '?' + queryParams : '';
      const url = '/dashboard/export-compliance/' + this.controlId +  params;
      const fileName = this.control.associatedDashboard.name.replace(/ /g, '_');
      this.exportPdfService.getPdf(url, fileName, 'PDF_TYPE_TOKEN').subscribe(response => {
        this.spinner.hide('buildPrintPDF').then(() =>
          this.exportPdfService.handlePdfResponse(response));
      }, error => {
        this.spinner.hide('buildPrintPDF').then(() =>
          this.utmToastService.showError('Error', 'An error occurred while creating a PDF.'));
      });
    });*/
  }
  viewSolution(solution: string): void {
    this.configSolution = solution;
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onChartInit(chart: any) {
    /*this.chartInstance = chart;

    // Limpia listeners previos para evitar duplicados al refrescar datos
    this.chartInstance.off('click');

    // Listener principal
    this.chartInstance.on('click', (params: any) => {
      const point = params.data;

      // Encuentra la evaluación seleccionada
      const evaluation = this.evaluations.find(
        e => e.timestamp === point.timestamp
      );

      if (evaluation) {
        this.selectedEvaluation = evaluation;
        // Aquí puedes abrir panel lateral, modal, etc.
      }
    });
     */
  }
}
