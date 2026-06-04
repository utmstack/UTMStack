import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {ActiveDirectoryType} from '../../../../active-directory/shared/types/active-directory.type';
import {DEMO_URL} from '../../../../shared/constants/global.constant';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {UtmDateFormatEnum} from '../../../../shared/enums/utm-date-format.enum';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-ad-inactive-admin',
  templateUrl: './chart-ad-inactive-admin.component.html',
  styleUrls: ['./chart-ad-inactive-admin.component.scss']
})
export class ChartAdInactiveAdminComponent implements OnInit, OnDestroy {
  daysInactiveAdmin = 15;
  loadingInactiveAdmins = true;
  @Input() refreshInterval;
  interval: any;
  echartEnum = ChartTypeEnum;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  time: TimeFilterType;
  inactiveAdmins: ActiveDirectoryType[] = [];
  formatDateEnum = UtmDateFormatEnum;
  viewEvent: ActiveDirectoryType;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadInactiveLastDaysAdmin();
      }, this.refreshInterval);
    }
    this.loadInactiveAdmin();
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }


  loadInactiveLastDaysAdmin() {
    if (this.daysInactiveAdmin) {
      setTimeout(() => this.loadInactiveAdmin(), 1500);
    }
  }

  loadInactiveAdmin() {
    if (this.daysInactiveAdmin) {
      let req;
      if (!window.location.href.includes(DEMO_URL)) {
        req = {
          daysToBeInactiveUser: this.daysInactiveAdmin
        };
      } else {
        req = {
          daysToBeInactiveUser: 1
        };
      }

      this.adDashboardService.getInactiveAdmins(req).subscribe(inactive => {
        this.inactiveAdmins = inactive.body ? inactive.body : [];
        this.loadingInactiveAdmins = false;
      });
    }
  }
}
