import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-ad-permissions',
  templateUrl: './chart-ad-permissions.component.html',
  styleUrls: ['./chart-ad-permissions.component.scss']
})
export class ChartAdPermissionsComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  amountOfUsersScaledPrivileges: number;
  amountChangesTrackedUsers: number;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  // defaultTime: TimeFilterType = resolveRangeByTime('week');
  dateQuickPermission: TimeFilterType;
  loading = true;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.onQuickPermissionDateChange(this.dateQuickPermission);
      }, this.refreshInterval);
    }
  }

  onQuickPermissionDateChange($event: TimeFilterType) {
    this.dateQuickPermission = $event;
    const req = {
      from: $event.timeFrom,
      to: $event.timeTo
    };
    this.adDashboardService.getAmountOfUsersScaledPrivileges(req).subscribe(amountScaled => {
      this.amountOfUsersScaledPrivileges = amountScaled.body;
      this.adDashboardService.getAmountOfTrackedUserWithChanges(req).subscribe(tracker => {
        this.amountChangesTrackedUsers = tracker.body;
        this.loading = false;
      });
    });
  }

  navigateToUsersPrivileges() {
    this.router.navigate(['/active-directory/detail/users'], {
      queryParams: {
        permissions: 'scaled',
        from: this.dateQuickPermission.timeFrom,
        to: this.dateQuickPermission.timeTo,
        rangeDate: this.dateQuickPermission.range
      }
    });
  }

  navigateToTrackers() {
    this.router.navigate(['/active-directory/tracker'], {
      queryParams: {
        lastEventFrom: this.dateQuickPermission.timeFrom,
        lastEventTo: this.dateQuickPermission.timeTo,
        rangeDate: this.dateQuickPermission.range
      }
    });
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }
}
