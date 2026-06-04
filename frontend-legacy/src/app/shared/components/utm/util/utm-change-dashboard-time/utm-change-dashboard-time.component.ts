import {Component, OnInit} from '@angular/core';
import {TimeFilterBehavior} from '../../../../behaviors/time-filter.behavior';
import {TimeFilterType} from '../../../../types/time-filter.type';

@Component({
  selector: 'app-utm-change-dashboard-time',
  templateUrl: './utm-change-dashboard-time.component.html',
  styleUrls: ['./utm-change-dashboard-time.component.scss']
})
export class UtmChangeDashboardTimeComponent implements OnInit {

  constructor(private timeFilterBehavior: TimeFilterBehavior) {
  }

  ngOnInit() {
  }

  onTimeFilterChange($event: TimeFilterType) {

  }
}
