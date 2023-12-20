import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ElasticFilterDefaultTime} from '../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {ElasticDataService} from '../../shared/services/elasticsearch/elastic-data.service';
import {ElasticFilterType} from '../../shared/types/filter/elastic-filter.type';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {AppLogTypeEnum} from './shared/enum/app-log-type.enum';
import {AppLogType} from './shared/type/app-log.type';

@Component({
  selector: 'app-app-logs',
  templateUrl: './app-logs.component.html',
  styleUrls: ['./app-logs.component.css']
})
export class AppLogsComponent implements OnInit {
  logs: AppLogType[] = [];
  links: any;
  totalItems: number;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-7d', 'now');
  loading = true;
  sortBy = '@timestamp,desc';
  message: string;
  req: { filters: ElasticFilterType[], index: string, top: number } = {
    filters: [
      {field: '@timestamp', operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-7d', 'now']}
    ], index: '.utmstack-logs*', top: 10000000
  };
  sources = ['PANEL', 'AGENT'];
  types = ['ERROR', 'WARNING', 'INFO'];

  constructor(private elasticDataService: ElasticDataService) {
    this.itemsPerPage = ITEMS_PER_PAGE;
  }

  ngOnInit() {
    this.loadAll();
  }


  loadAll() {
    this.elasticDataService.genericSearchInIndex(this.req, this.page, this.itemsPerPage, this.sortBy)
      .subscribe(
        (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
        (res: HttpResponse<any>) => this.onError(res.body)
      );
  }

  loadPage(page: number) {
    this.page = page;
    this.loadAll();
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.logs = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  getClassByType(type: AppLogTypeEnum) {
    switch (type) {
      case AppLogTypeEnum.ERROR:
        return 'border-danger-600 text-danger-600';
      case AppLogTypeEnum.INFO:
        return 'border-info-600 text-info-600';
      case AppLogTypeEnum.WARNING:
        return 'border-warning text-warning';
    }
  }

  onTimeFilterChange($event: TimeFilterType) {
    const index = this.req.filters.findIndex(value => value.field === '@timestamp');
    this.req.filters[index].value = [$event.timeFrom, $event.timeTo];
    this.loadAll();
  }

  filterBySelect($event: any, field: 'source' | 'type') {
    const index = this.req.filters.findIndex(value => value.field === field);
    if (index === -1) {
      this.req.filters.push({field, operator: ElasticOperatorsEnum.IS, value: $event});
    } else {
      if ($event) {
        this.req.filters[index].value = $event;
      } else {
        this.req.filters.splice(index, 1);
      }
    }
    this.loadAll();
  }

  filterByType($event: any) {

  }

  getPreview(message: string): string {
    const preview = message.substring(0, message.indexOf(':')).substring(0, 200);
    return preview.length > 200 ? (preview + '..') : preview;
  }
}
