import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Subject} from 'rxjs';
import {ALERT_STATUS_FIELD} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ALL_STATUS, AUTOMATIC_REVIEW, CLOSED, IGNORED, OPEN, REVIEW} from '../../../../../../shared/constants/alert/alert-status.constant';
import {ALERT_INDEX_PATTERN} from '../../../../../../shared/constants/main-index-pattern.constant';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../../../../../shared/enums/nature-data.enum';
import {ElasticSearchIndexService} from '../../../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {AlertFiltersBehavior} from '../../../behavior/alert-filters.behavior';
import {AlertStatusBehavior} from '../../../behavior/alert-status.behavior';
import {EventDataTypeEnum} from '../../../enums/event-data-type.enum';
import {getCurrentAlertStatus} from '../../../util/alert-util-function';
import {takeUntil} from "rxjs/operators";

@Component({
  selector: 'app-status-filter',
  templateUrl: './status-filter.component.html',
  styleUrls: ['./status-filter.component.scss']
})
export class StatusFilterComponent implements OnInit, OnDestroy {
  @Input() dataNature: DataNatureTypeEnum;
  @Input() filters: ElasticFilterType[];
  @Input() dataType: EventDataTypeEnum;
  @Output() filterStatusChange = new EventEmitter<ElasticFilterType>();
  @Input() statusFilter: number = ALL_STATUS;
  eventDataTypeEnum = EventDataTypeEnum;
  statusValue: object = {
    2: 0,
    3: 0,
    5: 0
  };
  statusValueArray: { key: number, value: number }[] = [
    {key: 2, value: 0},
    {key: 3, value: 0},
    {key: 5, value: 0}
  ];
  allStatusValue = ALL_STATUS;
  autoReview = AUTOMATIC_REVIEW;
  destroy$: Subject<void> = new Subject();

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private alertFiltersBehavior: AlertFiltersBehavior,
              private updateStatusServiceBehavior: AlertStatusBehavior) {
  }

  ngOnInit() {
    if (typeof this.statusFilter === 'string') {
      this.statusFilter = Number(this.statusFilter);
    }
    this.getValuesOfStatus();
    /**
     * Update amount on alert status change
     */
    this.updateStatusServiceBehavior.$updateStatus
      .pipe(takeUntil(this.destroy$))
      .subscribe((update) => {
      if (update) {
        this.getValuesOfStatus();
      }
    });
    /**
     * Update amount on filter change
     */
    this.alertFiltersBehavior.$filters
      .pipe(takeUntil(this.destroy$))
      .subscribe(value => {
      if (value) {
        this.filters = value;
        this.statusFilter = getCurrentAlertStatus(this.filters);
        this.getValuesOfStatus();
      }
    });
  }

  getValuesOfStatus() {
    const req = {
      field: ALERT_STATUS_FIELD,
      filters: this.filters.filter(value => value.field !== ALERT_STATUS_FIELD),
      index: ALERT_INDEX_PATTERN,
      top: 10
    };
    this.elasticSearchIndexService.getValuesWithCount(req).subscribe(value => {
      this.reInitArrayValues().then(() => {
        Object.keys(value.body).forEach(keys => {
          const statusNum = Number(keys);
          if (statusNum !== 1) {
            const staIndex = this.statusValueArray.findIndex(arr => arr.key === statusNum);
            this.statusValueArray[staIndex].value = Number(value.body[keys]);
          }
        });
      });
    });
  }

  filterByStatus(status: { key: any, value: any }) {
    this.statusFilter = status.key;
    // If all is selected active filter is get all distinct review status
    const filter = {
      field: ALERT_STATUS_FIELD,
      operator: status.key === ALL_STATUS ? ElasticOperatorsEnum.IS_NOT : ElasticOperatorsEnum.IS,
      value: status.key === ALL_STATUS ? AUTOMATIC_REVIEW : status.key
    };
    this.filterStatusChange.emit(filter);
  }

  reInitArrayValues(): Promise<void> {
    return new Promise<void>(resolve => {
      this.statusValueArray = [
        {key: 2, value: 0},
        {key: 3, value: 0},
         {key: 5, value: 0}
      ];
      resolve();
    });
  }

  resolveStatusLabel(key) {
    const status = Number(key);
    switch (status) {
      case OPEN:
        return 'alertStatus.open';
      case REVIEW:
        return 'alertStatus.inReview';
      case CLOSED:
        return 'alertStatus.closed';
      case ALL_STATUS:
        return 'alertStatus.all';
      default:
        return 'alertStatus.pending';
    }
  }

  getTotalStatusTotal(): number {
    let total = 0;
    this.statusValueArray.forEach(value => total += value.value);
    return total;
  }

  showStatus(key: number) {
    if (this.dataType === this.eventDataTypeEnum.INCIDENT) {
      return key !== IGNORED;
    } else {
      return true;
    }
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
