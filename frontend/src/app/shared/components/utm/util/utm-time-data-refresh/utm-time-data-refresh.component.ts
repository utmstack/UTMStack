import {Component, OnInit} from '@angular/core';
import {TIME_DASHBOARD_REFRESH} from '../../../../constants/time-refresh.const';

@Component({
  selector: 'app-utm-time-data-refresh',
  templateUrl: './utm-time-data-refresh.component.html',
  styleUrls: ['./utm-time-data-refresh.component.scss']
})
export class UtmTimeDataRefreshComponent implements OnInit {
  times = TIME_DASHBOARD_REFRESH;
  timeRefresh: { time: string, milliseconds: number } = this.times[0];

  constructor() {
  }

  ngOnInit() {
  }

}
