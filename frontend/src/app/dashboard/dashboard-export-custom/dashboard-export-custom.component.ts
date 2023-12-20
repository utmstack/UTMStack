import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {captureScreen, pdfPreview} from '../../shared/util/export-to-pdf-util';
import {UtmRenderVisualization} from '../shared/services/utm-render-visualization.service';

@Component({
  selector: 'app-dashboard-export-custom',
  templateUrl: './dashboard-export-custom.component.html',
  styleUrls: ['./dashboard-export-custom.component.scss']
})
export class DashboardExportCustomComponent implements OnInit {
  dashboardId: number;
  UUID = UUID.UUID();
  loadingDashboard = true;
  loadingPreview = true;
  dashboard: UtmDashboardType;
  pdfExport = false;
  visualization: UtmDashboardVisualizationType[];
  gridsterOptions: GridsterConfig = {
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
      enabled: true,
    },
    resizable: {
      enabled: true,
    },
    itemChangeCallback: DashboardExportCustomComponent.itemChange,
    swap: true,
    disableWindowResize: false,
  };
  activeModal: number;
  preview = false;
  pdf: any;
  build = false;
  showNotice = true;

  constructor(private activatedRoute: ActivatedRoute,
              private modalService: NgbModal,
              private utmRenderVisualization: UtmRenderVisualization) {
  }

  static itemChange(item, itemComponent) {
  }

  ngOnInit() {
    this.activatedRoute.params.subscribe(params => {
      this.dashboardId = params.id;
      if (this.dashboardId) {
        const request = {
          page: 0,
          size: 10000,
          'idDashboard.equals': this.dashboardId,
          sort: 'order,asc'
        };
        this.utmRenderVisualization.query(request).subscribe(vis => {
          this.visualization = vis.body;
          this.dashboard = this.visualization.length > 0 ? this.visualization[0].dashboard : null;
          this.loadingDashboard = false;
        });
      }
    });
  }

  exportToPdf() {
    this.pdfExport = true;
    captureScreen(this.UUID, this.dashboard.name).then((finish) => {
      this.pdfExport = false;
    });
  }

  deleteVisualization(index: number) {
    this.visualization.splice(index, 1);
  }

  buildPdf() {
    this.build = true;
    pdfPreview(this.UUID, this.dashboard.name).then(pdf => {
      this.build = false;
      setTimeout(() => {
        this.pdf = pdf;
        this.preview = true;
      }, 500);
    });
  }
}
