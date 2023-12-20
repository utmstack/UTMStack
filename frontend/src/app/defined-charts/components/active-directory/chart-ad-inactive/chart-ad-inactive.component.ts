import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {DEMO_URL} from '../../../../shared/constants/global.constant';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';

@Component({
  selector: 'app-chart-ad-inactive',
  templateUrl: './chart-ad-inactive.component.html',
  styleUrls: ['./chart-ad-inactive.component.scss']
})
export class ChartAdInactiveComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  daysInactive = 15;
  amountOfInactiveUsers: number;
  loadingKpi = true;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.loadInactiveLastDays();
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadInactiveLastDays();
      }, this.refreshInterval);
    }
  }

  loadInactiveLastDays() {
    if (this.daysInactive) {
      let req;
      if (!window.location.href.includes(DEMO_URL)) {
        req = {
          daysToBeInactiveUser: this.daysInactive
        };
      } else {
        req = {
          daysToBeInactiveUser: 1
        };
      }
      this.loadAmountOfInactiveUsers(req);
    }
  }

  loadAmountOfInactiveUsers(req?: any) {
    this.adDashboardService.getAmountOfInactiveUsers(req).subscribe(inactive => {
      this.amountOfInactiveUsers = inactive.body;
      this.loadingKpi = false;
    });
  }

  navigateToUserDetails(status: string) {
    const queryParams = {
      status,
      inactiveDays: this.daysInactive
    };
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/active-directory/detail/users'], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

}
