import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {ComplianceTemplateService} from '../../compliance/shared/services/compliance-template.service';
import {UtmRenderVisualization} from '../../dashboard/shared/services/utm-render-visualization.service';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {normalizeString} from '../../shared/util/string-util';
import {ReportParamsEnum} from '../shared/enums/report-params.enum';
import {ReportService} from '../shared/service/report.service';
import {ReportType} from '../shared/type/report.type';

@Component({
  selector: 'app-report-template-result',
  templateUrl: './report-template-result.component.html',
  styleUrls: ['./report-template-result.component.scss']
})
export class ReportTemplateResultComponent implements OnInit {
  reportId: number;
  report: ReportType;
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

  constructor(private activeRoute: ActivatedRoute,
              private reportsService: ReportService,
              private utmToastService: UtmToastService,
              private modalService: NgbModal,
              private complianceTemplateService: ComplianceTemplateService,
              private utmRenderVisualization: UtmRenderVisualization) {

    this.activeRoute.queryParams.subscribe((params) => {
      this.reportId = params[ReportParamsEnum.TEMPLATE];
    });
  }

  ngOnInit() {
    this.getTemplate();
  }

  /**
   * Return template
   */
  getTemplate() {
    this.reportsService.find(this.reportId).subscribe(response => {
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
    window.open('/dashboard/export-report/' + this.report.id + '/' + normalizeString(this.report.repName), '_blank');
  }
}
