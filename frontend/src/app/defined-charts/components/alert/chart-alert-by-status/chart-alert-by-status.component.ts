import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ALERT_GLOBAL_FIELD, ALERT_STATUS_FIELD, ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {
  CLOSED,
  CLOSED_ICON,
  IGNORED,
  IGNORED_ICON,
  OPEN,
  OPEN_ICON,
  REVIEW,
  REVIEW_ICON
} from '../../../../shared/constants/alert/alert-status.constant';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {OverviewAlertDashboardService} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {ChartSerieValueType} from '../../../../shared/types/chart-reponse/chart-serie-value.type';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-alert-by-status',
  templateUrl: './chart-alert-by-status.component.html',
  styleUrls: ['./chart-alert-by-status.component.scss']
})
export class ChartAlertByStatusComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  time: TimeFilterType;
  status: ChartSerieValueType[];
  loadingStatusAlert = true;
  //
  OPEN = OPEN;
  REVIEW = REVIEW;
  IGNORED = IGNORED;
  CLOSED = CLOSED;
  OPEN_ICON = OPEN_ICON;
  REVIEW_ICON = REVIEW_ICON;
  IGNORED_ICON = IGNORED_ICON;
  CLOSED_ICON = CLOSED_ICON;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getAlertByStatus(this.time);
      }, this.refreshInterval);
    }
  }

  onChangeAlertByStatus($event: TimeFilterType) {
    this.time = $event;
    this.getAlertByStatus($event);
  }

  getAlertByStatus(time: TimeFilterType) {
    const req = {
      to: time.timeTo,
      from: time.timeFrom
    };
    this.overviewAlertDashboardService.getCardAlertByStatus(req).subscribe(res => {
      this.status = res.body;
      this.loadingStatusAlert = false;
    });
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  chartEvent(status: number) {
    const queryParams = {};
    queryParams[ALERT_GLOBAL_FIELD] = 'ALERT';
    queryParams[ALERT_STATUS_FIELD] = ElasticOperatorsEnum.IS + '->' + status;
    queryParams[ALERT_TIMESTAMP_FIELD] = ElasticOperatorsEnum.IS_BETWEEN + '->'
      + this.time.timeFrom + ',' + this.time.timeTo;
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/data/alert/view'], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }
}
