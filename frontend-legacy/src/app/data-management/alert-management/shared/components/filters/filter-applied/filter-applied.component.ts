import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {
  ALERT_FILTERS_FIELDS,
  ALERT_GLOBAL_FIELD,
  ALERT_INCIDENT_FLAG_FIELD,
  ALERT_STATUS_FIELD,
  ALERT_TAGS_FIELD,
  ALERT_TIMESTAMP_FIELD,
  EVENT_IS_ALERT
} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {AlertFiltersBehavior} from '../../../behavior/alert-filters.behavior';
import {EventDataTypeEnum} from '../../../enums/event-data-type.enum';
import {AlertFilterManagementService} from '../../../services/alert-filter-management.service';
import {resolveFieldNameByFilter} from '../../../util/alert-util-function';

@Component({
  selector: 'app-filter-applied',
  templateUrl: './filter-applied.component.html',
  styleUrls: ['./filter-applied.component.scss']
})
export class FilterAppliedComponent implements OnInit {
  @Input() dataType: EventDataTypeEnum;
  filtersActive: ElasticFilterType[] = [];
  operatorEnum = ElasticOperatorsEnum;
  filters: ElasticFilterType[] = [];
  @Output() filterAppliedChange = new EventEmitter<{ filter: ElasticFilterType, valueDelete: string }>();

  constructor(private alertFiltersBehavior: AlertFiltersBehavior,
              private alertFilterManagementService: AlertFilterManagementService) {
  }

  ngOnInit() {
    this.alertFiltersBehavior.$filters.subscribe(filters => {
      if (filters) {
        this.filtersActive = filters.filter(value => value.field !== ALERT_STATUS_FIELD
          && value.field !== ALERT_TIMESTAMP_FIELD
          && value.field !== EVENT_IS_ALERT
          && value.field !== ALERT_INCIDENT_FLAG_FIELD
          && value.field !== ALERT_GLOBAL_FIELD &&
          value.operator === ElasticOperatorsEnum.IS_ONE_OF);
        /**
         * If field is not visible then set visible filter
         */
        for (const filter of this.filtersActive) {
          const indexFilter = ALERT_FILTERS_FIELDS.findIndex(value => value.field === filter.field);
          if (indexFilter !== -1) {
            if (!ALERT_FILTERS_FIELDS[indexFilter].visible) {
              ALERT_FILTERS_FIELDS[indexFilter].visible = true;
            }
          }
        }
      }
    });
  }

  showTagFilter() {
    return this.dataType !== EventDataTypeEnum.FALSE_POSITIVE &&
      this.filters.findIndex(value => value.field === ALERT_TAGS_FIELD
        && value.operator === ElasticOperatorsEnum.IS_NOT) !== -1;
  }


  resolveFilterName(filter: ElasticFilterType): string {
    return resolveFieldNameByFilter(filter, this.dataType);
  }

  onDelete(filter: ElasticFilterType, value?: string): Promise<ElasticFilterType> {
    return new Promise<ElasticFilterType>(resolve => {
      if (value) {
        const indexVal = filter.value.findIndex(val => val === value);
        filter.value.splice(indexVal, 1);
        if (filter.value.length === 0) {
          const activeFilterIndex = this.filtersActive.findIndex(fil => fil.field === filter.field);
          if (activeFilterIndex !== -1) {
            this.filtersActive.splice(activeFilterIndex, 1);
          }
        }
      }
      resolve(filter);
    });
  }

  deleteFilter(filter: ElasticFilterType, value?: string) {
    this.onDelete(filter, value).then(fil => {
      this.filterAppliedChange.emit({filter: fil, valueDelete: value});
    });
  }
}
