import {AfterViewInit, ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CompactType, GridsterConfig, GridsterItem, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {IComponent} from '../../graphic-builder/dashboard-builder/shared/services/layout.service';
import {rebuildVisualizationFilterTime} from '../../graphic-builder/shared/util/chart-filter/chart-filter.util';
import {DashboardBehavior} from '../../shared/behaviors/dashboard.behavior';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {VisualizationType} from '../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {DashboardFilterType} from '../../shared/types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {mergeParams, sanitizeFilters} from '../../shared/util/elastic-filter.util';
import {filtersToStringParam} from '../../shared/util/query-params-to-filter.util';
import {normalizeString} from '../../shared/util/string-util';
import {RenderLayoutService} from '../shared/services/render-layout.service';
import {UtmRenderVisualization} from '../shared/services/utm-render-visualization.service';
import {ExportPdfService} from "../../shared/services/util/export-pdf.service";
import {NgxSpinnerService} from "ngx-spinner";

@Component({
  selector: 'app-dashboard-render',
  templateUrl: './dashboard-render.component.html',
  styleUrls: ['./dashboard-render.component.scss']
})
export class DashboardRenderComponent implements OnInit, OnDestroy, AfterViewInit {
  dashboardId: number;
  UUID = UUID.UUID();
  visualizationRender: UtmDashboardVisualizationType[];
  loadingVisualizations = true;
  interval: any;
  dashboard: UtmDashboardType;
  pdfExport = false;
  filters: DashboardFilterType[];
  activeTimeGridster: number;
  public options: GridsterConfig = {
    gridType: GridType.ScrollVertical,
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
    // itemResizeCallback: DashboardCreateComponent.itemResize,
    swap: false
  };
  filtersValues: ElasticFilterType[] = [];
  timeEnable: number[] = [];

  constructor(private activatedRoute: ActivatedRoute,
              private layoutService: RenderLayoutService,
              private cdr: ChangeDetectorRef,
              private modalService: NgbModal,
              private dashboardBehavior: DashboardBehavior,
              private timeFilterBehavior: TimeFilterBehavior,
              private utmRenderVisualization: UtmRenderVisualization,
              private exportPdfService: ExportPdfService,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    console.log('Init dashboard');

    document.body.classList.add('overflow-hidden');
    this.loadingVisualizations = true;
    this.activatedRoute.params.subscribe(params => {
      this.dashboardId = params.id;
      if (this.dashboardId) {
        const request = {
          page: 0,
          size: 10000,
          'idDashboard.equals': this.dashboardId,
          sort: 'order,asc'
        };
        this.utmRenderVisualization.query(request).subscribe(response => {
          this.layoutService.layout = [];
          this.timeEnable = [];
          this.visualizationRender = response.body;
          this.dashboard = this.visualizationRender.length > 0 ? this.visualizationRender[0].dashboard : null;
          this.filters = this.dashboard.filters ? JSON.parse(this.dashboard.filters) : [];
          if (this.dashboard.refreshTime) {
            this.onRefreshTime(this.dashboard.refreshTime);
          }
          for (const vis of this.visualizationRender) {
            if (vis.showTimeFilter) {
              this.timeEnable.push(vis.visualization.id);
            }
            const grid = JSON.parse(vis.gridInfo);
            this.layoutService.addItem(vis.visualization, grid);
          }
          this.loadingVisualizations = false;
        });
      }
    });
    this.dashboardBehavior.$filterDashboard.subscribe(dashboardFilter => {
      if (dashboardFilter) {
        mergeParams(dashboardFilter.filter, this.filtersValues).then(newFilters => {
          this.filtersValues = sanitizeFilters(newFilters);
        });
      }
    });
    this.timeFilterBehavior.$time.subscribe(time => {
      if (time) {
        rebuildVisualizationFilterTime({timeFrom: time.from, timeTo: time.to}, this.filtersValues).then(filters => {
          this.filtersValues = filters;
        });
      }
    });
  }


  get layout(): { grid: GridsterItem, visualization: VisualizationType } [] {
    return this.layoutService.layout;
  }

  get components(): IComponent[] {
    return this.layoutService.components;
  }


  populateDashboard($event: VisualizationType[], override?: boolean) {
    for (const vis of $event) {
      const indexVis = this.layout.findIndex(value => value.visualization.id === vis.id);
      vis.chartConfig = (typeof vis.chartConfig === 'string' && vis.chartType !== ChartTypeEnum.TEXT_CHART) ?
        JSON.parse(vis.chartConfig) : vis.chartConfig;
      if (indexVis === -1) {
        this.layoutService.addItem(vis);
      }
    }
  }

  ngOnDestroy(): void {
    document.body.classList.remove('overflow-hidden');
    clearInterval(this.interval);
    this.visualizationRender = [];
    this.layoutService.layout = [];
    this.timeEnable = [];
    this.dashboardBehavior.$filterDashboard.next(null);
  }

  ngAfterViewInit(): void {
    this.cdr.detectChanges();
  }

  onRefreshTime(time: number) {
    this.interval = setInterval(() => {
    }, time);
  }

  exportToPdf() {
    filtersToStringParam(this.filtersValues).then(queryParams => {
      this.spinner.show('buildPrintPDF');
      const url = '/dashboard/export/' + this.dashboardId + '/' + normalizeString(this.dashboard.name) + '?' + queryParams;
      // window.open('/dashboard/export/' + this.dashboardId + '/' + normalizeString(this.dashboard.name) + '?' + queryParams, '_blank');
      this.exportPdfService.getPdf(url, this.dashboard.name, 'PDF_TYPE_TOKEN').subscribe(response => {
        this.spinner.hide('buildPrintPDF').then(() =>
          this.exportPdfService.handlePdfResponse(response));
      }, error => {
        this.spinner.hide('buildPrintPDF');
        console.error('Error downloading PDF:', error);
      });
    });
  }
}

