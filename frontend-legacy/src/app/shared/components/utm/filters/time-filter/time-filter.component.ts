import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import * as moment from 'moment';
import {ModalService} from '../../../../../core/modal/modal.service';
import {DEMO_URL} from '../../../../constants/global.constant';
import {TimeFilterType} from '../../../../types/time-filter.type';
import {resolveFilterName, resolveRangeByTime} from '../../../../util/resolve-date';
import {DateRangeComponent} from '../date-range/date-range.component';


@Component({
  selector: 'app-filter-time',
  templateUrl: './time-filter.component.html',
  styleUrls: ['./time-filter.component.scss'],
})
export class TimeFilterComponent implements OnInit {
  @Input() timeDefault: TimeFilterType;
  @Input() hasAll: boolean;
  @Input() emitAtStart: boolean;
  time: string;
  label: string;
  dateRange: { timeFrom: Date, timeTo: Date };
  @Output() timeFilterChange: EventEmitter<TimeFilterType> = new EventEmitter();

  constructor(private modalService: ModalService) {
  }

  ngOnInit() {
    if (this.emitAtStart && this.timeDefault) {
      this.label = resolveFilterName(this.timeDefault.range);
      this.time = this.timeDefault.range;
      this.applyFilter();
    } else if (this.emitAtStart && !this.timeDefault) {
      this.label = resolveFilterName('day');
      this.time = 'day';
      this.applyFilter();
    } else if (this.emitAtStart && this.hasAll) {
      this.time = 'all';
      this.label = resolveFilterName(this.time);
      this.applyFilter();
    } else if (!this.emitAtStart && this.timeDefault) {
      this.label = resolveFilterName(this.timeDefault.range);
      this.time = this.timeDefault.range;
    } else if (this.hasAll) {
      this.time = 'all';
      this.label = resolveFilterName(this.time);
    } else if (!this.timeDefault && !this.emitAtStart) {
      this.label = resolveFilterName('week');
      this.time = 'week';
      this.applyFilter();
    }
  }

  applyFilter() {
    this.label = resolveFilterName(this.time);
    // avoid filter when is in demo
    if (!window.location.href.includes(DEMO_URL)) {
      this.timeFilterChange.emit(resolveRangeByTime(this.time));
    } else {
      this.timeFilterChange.emit(
        {
          timeFrom: '2019-08-01T00:00:00.000Z',
          timeTo: '2025-12-10T23:59:59.999Z',
          range: this.label
        });
    }
  }

  openCustomRangeModal() {
    const modalRef = this.modalService.open(DateRangeComponent, undefined, 'modal-date-range');
    modalRef.componentInstance.dateRange.subscribe((range) => {
      this.dateRange = range;
      this.processRange();
    });
  }

  processRange() {
    const to = this.dateRange.timeTo;
    const from = this.dateRange.timeFrom;
    this.label = moment(from).format('DD-MM-YYYY') + ' to ' + moment(to).format('DD-MM-YYYY');
    // avoid filter when is in demo
    if (!window.location.href.includes(DEMO_URL)) {
      this.timeFilterChange.emit(
        {
          timeFrom: moment(from).format('YYYY-MM-DD') + 'T00:00:00.000Z',
          timeTo: moment(to).format('YYYY-MM-DD') + 'T23:59:59.999Z',
          range: this.label
        });
    } else {
      this.timeFilterChange.emit(
        {
          timeFrom: '2019-08-01T00:00:00.000Z',
          timeTo: '2020-12-10T23:59:59.999Z',
          range: this.label
        });
    }
  }
}

