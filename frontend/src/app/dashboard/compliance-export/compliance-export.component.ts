import {AfterViewInit, ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {ActivatedRoute} from '@angular/router';
import {CompactType, GridsterConfig, GridType} from 'angular-gridster2';
import {UUID} from 'angular2-uuid';
import {NgxSpinnerService} from 'ngx-spinner';
import {CpReportsService} from '../../compliance/shared/services/cp-reports.service';
import {ComplianceReportType} from '../../compliance/shared/type/compliance-report.type';
import {AccountService} from '../../core/auth/account.service';
import {Account} from '../../core/user/account.model';
import {DashboardBehavior} from '../../shared/behaviors/dashboard.behavior';
import {ThemeChangeBehavior} from '../../shared/behaviors/theme-change.behavior';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {
  parseQueryParamsToFilterWithPattern
} from '../../shared/util/query-params-to-filter.util';
import {buildFormatInstantFromDate} from '../../shared/util/utm-time.util';
import {UtmRenderVisualization} from '../shared/services/utm-render-visualization.service';


@Component({
  selector: 'app-compliance-export',
  templateUrl: './compliance-export.component.html',
  styleUrls: ['./compliance-export.component.scss']
})
export class ComplianceExportComponent implements OnInit, AfterViewInit {
  dashboardId: number;
  visualizationRender: UtmDashboardVisualizationType[];
  loadingVisualizations = true;
  printFormat = true;
  chartTypeEnum = ChartTypeEnum;
  reportId: number;
  report: ComplianceReportType;
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
  account: Account;
  date = new Date();
  preparingPrint = true;
  cover: string;
  filters: ElasticFilterType[] = [];
  filterTime: { from: string, to: string };

  constructor(private activatedRoute: ActivatedRoute,
              private cpReportsService: CpReportsService,
              private utmRenderVisualization: UtmRenderVisualization,
              private accountService: AccountService,
              private spinner: NgxSpinnerService,
              private themeChangeBehavior: ThemeChangeBehavior,
              private dashboardBehavior: DashboardBehavior,
              private timeFilterBehavior: TimeFilterBehavior,
              public sanitizer: DomSanitizer,
              private cdr: ChangeDetectorRef) {
  }

  ngAfterViewInit(): void {
    this.themeChangeBehavior.$themeReportCover.subscribe(img => {
      this.cover = img;
    });
    this.cdr.detectChanges();
  }


  print() {
    this.printFormat = true;
    setTimeout(() => {
      window.print();
    }, 2000);
  }

  ngOnInit() {
   /* this.spinner.show('buildPrint');
    window.addEventListener('beforeprint', (event) => {
      this.printFormat = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.printFormat = false;
    });*/
    this.activatedRoute.params.subscribe(params => {
      this.reportId = params.id;
      this.getTemplate();
    });
    this.activatedRoute.queryParams.subscribe(params => {
      const queryParams = Object.entries(params).length > 0 ? params : null;
      if (queryParams) {
        parseQueryParamsToFilterWithPattern(queryParams).then((filters) => {
          this.filters = filters;
          this.getTimeFilterValue();
        });
      }
    });
    this.accountService.identity().then(account => {
      this.account = account;
    });
  }

  getTimeFilterValue() {
    if (this.getTime()) {
      this.filterTime = {
        from: this.resolveFromDate(this.getTime()),
        to: this.resolveToDate(this.getTime()),
      };
    }
  }

  getTime() {
    const indexTime = this.filters.findIndex(value => value.field === '@timestamp');
    if (indexTime !== -1) {
      return {
        from: this.filters[indexTime].value[0],
        to: this.filters[indexTime].value[1]
      };
    }
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

  resolveToDate(date: { from: any, to: any }): string {
    if (!isNaN(Date.parse(date.to))) {
      return date.to;
    } else {
      return new Date().toString();
    }
  }

  setVisFilter(): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      for (const dashFilter of this.getFilterByIndexPattern()) {
        this.dashboardBehavior.$filterDashboard.next(dashFilter);
      }
      if (this.filterTime) {
        this.timeFilterBehavior.$time.next(this.getTime());
      }
      setTimeout(() => resolve(true), 5000);
    });
  }

  getFilterByIndexPattern(): { filter: ElasticFilterType[], indexPattern: string }[] {
    const filterDefs = [];

    this.filters.forEach(filterType => {
      const existingFilterDef = filterDefs.find(def => def.indexPattern === filterType.pattern);

      if (existingFilterDef) {
        existingFilterDef.filter.push(filterType);
      } else {
        filterDefs.push({
          indexPattern: filterType.pattern,
          filter: [filterType]
        });
      }
    });
    return filterDefs;
  }

  onVisualizationLoaded() {
    this.preparingPrint = false;
    console.log('onVisualizationLoaded');
    /*this.setVisFilter().then(() => {
      this.spinner.hide('buildPrint').then(() => {
        this.preparingPrint = false;
        this.print();
      });
    });*/
  }

  resolveFromDate(date: { from: any, to: any }): string {
    if (date && !isNaN(Date.parse(date.from))) {
      return date.from;
    } else {
      return buildFormatInstantFromDate(date).timeFrom;
    }
  }

}
