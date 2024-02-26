import {AfterViewInit, ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {ActivatedRoute} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {AccountService} from '../../core/auth/account.service';
import {Account} from '../../core/user/account.model';
import {DashboardBehavior} from '../../shared/behaviors/dashboard.behavior';
import {ThemeChangeBehavior} from '../../shared/behaviors/theme-change.behavior';
import {TimeFilterBehavior} from '../../shared/behaviors/time-filter.behavior';
import {UtmDashboardVisualizationType} from '../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {DashboardFilterType} from '../../shared/types/filter/dashboard-filter.type';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {parseQueryParamsToFilter} from '../../shared/util/query-params-to-filter.util';
import {buildFormatInstantFromDate} from '../../shared/util/utm-time.util';
import {UtmRenderVisualization} from '../shared/services/utm-render-visualization.service';


@Component({
  selector: 'app-dashboard-export-pdf',
  templateUrl: './dashboard-export-pdf.component.html',
  styleUrls: ['./dashboard-export-pdf.component.scss']
})
export class DashboardExportPdfComponent implements OnInit, AfterViewInit {
  dashboardId: number;
  dashboardName: string;
  visualizationRender: UtmDashboardVisualizationType[];
  loadingVisualizations = true;
  printFormat = true;
  chartTypeEnum = ChartTypeEnum;
  dashboardDescription: string;
  date = new Date();
  account: Account;
  preparingPrint = true;
  dashboardFilters: DashboardFilterType[] = [];
  filtersValues: ElasticFilterType[] = [];
  filterTime: { from: string, to: string };
  cover: string;

  constructor(private activatedRoute: ActivatedRoute,
              private accountService: AccountService,
              private utmRenderVisualization: UtmRenderVisualization,
              private spinner: NgxSpinnerService,
              private dashboardBehavior: DashboardBehavior,
              private timeFilterBehavior: TimeFilterBehavior,
              private themeChangeBehavior: ThemeChangeBehavior,
              public sanitizer: DomSanitizer,
              private cdr: ChangeDetectorRef) {
  }

  ngAfterViewInit(): void {
    this.cdr.detectChanges();
    this.themeChangeBehavior.$themeReportCover.subscribe(img => {
      this.cover = img;
    });
  }

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(params => {
      const queryParams = Object.entries(params).length > 0 ? params : null;
      if (queryParams) {
        parseQueryParamsToFilter(queryParams).then((filters) => {
          this.filtersValues = filters;
          this.getTimeFilterValue();
        });
      }
    });
    this.spinner.show('buildPrint');
    this.accountService.identity().then(account => {
      this.account = account;
    });
    window.addEventListener('beforeprint', (event) => {
      this.printFormat = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.printFormat = false;
    });
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
          this.visualizationRender = vis.body;
          this.dashboardName = this.visualizationRender[0].dashboard.name;
          this.dashboardDescription = this.visualizationRender[0].dashboard.description;
          const filters = JSON.parse(this.visualizationRender[0].dashboard.filters);
          this.dashboardFilters = filters ? filters : [];
          this.loadingVisualizations = false;
        });
      }
    });
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
    const indexPatterns = Object.keys(
      this.dashboardFilters.reduce((r, {indexPattern}) => (r[indexPattern] = '', r), {}));
    const filters: {
      filter: ElasticFilterType[],
      indexPattern: string
    }[] = [];
    for (const ip of indexPatterns) {
      const filterIp: {
        filter: ElasticFilterType[],
        indexPattern: string
      } = {indexPattern: ip, filter: []};
      const dashFilters = this.dashboardFilters.filter(value => value.indexPattern === ip);
      for (const df of dashFilters) {
        filterIp.filter.push(this.getFilter(df));
      }
      filters.push(filterIp);
    }
    return filters;
  }

  getFilter(dashFilter: DashboardFilterType): ElasticFilterType {
    const indexFilter = this.filtersValues.findIndex(value => value.field === dashFilter.field);
    if (indexFilter !== -1) {
      return this.filtersValues[indexFilter];
    }
  }

  print() {
    this.printFormat = true;
    setTimeout(() => {
      window.print();
    }, 2000);
  }

  onVisualizationLoaded() {
    this.setVisFilter().then(() => {
      this.spinner.hide('buildPrint').then(() => {
        this.preparingPrint = false;
        // this.print();
      });
    });
  }

  getTimeFilterValue() {
    this.filterTime = {
      from: this.resolveFromDate(this.getTime()),
      to: this.resolveToDate(this.getTime()),
    };
  }

  getTime() {
    const indexTime = this.filtersValues.findIndex(value => value.field === '@timestamp');
    if (indexTime !== -1) {
      return {
        from: this.filtersValues[indexTime].value[0],
        to: this.filtersValues[indexTime].value[1]
      };
    }
  }

  resolveFromDate(date: { from: any, to: any }): string {
    if (!isNaN(Date.parse(date.from))) {
      return date.from;
    } else {
      return buildFormatInstantFromDate(date).timeFrom;
    }
  }

  resolveToDate(date: { from: any, to: any }): string {
    if (!isNaN(Date.parse(date.to))) {
      return date.to;
    } else {
      return new Date().toString();
    }
  }

  getFilterValue(filter: DashboardFilterType): string {
    const indexFilter = this.filtersValues.findIndex(value => value.field === filter.field);
    if (indexFilter !== -1) {
      return this.filtersValues[indexFilter].value.toString();
    }
  }

  showFilters(): boolean {
    return this.filtersValues.filter(value => value.field !== '@timestamp').length > 0;
  }
}
