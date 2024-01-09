import {Component, Input, OnInit} from '@angular/core';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {IncidentHistoryActionEnum} from '../../../../shared/enums/incident/incident-history-action.enum';
import {UtmDateFormatEnum} from '../../../../shared/enums/utm-date-format.enum';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {IncidentHistoryType} from '../../../../shared/types/incident/incident-history.type';
import {UtmIncidentType} from '../../../../shared/types/incident/utm-incident.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {resolveInstantDate} from '../../../../shared/util/utm-time.util';
import {UtmIncidentHistoryService} from '../../services/utm-incident-history.service';

@Component({
  selector: 'app-incident-history',
  templateUrl: './incident-history.component.html',
  styleUrls: ['./incident-history.component.scss']
})
export class IncidentHistoryComponent implements OnInit {
  @Input() incident: UtmIncidentType;
  sevenDaysRange: ElasticFilterCommonType = {last: 30, time: ElasticTimeEnum.DAY, label: 'last 30 days'};
  items: IncidentHistoryType[] = [];
  loadingMore = false;
  page = 1;
  filterTime: TimeFilterType = resolveInstantDate(ElasticTimeEnum.DAY, 30);
  loading = true;
  formatDateEnum = UtmDateFormatEnum;
  action: IncidentHistoryActionEnum;
  noMoreResult = false;
  totalItems: any;
  itemsPerPage = 15;
  actions: IncidentHistoryActionEnum[] = [];

  constructor(private incidentHistoryService: UtmIncidentHistoryService) {
  }

  ngOnInit(): void {
    this.page = 1;
    this.items = [];
    this.getIncidentHistory();

  }

  getIncidentHistory() {
    const req = {
      'incidentId.equals': this.incident.id,
      'actionType.equals': this.action,
      'actionDate.greaterThanOrEqual': this.filterTime.timeFrom,
      'actionDate.lessThanOrEqual': this.filterTime.timeTo,
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: 'actionDate,desc'
    };
    this.incidentHistoryService.query(req).subscribe(response => {
      this.loadingMore = false;
      this.loading = false;
      this.items = response.body;
      this.totalItems = Number(response.headers.get('X-Total-Count'));
      this.actions = Array.from(new Set(response.body.concat(this.items).map( item => item.actionType)));
    });
  }

  onScroll() {
    this.loadingMore = true;
    this.page += 1;
    this.getIncidentHistory();
  }

  resolveClassByAction(action: IncidentHistoryActionEnum): string {
    switch (action) {
      case IncidentHistoryActionEnum.INCIDENT_CREATED:
        return 'utm_tmlabel_incident';
      case IncidentHistoryActionEnum.INCIDENT_NOTE_ADD:
        return 'utm_tmlabel_note';
      case IncidentHistoryActionEnum.INCIDENT_STATUS_CHANGE:
        return 'utm_tmlabel_status';
      case IncidentHistoryActionEnum.INCIDENT_ALERT_ADD:
        return 'utm_tmlabel_alert';
      case IncidentHistoryActionEnum.INCIDENT_ALERT_STATUS_CHANGED:
        return 'utm_tmlabel_status';
      case IncidentHistoryActionEnum.INCIDENT_COMMAND_EXECUTED:
        return 'utm_tmlabel_command';
      default:
        return 'utm_tmlabel_not_found';
    }
  }

  onFilterTimeChange($event: TimeFilterType) {
    this.filterTime = $event;
    this.page = 1;
    this.items = [];
    this.getIncidentHistory();
  }

  filterByAction($event) {
    this.items = [];
    this.page = 1;
    this.action = $event;
    this.getIncidentHistory();
  }

  loadPage($event: number) {
    this.page = $event;
    this.getIncidentHistory();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.getIncidentHistory();

  }
}
