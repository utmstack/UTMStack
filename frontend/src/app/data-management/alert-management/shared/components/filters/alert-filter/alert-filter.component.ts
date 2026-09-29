import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {
  ADVERSARY_FILTERS_FIELDS,
  ALERT_FILTERS_FIELDS,
  EVENT_FILTERS_FIELDS,
  INCIDENT_FILTERS_FIELDS
} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {NatureDataPrefixEnum} from '../../../../../../shared/enums/nature-data.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {AlertDataTypeBehavior} from '../../../behavior/alert-data-type.behavior';
import {EventDataTypeEnum} from '../../../enums/event-data-type.enum';

@Component({
  selector: 'app-alert-filter',
  templateUrl: './alert-filter.component.html',
  styleUrls: ['./alert-filter.component.scss']
})
export class AlertFilterComponent implements OnInit, OnDestroy {

  fieldFilters: UtmFieldType[];
  alertPrefix = NatureDataPrefixEnum.ALERT + '*';
  @Input() dataType: EventDataTypeEnum;
  @Output() filterChange = new EventEmitter<ElasticFilterType>();
  @Output() filterReset = new EventEmitter<boolean>();
  canClickFor = 5000;
  isResettingFilters = false;
  destroy$ = new Subject<void>();

  filterFields: Record<EventDataTypeEnum, UtmFieldType[]> = {
    [EventDataTypeEnum.INCIDENT]: INCIDENT_FILTERS_FIELDS,
    [EventDataTypeEnum.EVENT]: EVENT_FILTERS_FIELDS,
    [EventDataTypeEnum.ALERT]: ALERT_FILTERS_FIELDS,
    [EventDataTypeEnum.FALSE_POSITIVE]: ALERT_FILTERS_FIELDS,
    [EventDataTypeEnum.ADVERSARY]: ADVERSARY_FILTERS_FIELDS
  };


  constructor(private alertDataTypeBehavior: AlertDataTypeBehavior) {
  }

  ngOnInit() {
    this.alertDataTypeBehavior.$alertDataType
      .pipe(takeUntil(this.destroy$))
      .subscribe(dataType => {
      if (dataType) {
        this.fieldFilters = this.resolveFiltersField(dataType);
      }
    });
  }

  resolveFiltersField(dataType: EventDataTypeEnum): UtmFieldType[] {
    return this.filterFields[dataType];
  }

  onFilterGenericChange($event: ElasticFilterType) {
    this.filterChange.emit($event);
  }

  onAlertSearch($event: string) {
    const filter: ElasticFilterType = {
      field: this.alertPrefix,
      operator: ElasticOperatorsEnum.IS_IN_FIELD,
      value: $event
    };
    this.filterChange.emit(filter);
  }


  resetAllFilters() {
    this.isResettingFilters = true;
    this.filterReset.emit(true);
    setTimeout(() => {
      this.isResettingFilters = false;
    }, this.canClickFor);
  }

  ngOnDestroy(): void {
   this.destroy$.next();
   this.destroy$.complete();
  }
}
