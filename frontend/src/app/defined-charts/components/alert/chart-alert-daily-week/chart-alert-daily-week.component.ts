import {ChangeDetectionStrategy, ChangeDetectorRef, Component, EventEmitter, OnDestroy, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, filter, map, startWith, switchMap, takeUntil, tap} from 'rxjs/operators';
import {ALERT_GLOBAL_FIELD, ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {
  OverviewAlertDashboardService
} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {RefreshService, RefreshType} from '../../../../shared/services/util/refresh.service';
import {ChartSerieValueType} from '../../../../shared/types/chart-reponse/chart-serie-value.type';
import {TimeFilterBehavior} from "../../../../shared/behaviors/time-filter.behavior";

@Component({
    standalone: false,
    selector: 'app-chart-alert-daily-week',
    templateUrl: './chart-alert-daily-week.component.html',
    styleUrls: ['./chart-alert-daily-week.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ChartAlertDailyWeekComponent implements OnInit, OnDestroy {
  @Output() loaded = new EventEmitter<void>();
  interval: any;
  dailyAlert: ChartSerieValueType[] = [];
  loadingChartDailyAlert = true;
  destroy$ = new Subject<void>();
  dailyAlert$: Observable<ChartSerieValueType[]>;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private refreshService: RefreshService,
              private router: Router,
              private spinner: NgxSpinnerService,
              private cdr: ChangeDetectorRef) {
  }

  ngOnInit() {

    this.dailyAlert$ = this.refreshService.refresh$
      .pipe(
        takeUntil(this.destroy$),
        filter((refreshType: string) => (
          refreshType === RefreshType.ALL)),
        startWith(RefreshType.ALL),
        switchMap(() => {
          return this.getDailyAlert();
        })
      );
  }

  init() {
    return this.getDailyAlert();
  }

  getDailyAlert() {
    this.loadingChartDailyAlert = true;
    this.cdr.markForCheck();
    return this.overviewAlertDashboardService
      .getCardAlertTodayWeek()
        .pipe(
          map(response => {
            return response.body;
          }),
          tap(() => {
            this.loaded.emit();
            this.loadingChartDailyAlert = false;
            this.cdr.markForCheck();
          }),
          catchError(() => {
            this.loadingChartDailyAlert = false;
            this.cdr.markForCheck();
            return EMPTY;
          }));
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

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

}
