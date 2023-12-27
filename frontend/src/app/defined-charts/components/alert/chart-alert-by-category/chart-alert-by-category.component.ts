import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Legend} from '../../../../shared/chart/types/charts/chart-properties/legend/legend';
import {Tooltip} from '../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {ALERT_CATEGORY_FIELD, ALERT_GLOBAL_FIELD, ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {OverviewAlertDashboardService} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {deleteNullValues} from '../../../../shared/util/object-util';

@Component({
  selector: 'app-chart-alert-by-category',
  templateUrl: './chart-alert-by-category.component.html',
  styleUrls: ['./chart-alert-by-category.component.scss']
})
export class ChartAlertByCategoryComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  interval: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  queryParams = {from: 'now-7d', to: 'now', top: 10};
  loadingBarOption = true;
  chartEnumType = ChartTypeEnum;
  barOption: any;
  noData = false;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.queryParams.top = 10;
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getCategoryData();
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.queryParams.from = $event.timeFrom;
    this.queryParams.to = $event.timeTo;
    this.getCategoryData();
  }


  getCategoryData() {
    this.overviewAlertDashboardService.getAlertByCategory(this.queryParams)
      .subscribe((category) => {
        this.loadingBarOption = false;
        if (category.body.categories.length > 0) {
          this.noData = false;
          this.buildCategoryChart(category.body);
        } else {
          this.noData = true;
        }
      });
  }

  buildCategoryChart(data: { categories: string[], series: number[] }) {
    if (data.categories.length > 0) {
      this.barOption = {
        color: UTM_COLOR_THEME,
        animation: true,
        tooltip: new Tooltip('item'),
        legend: new Legend(true, data.categories, 'scroll', 'horizontal',
          'bottom', 'center'),
        xAxis: {
          type: 'category',
          z: 10
        },
        yAxis: {
          type: 'value'
        },
        series: this.resolveBarSeries(data)
      };
      this.barOption = deleteNullValues(this.barOption);
    }
  }

  resolveBarSeries(data: { categories: string[], series: number[] }) {
    const seriesData: BarSimpleSeries[] = [];
    for (let i = 0; i < data.categories.length; i++) {
      seriesData.push({
        name: data.categories[i],
        type: 'bar',
        barMaxWidth: '50px',
        data: [{value: data.series[i], name: 'Count'}]
      });
    }
    return seriesData;
  }

  buildDataSeries(data: { categories: string[], series: number[] }): { name: string, value: number }[] {
    const seriesData: { name: string, value: number }[] = [];
    for (let i = 0; i < data.categories.length; i++) {
      seriesData.push({value: data.series[i], name: data.categories[i]});
    }
    return seriesData;
  }

  chartClick($event) {
    const queryParams = {};
    queryParams[ALERT_CATEGORY_FIELD] = $event.seriesName;
    queryParams[ALERT_GLOBAL_FIELD] = 'ALERT';
    queryParams[ALERT_TIMESTAMP_FIELD] = ElasticOperatorsEnum.IS_BETWEEN + '->'
      + this.queryParams.from + ',' + this.queryParams.to;
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/data/alert/view'], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

}

export class BarSimpleSeries {
  name: string;
  barMaxWidth: string;
  type: 'bar';
  data: { value: number, name: string }[];
}
