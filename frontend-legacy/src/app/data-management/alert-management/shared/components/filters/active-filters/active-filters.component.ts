import {Component, Input, OnInit} from '@angular/core';
import {
  ALERT_GLOBAL_FIELD,
  ALERT_INCIDENT_FLAG_FIELD,
  ALERT_SEVERITY_FIELD,
  ALERT_STATUS_FIELD,
  EVENT_IS_ALERT
} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {EventDataTypeEnum} from '../../../enums/event-data-type.enum';
import {resolveFieldNameByFilter} from '../../../util/alert-util-function';

@Component({
  selector: 'app-active-filters',
  templateUrl: './active-filters.component.html',
  styleUrls: ['./active-filters.component.scss']
})
export class ActiveFiltersComponent implements OnInit {
  @Input() filters: ElasticFilterType[] = [];
  @Input() dataType: EventDataTypeEnum;
  operatorsEnum = ElasticOperatorsEnum;
  STATUS_FIELD = ALERT_STATUS_FIELD;
  SEVERITY_FIELD = ALERT_SEVERITY_FIELD;
  GLOBAL_FIELD = ALERT_GLOBAL_FIELD;
  EVENT_IS_ALERT = EVENT_IS_ALERT;
  ALERT_INCIDENT_FLAG_FIELD = ALERT_INCIDENT_FLAG_FIELD;

  constructor() {
  }

  ngOnInit() {
  }

  resolveFieldName(filter: ElasticFilterType) {
    return resolveFieldNameByFilter(filter, this.dataType);
  }
}
