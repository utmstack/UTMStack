import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {UtmRenderVisualization} from '../../dashboard/shared/services/utm-render-visualization.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {ComplianceParamsEnum} from '../shared/enums/compliance-params.enum';
import {ComplianceEndpointService} from '../shared/services/compliance-endpoint.service';
import {ComplianceTemplateService} from '../shared/services/compliance-template.service';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';
import {HippaSignaturesType} from '../shared/type/hippa-signatures.type';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {filtersToStringParam} from '../../shared/util/query-params-to-filter.util';
import {rebuildVisualizationFilterTime} from '../../graphic-builder/shared/util/chart-filter/chart-filter.util';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {NgxSpinnerService} from 'ngx-spinner';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';

@Component({
  selector: 'app-compliance-result-view',
  templateUrl: './compliance-result-view.component.html',
  styleUrls: ['./compliance-result-view.component.scss']
})
export class ComplianceResultViewComponent implements OnInit, OnDestroy {
  reportId: number;
  report: ComplianceReportType;
  signatures: HippaSignaturesType[] = [];
  dashboardId: number;
  UUID = UUID.UUID();
  visualizationRender: UtmDashboardVisualizationType[];
  loadingVisualizations = true;
  interval: any;
  dashboard: UtmDashboardType;
  pdfExport = false;
  public options: GridsterConfig = {
    gridType: GridType.ScrollVertical,
    setGridSize: true,
    compactType: CompactType.None,
    minCols: 30,
    // maxCols: 30,
    minRows: 1,
    minItemRows: 1,
    fixedRowHeight: 430,
    fixedColWidth: 500,
    // maxItemRows: 100,
    defaultItemCols: 1,
    defaultItemRows: 1,
    draggable: {
      enabled: false,
    },
    resizable: {
      enabled: false,
    },
    // itemChangeCallback: LayoutService.itemChange,
    swap: false,
    disableWindowResize: false,
  };
  standardId: number;
  sectionId: number;
  configSolution: string;
  filtersValues: ElasticFilterType[] = [];
  destroy$: Subject<void> = new Subject<void>();

  constructor(private activeRoute: ActivatedRoute,
              private cpReportsService: CpReportsService,
              private complianceEndpointService: ComplianceEndpointService,
              private utmToastService: UtmToastService,
              private modalService: NgbModal,
              private complianceTemplateService: ComplianceTemplateService,
              private utmRenderVisualization: UtmRenderVisualization,
              private timeFilterBehavior: TimeFilterBehavior,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService) {

    this.activeRoute.queryParams.subscribe((params) => {
      this.reportId = params[ComplianceParamsEnum.TEMPLATE];
      this.standardId = params[ComplianceParamsEnum.STANDARD_ID];
      this.sectionId = params[ComplianceParamsEnum.SECTION_ID];
    });
  }

  ngOnInit() {
    this.getTemplate();

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

  /**
   * Return template
   */
  getTemplate() {
    this.cpReportsService.find(this.reportId).subscribe(response => {
      this.report = response.body;
      if (this.report.dashboardId) {
        this.loadVisualizations(this.report.dashboardId);
      }
    });
  }

  loadVisualizations(dashboardId) {
    this.dashboardId = dashboardId;
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
    }
  }
  exportToPdf() {
    filtersToStringParam(this.filtersValues).then(queryParams => {
      this.spinner.show('buildPrintPDF');
      const params = queryParams !== '' ? '?' + queryParams : '';
      const url = '/dashboard/export-compliance/' + this.reportId +  params;
      const fileName = this.report.associatedDashboard.name.replace(/ /g, '_');
      this.exportPdfService.getPdf(url, fileName, 'PDF_TYPE_TOKEN').subscribe(response => {
        this.spinner.hide('buildPrintPDF').then(() =>
          this.exportPdfService.handlePdfResponse(response));
      }, error => {
        this.spinner.hide('buildPrintPDF').then(() =>
          this.utmToastService.showError('Error', 'An error occurred while creating a PDF.'));
      });
    });
  }
  viewSolution(solution: string): void {
    this.configSolution = solution;
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
