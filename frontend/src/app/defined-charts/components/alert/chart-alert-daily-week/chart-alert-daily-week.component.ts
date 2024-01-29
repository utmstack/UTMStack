import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ALERT_GLOBAL_FIELD, ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {OverviewAlertDashboardService} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {ChartSerieValueType} from '../../../../shared/types/chart-reponse/chart-serie-value.type';

@Component({
  selector: 'app-chart-alert-daily-week',
  templateUrl: './chart-alert-daily-week.component.html',
  styleUrls: ['./chart-alert-daily-week.component.scss']
})
export class ChartAlertDailyWeekComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  dailyAlert: ChartSerieValueType[] = [];
  loadingChartDailyAlert = true;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.getDailyAlert();
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getDailyAlert();
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  getDailyAlert() {
    this.overviewAlertDashboardService.getCardAlertTodayWeek().subscribe(response => {
      this.dailyAlert = response.body;
      this.loadingChartDailyAlert = false;
    });
  }

  chartEvent(type: 'today' | 'week') {
    const queryParams = {};
    queryParams[ALERT_GLOBAL_FIELD] = 'ALERT';
    if (type !== 'today') {
      queryParams[ALERT_TIMESTAMP_FIELD] = ElasticOperatorsEnum.IS_BETWEEN + '->now-7d/d,now';
    } else {
      queryParams[ALERT_TIMESTAMP_FIELD] = ElasticOperatorsEnum.IS_BETWEEN + '->now/d,now';
    }
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/data/alert/view'], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

}
