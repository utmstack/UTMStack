import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-schedule-filter',
  templateUrl: './schedule-filter.component.html',
  styleUrls: ['./schedule-filter.component.scss']
})
export class ScheduleFilterComponent implements OnInit {
  @Output() scheduleFilterChange = new EventEmitter<any>();
  filter = {};
  periodTimeFilterMin = ['hour', 'day', 'week', 'month'];
  periodTimeFilterMax = [];
  durationTimeFilterMin = ['hour', 'day', 'week', 'month'];
  durationTimeFilterMax = [];
  periodMin: string;
  periodMax: any;
  durationMin: any;
  durationMax: string;

  constructor() {
  }

  ngOnInit() {
  }

  searchByName() {
    this.scheduleFilterChange.emit(this.filter);
  }


  firstFilterTime($event: TimeFilterType) {
    this.filter['firstRun.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filter['firstRun.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.scheduleFilterChange.emit(this.filter);
  }

  nextFilterTime($event: TimeFilterType) {
    this.filter['nextRun.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filter['nextRun.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.scheduleFilterChange.emit(this.filter);
  }

  searchByPeriod() {
    this.scheduleFilterChange.emit(this.filter);
  }

  searchDuration() {
  }

  searchByDuration() {
  }

  setMaxPeriodTime(time: string) {
    this.periodTimeFilterMax = this.resolveMaxArray(time);
  }

  setMaxDurationTime(time: string) {
    this.durationTimeFilterMax = this.resolveMaxArray(time);
  }

  resolveMaxArray(time: string): string[] {
    if (time === 'hour') {
      return ['hour', 'day', 'week', 'month'];
    } else if (time === 'day') {
      return ['day', 'week', 'month'];
    } else if (time === 'week') {
      return ['week', 'month'];
    } else if (time === 'month') {
      return ['month'];
    }
  }
}
