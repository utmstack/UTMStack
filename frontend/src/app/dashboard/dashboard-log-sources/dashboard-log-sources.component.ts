import {Component, OnInit} from '@angular/core';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {Observable} from 'rxjs';
import {UtmDashboardService} from '../../graphic-builder/dashboard-builder/shared/services/utm-dashboard.service';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {LOG_SOURCE_DASHBOARD_NAME} from '../../shared/constants/global.constant';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {UtmRenderVisualization} from '../shared/services/utm-render-visualization.service';

@Component({
  selector: 'app-dashboard-log-sources',
  templateUrl: './dashboard-log-sources.component.html',
  styleUrls: ['./dashboard-log-sources.component.css']
})
export class DashboardLogSourcesComponent implements OnInit {
  pdfExport = false;
  UUID = UUID.UUID();
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
  visualizations: UtmDashboardVisualizationType[];
  loading = true;
  withDashboard = false;
  chartTypeEnum = ChartTypeEnum;

  constructor(private utmRenderVisualization: UtmRenderVisualization,
              private utmDashboardService: UtmDashboardService) {
  }

  ngOnInit() {
    window.addEventListener('beforeprint', (event) => {
      this.pdfExport = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.pdfExport = false;
    });
    this.getDashboardByName().subscribe(id => {
      if (id !== -1) {
        this.withDashboard = true;
        this.getDashboard(id);
      } else {
        this.withDashboard = false;
      }
    });
  }

  getDashboard(id: number) {
    const request = {
      page: 0,
      size: 10000,
      'idDashboard.equals': id,
      sort: 'order,asc'
    };
    this.utmRenderVisualization.query(request).subscribe(vis => {
      this.visualizations = vis.body;
      this.loading = false;
    });
  }

  getDashboardByName(): Observable<number> {
    return new Observable<number>(subscriber => {
      this.utmDashboardService.query({'name.equals': LOG_SOURCE_DASHBOARD_NAME}).subscribe(response => {
        if (response.body.length > 0) {
          subscriber.next(response.body[0].id);
        } else {
          subscriber.next(-1);
        }
      });
    });
  }

  exportToPdf() {
    this.pdfExport = true;
    // captureScreen('utmDashboardAlert').then((finish) => {
    //   this.pdfExport = false;
    // });
    setTimeout(() => {
      window.print();
    }, 1000);
  }
}
