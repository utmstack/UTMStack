import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {buildMultilineObject} from '../../../../shared/chart/util/build-multiline-option.util';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {IndexPatternSystemEnumID, IndexPatternSystemEnumName} from '../../../../shared/enums/index-pattern-system.enum';
import {OverviewAlertDashboardService} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-event-in-time',
  templateUrl: './chart-event-in-time.component.html',
  styleUrls: ['./chart-event-in-time.component.scss']
})
export class ChartEventInTimeComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  queryParams = {from: 'now-7d', to: 'now', interval: 'Day'};
  loadingPieOption = true;
  chartEnumType = ChartTypeEnum;
  multilineOption: any;
  noData = false;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getEventByTime();
      }, this.refreshInterval);
    }
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.queryParams.from = $event.timeFrom;
    this.queryParams.to = $event.timeTo;
    this.getEventByTime();
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

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  private getEventByTime() {
    this.overviewAlertDashboardService.getEventInTime(this.queryParams)
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
      });

  }
}
