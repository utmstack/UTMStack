import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {SerieValue} from '../../../../shared/chart/types/response/multiline-response.type';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-ad-most-active-user',
  templateUrl: './chart-ad-most-active-user.component.html',
  styleUrls: ['./chart-ad-most-active-user.component.scss']
})
export class ChartAdMostActiveUserComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  @Input() top;
  interval: any;
  echartEnum = ChartTypeEnum;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  time: TimeFilterType;
  loadingMostActiveUser: boolean;
  topMostActive: SerieValue[] = [];

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadMostActiveUser({from: this.time.timeFrom, to: this.time.timeTo, top: this.top});
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  onMostActiveUserDateChange($event: TimeFilterType) {
    this.time = $event;
    const req = {
      from: $event.timeFrom,
      to: $event.timeTo,
      top: this.top,
    };
    this.loadMostActiveUser(req);
  }

  loadMostActiveUser(req) {
    this.adDashboardService.getTopMostActiveUsers(req).subscribe(mostActive => {
      this.loadingMostActiveUser = false;
      this.topMostActive = mostActive.body;
    });
  }

  navigateToUserMostActiveDetails(serie: string) {
    this.router.navigate(['/active-directory/detail/users'], {
      queryParams: {
        userChange: serie.toLowerCase()
      }
    });
  }
}
