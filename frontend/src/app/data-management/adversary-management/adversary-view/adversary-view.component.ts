import { Component, OnInit } from '@angular/core';
import {UtmToastService} from "../../../shared/alert/utm-toast.service";
import {
  ElasticFilterDefaultTime
} from '../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {
  ALERT_PARENT_ID,
  ALERT_STATUS_FIELD_AUTO,
  ALERT_TAGS_FIELD, ALERT_TIMESTAMP_FIELD,
  FALSE_POSITIVE_OBJECT
} from '../../../shared/constants/alert/alert-field.constant';
import {AUTOMATIC_REVIEW} from '../../../shared/constants/alert/alert-status.constant';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {AlertDataTypeBehavior} from '../../alert-management/shared/behavior/alert-data-type.behavior';
import {AlertFiltersBehavior} from '../../alert-management/shared/behavior/alert-filters.behavior';
import {EventDataTypeEnum} from '../../alert-management/shared/enums/event-data-type.enum';
import {AdversaryAlerts} from '../models';
import {AdversaryAlertService} from '../services/adversary-alert-.service';

@Component({
  selector: 'app-adversary-view',
  templateUrl: './adversary-view.component.html',
  styleUrls: ['./adversary-view.component.scss']
})
export class AdversaryViewComponent implements OnInit {

  filters: ElasticFilterType[] = [
    {field: ALERT_STATUS_FIELD_AUTO, operator: ElasticOperatorsEnum.IS_NOT, value: AUTOMATIC_REVIEW},
    {field: ALERT_TAGS_FIELD, operator: ElasticOperatorsEnum.IS_NOT, value: FALSE_POSITIVE_OBJECT.tagName},
    {field: ALERT_TIMESTAMP_FIELD, operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-7d', 'now']}
  ];
  defaultTime: ElasticFilterDefaultTime;
  adversaryAlerts: AdversaryAlerts[] = [];
  loading = false;

  constructor(private alertDataTypeBehavior: AlertDataTypeBehavior,
              private alertFiltersBehavior: AlertFiltersBehavior,
              private adversaryAlertService: AdversaryAlertService,
              private toastService: UtmToastService) { }

  ngOnInit() {
    this.alertDataTypeBehavior.$alertDataType.next(EventDataTypeEnum.ADVERSARY);
    this.alertFiltersBehavior.$filters.next(this.filters);
    this.defaultTime = new ElasticFilterDefaultTime('now-7d', 'now');

    this.loadData();
  }


  onFilterChange($event: any) {
    this.processFilters($event).then(filters => {
      this.filters = filters;
      this.loadData();
      // this.updateStatusServiceBehavior.$updateStatus.next(true);
      this.alertFiltersBehavior.$filters.next(this.filters);
    });
  }

  processFilters(filter: ElasticFilterType): Promise<ElasticFilterType[]> {
    return new Promise<ElasticFilterType[]>(resolve => {
      const indexFilters = this.filters.findIndex(value => filter.field === value.field && value.value !== FALSE_POSITIVE_OBJECT.tagName);
      if (indexFilters === -1) {
        this.filters.push(filter);
      } else {
        this.filters[indexFilters] = filter;
      }
      resolve(this.filters);
    });
  }

  onTimeFilterChange($event: TimeFilterType) {
    const timeFilterIndex = this.filters.findIndex(value => value.field === ALERT_TIMESTAMP_FIELD);
    if (timeFilterIndex === -1) {
      this.filters.push({
        field: ALERT_TIMESTAMP_FIELD,
        value: [$event.timeFrom, $event.timeTo],
        operator: ElasticOperatorsEnum.IS_BETWEEN
      });
    } else {
      this.filters[timeFilterIndex].value = [$event.timeFrom, $event.timeTo];
    }
    this.alertFiltersBehavior.$filters.next(this.filters);
    this.loadData();
  }

  loadData() {
    this.loading = true;
    this.adversaryAlertService.query(this.filters)
      .subscribe( data => {
        this.adversaryAlerts = data ? data : [];
        this.loading = false;
      },
       error => {
        this.loading = false;
        this.adversaryAlerts = [];
        this.toastService.showError('Error', 'An error occurred while loading adversary alerts data.');
       });
  }
}
