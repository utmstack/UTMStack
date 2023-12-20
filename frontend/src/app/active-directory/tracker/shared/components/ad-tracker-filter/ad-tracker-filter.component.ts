import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-ad-tracker-filter',
  templateUrl: './ad-tracker-filter.component.html',
  styleUrls: ['./ad-tracker-filter.component.scss']
})
export class AdTrackerFilterComponent implements OnInit {
  @Output() trackerFilterChange = new EventEmitter<any>();
  defaultTimeEvent: TimeFilterType = {range: 'all', timeFrom: null, timeTo: null};
  filter = {};
  objectTypes = ['USER', 'GROUP', 'COMPUTER', 'OBJECT'];
  object;

  constructor(private route: ActivatedRoute) {
  }

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      if (params.rangeDate) {
        this.defaultTimeEvent.timeTo = params.lastEventTo;
        this.defaultTimeEvent.timeFrom = params.lastEventFrom;
        this.defaultTimeEvent.range = params.rangeDate;
        this.searchByLastEvent(this.defaultTimeEvent);
      }
    });
  }

  searchByText() {
    this.trackerFilterChange.emit(this.filter);
  }

  searchByCreation($event: TimeFilterType) {
    this.filter['whenTracked.greaterThanOrEqual'] = $event.timeFrom;
    this.filter['whenTracked.lessThanOrEqual'] = $event.timeTo;
    this.trackerFilterChange.emit(this.filter);
  }

  searchByLastEvent($event: TimeFilterType) {
    this.filter['lastEventDate.greaterThanOrEqual'] = $event.timeFrom;
    this.filter['lastEventDate.lessThanOrEqual'] = $event.timeTo;
    this.trackerFilterChange.emit(this.filter);
  }
}
