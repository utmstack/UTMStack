import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {NgxSpinnerService} from 'ngx-spinner';
import {Subject} from 'rxjs';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {ComplianceParamsEnum} from '../shared/enums/compliance-params.enum';
import {ComplianceStatusExtendedEnum} from '../shared/enums/compliance-status.enum';
import {ComplianceStrategyEnum} from '../shared/enums/compliance-strategy.enum';
import {CpControlConfigService} from '../shared/services/cp-control-config.service';
import {ComplianceControlEvaluationType} from '../shared/type/compliance-control-evaluation.type';
import {ComplianceControlEvaluationsType} from '../shared/type/compliance-control-evaluations.type';

@Component({
  selector: 'app-compliance-evaluations-view',
  templateUrl: './compliance-evaluations-view.component.html',
  styleUrls: ['./compliance-evaluations-view.component.scss']
})
export class ComplianceEvaluationsViewComponent implements OnInit, OnDestroy {

  constructor(private activeRoute: ActivatedRoute,
              // private cpReportsService: CpReportsService,
              private cpControlConfigService: CpControlConfigService,
              private timeFilterBehavior: TimeFilterBehavior,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService) {
  }
  @Input() showExport = true;
  @Input() template: 'default' | 'compliance' = 'default';
  controlId: number;
  currentEvaluation: ComplianceControlEvaluationType;
  evaluationsHistory: ComplianceControlEvaluationsType[];
  loading = false;
  showDetails = false;
  showRemediation = false;
  showSolution = false;
  selectedEvaluation: ComplianceControlEvaluationsType;
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
  destroy$: Subject<void> = new Subject<void>();
  showBack = false;

  chartOption: any = {};

  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;
  ComplianceStrategyEnum = ComplianceStrategyEnum;

  ngOnInit() {
    this.loading = true;
    /*this.activeRoute.queryParams
      .pipe(
          takeUntil(this.destroy$),
          filter((params) => Object.keys(params).length > 0),
          tap(() => {
            this.showBack = true;
          }))
      .subscribe((params) => {
       //this.initializeReportParams(params);
    });*/

    this.cpControlConfigService.onLoadControl$
      .pipe(takeUntil(this.destroy$),
            filter(params => !!params),
      ).subscribe(params => {
        this.currentEvaluation = params.template;
        console.log('Current eval: ', this.currentEvaluation);
        if (this.currentEvaluation) {
          this.loadReport(params);
        }

      });

    /*this.timeFilterBehavior.$time
      .pipe(takeUntil(this.destroy$))
      .subscribe(time => {
        if (time) {
          rebuildVisualizationFilterTime({timeFrom: time.from, timeTo: time.to}, this.filtersValues).then(filters => {
            this.filtersValues = filters;
          });
        }
      });*/
  }

  loadReport(params: any) {
    this.showDetails = false;
    this.controlId = params.template && params.template.id ? params.template.id : null;
    this.standardId = params[ComplianceParamsEnum.STANDARD_ID];
    this.sectionId = params[ComplianceParamsEnum.SECTION_ID];
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
      this.loading = true;
      this.cpControlConfigService.evaluationsByControl(this.controlId)
        .subscribe(response => {
          this.evaluationsHistory = response.body.evaluations;
          /*this.chartOption = {
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
              min: new Date(response.body.startDate).getTime(),
              max: new Date(response.body.endDate).getTime(),
              axisLabel: {
                formatter: value => {
                  const d = new Date(value);
                  return d.toLocaleString('en-US', {
                    month: 'short',
                    day: '2-digit',
                    year: 'numeric',
                  });
                }
              },
              axisTick: {
                alignWithLabel: true
              },
              splitLine: {
                show: true
              },
              interval: 1000 * 60 * 60 * 24
            },
            yAxis: { show: false },
            dataZoom: [
              { type: 'slider', start: 0, end: 10 },
              { type: 'inside', start: 0, end: 10 }
            ],
            series: [
              {
                type: 'custom',
                renderItem: (params, api) => {
                  const xValue = api.value(0);
                  const x = api.coord([xValue, 0])[0];

                  const totalWidth = 40;
                  const margin = 2;
                  const width = totalWidth - margin * 2;

                  const yBottom = api.coord([xValue, 0])[1];
                  const yTop = api.coord([xValue, 1])[1];
                  const height = yBottom - yTop;

                  return {
                    type: 'group',
                    children: [
                      {
                        type: 'rect',
                        shape: {
                          x: x - totalWidth / 2 + margin,
                          y: yTop,
                          width,
                          height
                        },
                        style: api.style(),
                        // ⭐ Aquí viene la magia
                        clipPath: {
                          type: 'rect',
                          shape: {
                            x: x - totalWidth / 2 + margin,
                            y: yTop,
                            width,
                            height,
                            r: 5   // ⭐ radio de borde redondeado
                          }
                        }
                      }
                    ]
                  };
                },
                data: []
              }
            ]

          };*/

          //this.chartOption.series[0].data = this.buildChartData();
          //this.chartOption = { ...this.chartOption };
          this.loading = false;
        });
    }
  }

  /*private buildChartData() {
    return this.evaluationsHistory.map(e => ({
      value: [e.timestamp, 0],
      controlId: e.controlId,
      status: e.status,
      timestamp: e.timestamp,
      dateFormatted: new Date(e.timestamp).toLocaleString('en-US', {
        month: 'short',
        day: '2-digit',
        year: 'numeric'
      }),
      itemStyle: {
        color: e.status === 'COMPLIANT' ? '#4CAF50' : '#F44336'
      }
    }));
  }*/

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

  showEvaluationDetails(evaluation: ComplianceControlEvaluationsType) {
    this.showDetails = true;
    console.log(this.selectedEvaluation);
    this.selectedEvaluation = evaluation;
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
