import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, Subject} from 'rxjs';
import {filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {buildMultilineObject} from '../../../../shared/chart/util/build-multiline-option.util';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {IndexPatternSystemEnumID, IndexPatternSystemEnumName} from '../../../../shared/enums/index-pattern-system.enum';
import {OverviewAlertDashboardService} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {RefreshService, RefreshType} from '../../../../shared/services/util/refresh.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-event-in-time',
  templateUrl: './chart-event-in-time.component.html',
  styleUrls: ['./chart-event-in-time.component.scss']
})
export class ChartEventInTimeComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  @Output() loaded = new EventEmitter<void>();
  @Input() type: RefreshType;
  interval: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  queryParams = {from: 'now-7d', to: 'now', interval: 'Day'};
  loadingPieOption = true;
  chartEnumType = ChartTypeEnum;
  multilineOption: any;
  noData = false;
  destroy$ = new Subject<void>();
  data$: Observable<any>;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private refreshService: RefreshService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    /*if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getEventByTime();
      }, this.refreshInterval);
    }*/
    this.data$ = this.refreshService.refresh$
      .pipe(
        takeUntil(this.destroy$),
        filter(refreshType => (
          refreshType === RefreshType.ALL || refreshType === this.type)),
        switchMap(() => this.getEventByTime()));
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.queryParams.from = $event.timeFrom;
    this.queryParams.to = $event.timeTo;
    this.refreshService.sendRefresh(this.type);
  }

  onChartClick($event: any) {
    const from = $event.name.replace(' 00:00', 'T00:00:00.000Z');
    const to = $event.name.replace(' 00:00', 'T23:59:59.999Z');
    const params = {
      // 'global.type.keyword': 'logx',
      '@timestamp': ElasticOperatorsEnum.IS_BETWEEN + '->' + from + ',' + to,
      'dataType.keyword': $event.seriesName,
      patternId: IndexPatternSystemEnumID.LOG,
      indexPattern: IndexPatternSystemEnumName.LOG
    };

    this.spinner.show('loadingSpinner');
    this.router.navigate(['/discover/log-analyzer'], {
      queryParams: params
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
    //   global.type.keyword is logx
    //   logx.type.keyword exist
    //   logx.type.keyword is [type]
    // @timestamp is [date]
  }

  private getEventByTime() {
    /*this.overviewAlertDashboardService.getEventInTime(this.queryParams)
      .subscribe(event => {
        this.loadingPieOption = false;
        if (event.body.categories.length > 0) {
          this.noData = false;
          buildMultilineObject(event.body).then(option => {
            this.multilineOption = option;
          });
        } else {
          this.noData = true;
        }
      });*/

    return this.overviewAlertDashboardService.getEventInTime(this.queryParams)
      .pipe(
        map( response => response.body),
        tap(data => {
          this.loadingPieOption = false;
          if (data.categories.length > 0) {
            this.noData = false;
            buildMultilineObject(data).then(option => {
              this.multilineOption = option;
            });
          } else {
            this.noData = true;
          }
          this.loaded.emit();
        })
      );
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
