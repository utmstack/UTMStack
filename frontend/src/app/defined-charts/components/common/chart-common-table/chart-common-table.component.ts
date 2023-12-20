import {Component, Input, OnDestroy, OnInit, QueryList, ViewChildren} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {TableBuilderResponseType} from '../../../../shared/chart/types/response/table-builder-response.type';
import {SortableDirective} from '../../../../shared/directives/sortable/sortable.directive';
import {SortDirection} from '../../../../shared/directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {OverviewAlertDashboardService} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-common-table',
  templateUrl: './chart-common-table.component.html',
  styleUrls: ['./chart-common-table.component.scss']
})
export class ChartCommonTableComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  @Input() endpoint: string;
  @Input() header: string;
  @Input() navigateUrl: string;
  @Input() params;
  @Input() paramClick: string;
  interval: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  @ViewChildren(SortableDirective) headers: QueryList<SortableDirective>;
  totalItems: number;
  page = 1;
  itemsPerPage = 6;
  pageStart = 0;
  pageEnd = 6;
  data: TableBuilderResponseType;
  direction: SortDirection = '';
  loadingOption = true;
  responseRows: Array<{ value: any; metric: boolean }[]>;
  queryParams = {from: 'now-7d', to: 'now', top: 20};

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.queryParams.top = 20;
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getTopAlert();
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  getTopAlert() {
    this.overviewAlertDashboardService.getDataTable(this.endpoint, this.queryParams)
      .subscribe(response => {
        this.data = response.body;
        this.loadingOption = false;
        this.pageStart = 0;
        this.pageEnd = this.itemsPerPage;
        this.responseRows = this.data.rows;
        this.totalItems = this.data.rows.length;
      });
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

  onSort({column, direction}: SortEvent) {
    this.direction = direction;
    this.headers.forEach(header => {
      if (header.sortable !== column) {
        header.direction = '';
      }
    });
    if (direction === '') {
      this.data.rows = this.responseRows;
    } else {
      this.data.rows = this.data.rows.sort((a, b) => {
        return a[1][column] > b[1][column] ? 1 : -1;
      });
    }
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.queryParams.from = $event.timeFrom;
    this.queryParams.to = $event.timeTo;
    this.getTopAlert();
  }

  rowEvent(row: { value: any; metric: boolean }[]) {
    this.spinner.show('loadingSpinner');
    this.params[this.paramClick] = row[0];
    this.params['@timestamp'] =
      ElasticOperatorsEnum.IS_BETWEEN + '->' + this.queryParams.from + ',' + this.queryParams.to;
    this.router.navigate([this.navigateUrl], {
      queryParams: this.params
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

}
