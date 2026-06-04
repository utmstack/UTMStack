import {Component, Input, OnInit} from '@angular/core';
import {ITEMS_PER_PAGE} from '../../../../../shared/constants/pagination.constants';
import {ElasticTimeEnum} from '../../../../../shared/enums/elastic-time.enum';
import {UtmDateFormatEnum} from '../../../../../shared/enums/utm-date-format.enum';
import {AlertHistoryType} from '../../../../../shared/types/alert/alert-history.type';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterCommonType} from '../../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {AlertUpdateHistoryBehavior} from '../../behavior/alert-update-history.behavior';
import {AlertHistoryActionEnum} from '../../enums/alert-history-action.enum';
import {AlertHistoryService} from './alert-history.service';

@Component({
  selector: 'app-alert-history',
  templateUrl: './alert-history.component.html',
  styleUrls: ['./alert-history.component.scss']
})
export class AlertHistoryComponent implements OnInit {
  @Input() alert: UtmAlertType;
  sevenDaysRange: ElasticFilterCommonType = {last: 1, time: ElasticTimeEnum.YEAR, label: 'last 1 year'};
  items: AlertHistoryType[] = [];
  loadingMore = false;
  page = 1;
  filterTime: TimeFilterType;
  loading = true;
  formatDateEnum = UtmDateFormatEnum;
  action: AlertHistoryActionEnum;
  noMoreResult = false;
  totalItems: any;
  itemsPerPage = ITEMS_PER_PAGE;
  actions = [
    AlertHistoryActionEnum.STATUS,
    AlertHistoryActionEnum.TAG,
    AlertHistoryActionEnum.NOTE,
    AlertHistoryActionEnum.INCIDENT,
    AlertHistoryActionEnum.SOLUTION,
  ];

  constructor(private alertHistoryService: AlertHistoryService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior) {
  }

  ngOnInit(): void {
    this.alertUpdateHistoryBehavior.$refreshHistory.subscribe(value => {
      if (value) {
        this.page = 1;
        this.items = [];
        this.getAlertHistory();
      }
    });
  }

  getAlertHistory() {
    const req = {
      'alertId.equals': this.alert.id,
      'logAction.equals': this.action,
      'logDate.greaterThanOrEqual': this.filterTime.timeFrom,
      'logDate.lessThanOrEqual': this.filterTime.timeTo,
      page: this.page - 1,
      size: 8,
      sort: 'logDate,desc'
    };
    this.alertHistoryService.query(req).subscribe(response => {
      this.loadingMore = false;
      this.loading = false;
      this.alertUpdateHistoryBehavior.$refreshHistory.next(null);
      this.items = response.body;
      this.totalItems = Number(response.headers.get('X-Total-Count'));
    });
  }

  onScroll() {
    this.loadingMore = true;
    this.page += 1;
    this.getAlertHistory();
  }

  resolveClassByAction(action: AlertHistoryActionEnum): string {
    switch (action) {
      case AlertHistoryActionEnum.TAG:
        return 'utm_tmlabel_tag';
      case AlertHistoryActionEnum.NOTE:
        return 'utm_tmlabel_note';
      case AlertHistoryActionEnum.STATUS:
        return 'utm_tmlabel_status';
      case AlertHistoryActionEnum.INCIDENT:
        return 'utm_tmlabel_incident';
      case AlertHistoryActionEnum.SOLUTION:
        return 'utm_tmlabel_solution';
      default:
        return 'utm_tmlabel_not_found';
    }
  }

  onFilterTimeChange($event: TimeFilterType) {
    this.filterTime = $event;
    this.page = 1;
    this.items = [];
    this.getAlertHistory();
  }

  filterByAction($event) {
    this.items = [];
    this.page = 1;
    this.action = $event;
    this.getAlertHistory();
  }

  loadPage($event: number) {
    this.page = $event;
    this.getAlertHistory();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.getAlertHistory();

  }
}
