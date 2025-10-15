import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {
  ALERT_ADVERSARY_FIELD, ALERT_ECHOES_FIELDS, ALERT_PARENT_ID,
  ALERT_STATUS_FIELD_AUTO, ALERT_TAGS_FIELD,
  ALERT_TARGET_FIELD, ALERT_TIMESTAMP_FIELD, FALSE_POSITIVE_OBJECT
} from '../../../../../shared/constants/alert/alert-field.constant';
import {AUTOMATIC_REVIEW} from '../../../../../shared/constants/alert/alert-status.constant';
import {ITEMS_PER_PAGE} from '../../../../../shared/constants/pagination.constants';
import {SortDirection} from '../../../../../shared/directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../../../../shared/directives/sortable/type/sort-event';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../../../../shared/enums/nature-data.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {EventDataTypeEnum} from '../../enums/event-data-type.enum';

@Component({
  selector: 'app-alert-echoes',
  templateUrl: './alert-echoes.component.html',
  styleUrls: ['./alert-echoes.component.scss']
})
export class AlertEchoesComponent implements OnInit {

  @Input() alert: UtmAlertType = {} as UtmAlertType;

  page = 1;
  totalItems = 0;
  readonly fields = ALERT_ECHOES_FIELDS;
  readonly ALERT_ADVERSARY_FIELD = ALERT_ADVERSARY_FIELD;
  readonly ALERT_TARGET_FIELD = ALERT_TARGET_FIELD;
  itemsPerPage = ITEMS_PER_PAGE;
  dataType = EventDataTypeEnum.ALERT;
  dataNature = DataNatureTypeEnum.ALERT;
  loading = false;
  sortEvent: SortEvent;
  alerts: UtmAlertType[] = [];
  filters: ElasticFilterType[] = [
    {field: ALERT_STATUS_FIELD_AUTO, operator: ElasticOperatorsEnum.IS_NOT, value: AUTOMATIC_REVIEW},
    {field: ALERT_TAGS_FIELD, operator: ElasticOperatorsEnum.IS_NOT, value: FALSE_POSITIVE_OBJECT.tagName},
  ];
  direction: SortDirection = '';
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';

  constructor(private elasticDataService: ElasticDataService,
              private utmToastService: UtmToastService) { }

  ngOnInit() {

    this.filters.push({
      field: ALERT_PARENT_ID,
      operator: ElasticOperatorsEnum.IS,
      value: this.alert.id
    });
    this.loadChildrenAlerts();
  }

  loadChildrenAlerts() {
      this.loading = true;
      this.elasticDataService.search(this.page, this.itemsPerPage,
        100000000, this.dataNature,
        sanitizeFilters(this.filters), this.sortBy, true).subscribe(
        (res: HttpResponse<any>) => {
          this.totalItems = Number(res.headers.get('X-Total-Count'));
          this.alerts = res.body;
          this.loading = false;
        },
        (res: HttpResponse<any>) => {
          this.utmToastService.showError('Error', 'An error occurred while listing the alerts. Please try again later.');
          this.loading = false;
        }
      );
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.loadChildrenAlerts();
  }

  loadPage($event: number) {
    this.page = $event;
    this.loadChildrenAlerts();
  }

  onRefreshData($event: boolean) {

  }

  onItemsPerPageChange($event: number) {
    this.page = 1;
    this.itemsPerPage = $event;
    this.loadChildrenAlerts();
  }
}
