import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {buildMultilineObject} from '../../../../shared/chart/util/build-multiline-option.util';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-ad-event-in-time',
  templateUrl: './chart-ad-event-in-time.component.html',
  styleUrls: ['./chart-ad-event-in-time.component.scss']
})
export class ChartAdEventInTimeComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  loadingAmountOfObjectsInTime = true;
  eventByObjectOption: any;
  echartEnum = ChartTypeEnum;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  time: TimeFilterType;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadAmountOfObjectsInTime({from: this.time.timeFrom, to: this.time.timeTo});
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  onEventByObjectDateChange($event: TimeFilterType) {
    this.time = $event;
    this.loadAmountOfObjectsInTime({from: $event.timeFrom, to: $event.timeTo});
  }

  loadAmountOfObjectsInTime(req) {
    this.adDashboardService.getAmountOfObjectsInTime(req).subscribe(data => {
      this.loadingAmountOfObjectsInTime = false;
      if (data.body !== null) {
        buildMultilineObject(data.body).then(multiline => {
          this.eventByObjectOption = multiline;
          this.loadingAmountOfObjectsInTime = false;
        });
      } else {
        this.eventByObjectOption = null;
      }
    });
  }

  filterByTime($event: any) {
  }


}
