import {
  Component,
  EventEmitter,
  Input,
  OnChanges,
  OnDestroy,
  OnInit,
  Output,
  SimpleChanges,
  ViewChild
} from '@angular/core';
import {NgbActiveModal, NgbDate, NgbDateStruct, NgbPopover, NgbTimeStruct} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {TimeFilterBehavior} from '../../../../behaviors/time-filter.behavior';
import {ElasticTimeEnum} from '../../../../enums/elastic-time.enum';
import {ElasticFilterCommonType} from '../../../../types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../types/time-filter.type';
import {ngbDateToDate} from '../../../../util/date.util';
import {setMaxDateToday} from '../../../../util/datepicker-util';
import {buildFormatInstantFromDate, resolveInstantDate, resolveUTCDate} from '../../../../util/utm-time.util';

@Component({
  selector: 'app-elastic-filter-time',
  templateUrl: './elastic-filter-time.component.html',
  styleUrls: ['./elastic-filter-time.component.scss']
})
export class ElasticFilterTimeComponent implements OnInit, OnChanges, OnDestroy {
  @Input() invertContent: boolean;
  @Input() timeDefault: ElasticFilterCommonType;
  @Input() template: 'default' | 'log-explorer' = 'default';
  /**
   * Use this property to set default range, receive object type ElasticFilterDefaultTime
   */
  @Input() defaultTime: ElasticFilterDefaultTime;
  @Input() formatInstant: boolean;
  @Input() container: string;
  @Input() changeOnInit: 'YES' | 'NO';
  /**
   * Use this attribute to control multiple emmit, need to avoid recursive calling
   */
  @Input() isEmitter: boolean;
  @Output() timeFilterChange = new EventEmitter<TimeFilterType>();

  @ViewChild('popover') popover!: NgbPopover;


  times: { time: ElasticTimeEnum, label: string } [] = [
/*    {time: ElasticTimeEnum.YEAR, label: 'year'},*/
    {time: ElasticTimeEnum.MONTH, label: 'month'},
    {time: ElasticTimeEnum.WEEKS, label: 'weeks'},
    {time: ElasticTimeEnum.DAY, label: 'day'},
    {time: ElasticTimeEnum.HOUR, label: 'hour'},
    {time: ElasticTimeEnum.MINUTE, label: 'minute'},
    {time: ElasticTimeEnum.SECOND, label: 'second'}
  ];
  selected = '';
  lastTime: number;
  commonlyUsed: ElasticFilterCommonType[] = [
    {time: ElasticTimeEnum.MINUTE, last: 15, label: 'last 15 minutes'},
    {time: ElasticTimeEnum.MINUTE, last: 30, label: 'last 30 minutes'},
    {time: ElasticTimeEnum.HOUR, last: 1, label: 'last hour'},
    {time: ElasticTimeEnum.HOUR, last: 6, label: 'last 6 hours'},
    {time: ElasticTimeEnum.HOUR, last: 12, label: 'last 12 hours'},
    {time: ElasticTimeEnum.HOUR, last: 24, label: 'last 24 hours'},
    {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'},
   {time: ElasticTimeEnum.DAY, last: 30, label: 'last 30 days'},
    /* {time: ElasticTimeEnum.DAY, last: 90, label: 'last 90 days'},
    {time: ElasticTimeEnum.YEAR, last: 1, label: 'last year'},*/
  ];
  timeUnit: { time: ElasticTimeEnum, label: string } = this.times[0];
  dateFrom: string;
  dateTo: string = ElasticTimeEnum.NOW;
  rangeTimeTo: NgbDate;
  timeFrom: NgbTimeStruct = {hour: 0, minute: 0, second: 0};
  timeTo: NgbTimeStruct = {hour: 23, minute: 59, second: 59};
  rangeTimeFrom: NgbDate;
  maxDate = setMaxDateToday();
  public isCollapsed = false;
  isCollapsedCommon = true;
  maxDateFrom: NgbDateStruct;
  maxDateTo: NgbDateStruct;
  destroy$: Subject<void> = new Subject();

  constructor(public activeModal: NgbActiveModal,
              private timeFilterBehavior: TimeFilterBehavior) {
    /*this.maxDateFrom = this.maxDate;
    this.maxDateTo = this.maxDate;*/
  }


  ngOnChanges(changes: SimpleChanges): void {
    // TODO
    // if (changes && changes.defaultTime && changes.defaultTime.currentValue === null) {
    //   this.dateFrom = null;
    //   this.dateTo = null;
    // }
  }

  ngOnDestroy(): void {
    this.timeFilterBehavior.$time.next(null);
    this.destroy$.next();
    this.destroy$.complete();
  }

  ngOnInit() {
    this.timeFilterBehavior.$time
      .pipe(takeUntil(this.destroy$))
      .subscribe(time => {
      if (time && !this.isEmitter) {
        this.dateTo = time.to;
        this.dateFrom = time.from;
        if (time.update) {
          if (!this.formatInstant) {
            this.timeFilterChange.emit({timeFrom: time.from, timeTo: time.to});
          } else {
            this.timeFilterChange.emit(buildFormatInstantFromDate(time));
          }
        }
      }
    });

    if (this.changeOnInit !== 'NO') {
      if (this.timeDefault) {
        this.applyCommonDate(this.timeDefault);
      }
      if (this.defaultTime) {
        this.applyDefaultTime();
      }
    } else {
      if (this.defaultTime) {
        this.dateFrom = this.defaultTime.from;
        this.dateTo = this.defaultTime.to;
      }
    }
  }


  applyDate() {
    this.dateTo = 'now';
    this.dateFrom = 'last ' + this.lastTime + ' ' + this.timeUnit.label;
    if (!this.formatInstant) {
      const timeFrom = ElasticTimeEnum.NOW + '-' + this.lastTime + this.timeUnit.time;
      const timeTo = ElasticTimeEnum.NOW;
      this.emitElasticDate(timeFrom, timeTo);
    } else {
      this.emitFormatInstant(this.timeUnit.time, this.lastTime);
    }
  }


  applyCommonDate(common: { time: ElasticTimeEnum; last: number; label: string }) {
    this.dateTo = 'now';
    this.dateFrom = common.label;
    if (!this.formatInstant) {
      const timeFrom = ElasticTimeEnum.NOW + '-' + common.last + common.time;
      const timeTo = ElasticTimeEnum.NOW;
      this.emitElasticDate(timeFrom, timeTo);
    } else {
      this.emitFormatInstant(common.time, common.last);
    }
  }

  emitElasticDate(from: any, to: any) {
    if (this.isEmitter) {
      this.timeFilterBehavior.$time.next({from, to});
    } else {
      this.timeFilterChange.emit({timeFrom: from, timeTo: to});
    }
  }

  emitFormatInstant(unitTime, time) {
    const instant = resolveInstantDate(unitTime, time);
    if (this.isEmitter) {
      this.timeFilterBehavior.$time.next({from: instant.timeFrom, to: instant.timeTo});
    } else {
      this.timeFilterChange.emit({timeFrom: instant.timeFrom, timeTo: instant.timeTo});
    }
  }

  /*isValidDate() {
    if (this.rangeTimeFrom && this.rangeTimeTo) {
      const from = Number(new Date(this.extractDate(this.rangeTimeFrom, this.timeFrom)).getTime());
      const to = Number(new Date(this.extractDate(this.rangeTimeTo, this.timeTo)).getTime());
      return to - from >= 0;
    } else {
      return false;
    }
  }*/

  applyRange() {
    if (this.isValidDate()) {
      this.dateTo = this.extractDate(this.rangeTimeTo, this.timeTo);
      this.dateFrom = this.extractDate(this.rangeTimeFrom, this.timeFrom);
      if (this.isEmitter) {
        this.timeFilterBehavior.$time.next({
          from: this.extractDate(this.rangeTimeFrom, this.timeFrom),
          to: this.extractDate(this.rangeTimeTo, this.timeTo)
        });
      } else {
        this.timeFilterChange.emit({
          timeFrom: resolveUTCDate(this.dateFrom), timeTo: resolveUTCDate(this.dateTo)
        });
      }
    }
  }

  extractDate(date: NgbDate, time: NgbTimeStruct): string {
    return ngbDateToDate(date, time);
  }

  private applyDefaultTime() {
    this.dateFrom = this.defaultTime.from;
    this.dateTo = this.defaultTime.to;
    if (!this.formatInstant) {
      this.timeFilterChange.emit(
        {
          timeFrom: this.dateFrom,
          timeTo: this.dateTo
        });
    } else {
      // dateFrom always is now-amount+unit time,
      const time = this.dateFrom.match(/[\d\.]+|\D+/g);
      this.emitFormatInstant(time[2], Number(time[1]));
    }
  }

  // Function called every time the 'timeFrom' date is changed
  onTimeFromChange() {
    this.updateMaxDates('from');  // Update maxDates when timeFrom changes
  }

  // Function called every time the 'timeTo' date is changed
  onTimeToChange() {
    this.updateMaxDates('to');  // Update maxDates when timeTo changes
  }

  // Update the maxDate values based on selected 'timeFrom' and 'timeTo' dates
  updateMaxDates(type: 'from' | 'to') {
    if (this.rangeTimeFrom && type === 'from') {
      const maxDateTo = new Date(this.rangeTimeFrom.year, this.rangeTimeFrom.month - 1, this.rangeTimeFrom.day);
      maxDateTo.setDate(maxDateTo.getDate() + 30);  // Set maxDateTo to 30 days after timeFrom
      this.maxDateTo = {
        year: maxDateTo.getFullYear(),
        month: maxDateTo.getMonth() + 1,
        day: maxDateTo.getDate()
      };
    }

    if (this.rangeTimeTo && type === 'to') {
      const maxDateFrom = new Date(this.rangeTimeTo.year, this.rangeTimeTo.month - 1, this.rangeTimeTo.day);
      maxDateFrom.setDate(maxDateFrom.getDate() - 30);  // Set maxDateFrom to 30 days before timeTo
      this.maxDateFrom = {
        year: maxDateFrom.getFullYear(),
        month: maxDateFrom.getMonth() + 1,
        day: maxDateFrom.getDate()
      };
    }
  }

  isValidDate(): boolean {
    if (this.rangeTimeFrom && this.rangeTimeTo) {
      const from = new Date(this.rangeTimeFrom.year, this.rangeTimeFrom.month - 1, this.rangeTimeFrom.day);
      const to = new Date(this.rangeTimeTo.year, this.rangeTimeTo.month - 1, this.rangeTimeTo.day);
      const diffInTime = to.getTime() - from.getTime();
      const diffInDays = diffInTime / (1000 * 3600 * 24);
      return diffInDays >= 0 && diffInDays <= 30;
    }
    return false;
  }

  isDirty(){
    return this.rangeTimeFrom !== undefined && this.rangeTimeTo !== undefined;
  }

  maxTimeValue(): number {
    switch (this.timeUnit.time) {
      case ElasticTimeEnum.MONTH:
        return 1;
      case ElasticTimeEnum.WEEKS:
        return 4;
      case ElasticTimeEnum.DAY:
        return 30;
      case ElasticTimeEnum.HOUR:
        return 720;
      case ElasticTimeEnum.MINUTE:
        return 432000;
      case ElasticTimeEnum.SECOND:
        return 2592000;

      default:
        return 0;
    }
  }
}

export class ElasticFilterDefaultTime {
  from: string;
  to: string;

  constructor(from: string, to: string) {
    this.from = from;
    this.to = to;
  }
}


// "year" | "years" | "y" |
// "month" | "months" | "M" |
// "week" | "weeks" | "w" |
// "day" | "days" | "d" |
// "hour" | "hours" | "h" |
// "minute" | "minutes" | "m" |
// "second" | "seconds" | "s" |
