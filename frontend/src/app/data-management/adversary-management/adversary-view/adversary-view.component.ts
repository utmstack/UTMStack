import { Component, OnInit } from '@angular/core';
import {
  ALERT_PARENT_ID,
  ALERT_STATUS_FIELD_AUTO,
  ALERT_TAGS_FIELD, ALERT_TIMESTAMP_FIELD,
  FALSE_POSITIVE_OBJECT
} from '../../../shared/constants/alert/alert-field.constant';
import {AUTOMATIC_REVIEW} from '../../../shared/constants/alert/alert-status.constant';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {AlertDataTypeBehavior} from '../../alert-management/shared/behavior/alert-data-type.behavior';
import {AlertFiltersBehavior} from '../../alert-management/shared/behavior/alert-filters.behavior';
import {EventDataTypeEnum} from '../../alert-management/shared/enums/event-data-type.enum';

@Component({
  selector: 'app-adversary-view',
  templateUrl: './adversary-view.component.html',
  styleUrls: ['./adversary-view.component.scss']
})
export class AdversaryViewComponent implements OnInit {

  filters: ElasticFilterType[] = [
    {field: ALERT_STATUS_FIELD_AUTO, operator: ElasticOperatorsEnum.IS_NOT, value: AUTOMATIC_REVIEW},
    {field: ALERT_TAGS_FIELD, operator: ElasticOperatorsEnum.IS_NOT, value: FALSE_POSITIVE_OBJECT.tagName},
    {field: ALERT_PARENT_ID, operator: ElasticOperatorsEnum.DOES_NOT_EXIST},
    {field: ALERT_TIMESTAMP_FIELD, operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-7d', 'now']}
  ];

  constructor(private alertDataTypeBehavior: AlertDataTypeBehavior,
              private alertFiltersBehavior: AlertFiltersBehavior) { }

  ngOnInit() {
    this.alertDataTypeBehavior.$alertDataType.next(EventDataTypeEnum.ADVERSARY);
    this.alertFiltersBehavior.$filters.next(this.filters);
  }


  protected onFilterChange($event: any) {

  }
}
