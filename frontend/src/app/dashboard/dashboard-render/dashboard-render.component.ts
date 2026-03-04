import {AfterViewInit, ChangeDetectionStrategy, ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {CompactType, GridsterConfig, GridsterItem, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {IComponent} from '../../graphic-builder/dashboard-builder/shared/services/layout.service';
import {rebuildVisualizationFilterTime} from '../../graphic-builder/shared/util/chart-filter/chart-filter.util';
import {DashboardBehavior} from '../../shared/behaviors/dashboard.behavior';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../shared/chart/types/dashboard/utm-dashboard.type';
import {VisualizationType} from '../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {RefreshService} from '../../shared/services/util/refresh.service';
import {DashboardFilterType} from '../../shared/types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {mergeParams, sanitizeFilters} from '../../shared/util/elastic-filter.util';
import {filtersToStringParam} from '../../shared/util/query-params-to-filter.util';
import {normalizeString} from '../../shared/util/string-util';
import {RenderLayoutService} from '../shared/services/render-layout.service';

@Component({
  selector: 'app-dashboard-render',
  templateUrl: './dashboard-render.component.html',
  styleUrls: ['./dashboard-render.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
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
  layout$: Observable<{ grid: GridsterItem, visualization: VisualizationType } []>;

  constructor(private activatedRoute: ActivatedRoute,
              private layoutService: RenderLayoutService,
              private cdr: ChangeDetectorRef,
              private dashboardBehavior: DashboardBehavior,
              private timeFilterBehavior: TimeFilterBehavior,
              private exportPdfService: ExportPdfService,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {

    document.body.classList.add('overflow-hidden');
    this.loadingVisualizations = true;

    this.layout$ = this.activatedRoute.data.pipe(
      map(data => {
        const response = data && data.response ? data.response : [];

        this.dashboard = this.layoutService.dashboard ? this.layoutService.dashboard : null;

        try {
          this.filters = this.dashboard && this.dashboard.filters
            ? JSON.parse(this.dashboard.filters)
            : [];
        } catch {
          this.filters = [];
        }

        this.loadingVisualizations = false;

        return response;
      })
    );


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
      const safeParams = new URLSearchParams(queryParams).toString();
      const url = `/dashboard/export/${this.dashboard.id}/${normalizeString(this.dashboard.name)}?${safeParams}`;

      this.exportPdfService.getPdf(url, this.dashboard.name, 'PDF_TYPE_TOKEN').subscribe(response => {
        this.spinner.hide('buildPrintPDF').then(() =>
          this.exportPdfService.handlePdfResponse(response));
      }, error => {
        this.spinner.hide('buildPrintPDF');
        console.error('Error downloading PDF:', error);
      });
    });
  }

  trackByFn(index: number, item: { grid: GridsterItem, visualization: VisualizationType }) {
    return item.visualization.id;
  }

  ngOnDestroy(): void {
    document.body.classList.remove('overflow-hidden');
    clearInterval(this.interval);
    this.layoutService.clearLayout();
    this.dashboardBehavior.$filterDashboard.next(null);
  }
}

