import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';

@Component({
  selector: 'app-chart-ad-quick-info',
  templateUrl: './chart-ad-quick-info.component.html',
  styleUrls: ['./chart-ad-quick-info.component.scss']
})
export class ChartAdQuickInfoComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  amountOfUsersDisabled: number;
  amountOfUsersLockout: number;
  loadingKpi = true;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.loadQuickInfoMetrics();
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadQuickInfoMetrics();
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  loadQuickInfoMetrics() {
    this.adDashboardService.getAmountOfUsersLockout().subscribe(amountLockout => {
      this.amountOfUsersLockout = amountLockout.body;
      this.adDashboardService.getAmountOfUsersDisabled().subscribe(amountDisable => {
        this.amountOfUsersDisabled = amountDisable.body;
        this.loadingKpi = false;
      });
    });
  }

  navigateToUserDetails(status: string) {
    const queryParams = {
      status
    };
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/active-directory/detail/users'], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

}
