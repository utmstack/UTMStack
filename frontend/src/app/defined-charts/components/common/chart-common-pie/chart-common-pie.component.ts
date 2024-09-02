import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, Subject} from 'rxjs';
import {filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {Legend} from '../../../../shared/chart/types/charts/chart-properties/legend/legend';
import {SeriesPie} from '../../../../shared/chart/types/charts/chart-properties/series/pie/series-pie';
import {ItemStyle} from '../../../../shared/chart/types/charts/chart-properties/style/item-style';
import {Tooltip} from '../../../../shared/chart/types/charts/chart-properties/tooltip/tooltip';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {
  OverviewAlertDashboardService
} from '../../../../shared/services/charts-overview/overview-alert-dashboard.service';
import {RefreshService, RefreshType} from '../../../../shared/services/util/refresh.service';
import {PieResponseType} from '../../../../shared/types/chart-reponse/pie-response.type';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-common-pie',
  templateUrl: './chart-common-pie.component.html',
  styleUrls: ['./chart-common-pie.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartCommonPieComponent implements OnInit, OnDestroy {
  @Input() type: RefreshType;
  @Input() header: string;
  @Input() colors: string[];
  @Input() colorsMap: { value: string, color: string }[];
  @Input() endpoint: string;
  @Input() params;
  @Input() paramClick: string;
  @Input() navigateUrl: string;
  @Output() loaded = new EventEmitter<void>();
  interval: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  queryParams = {from: 'now-7d', to: 'now', top: 10};
  loadingPieOption = true;
  chartEnumType = ChartTypeEnum;
  pieOption: any;
  noData = false;
  destroy$ = new Subject<void>();
  data$: Observable<any>;

  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private router: Router,
              private refreshService: RefreshService,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.queryParams.top = 10;
    /*if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.getPieData();
      }, this.refreshInterval);
    }*/

    this.data$ = this.refreshService.refresh$
      .pipe(takeUntil(this.destroy$),
          filter(refreshType => (
              refreshType === RefreshType.ALL || refreshType === this.type)),
          switchMap(() => this.getPieData()));
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.queryParams.from = $event.timeFrom;
    this.queryParams.to = $event.timeTo;
    this.refreshService.sendRefresh(this.type);
  }

  getPieData() {
     return this.overviewAlertDashboardService.getDataPie(this.endpoint, this.queryParams)
      .pipe(
        map(response => response.body),
        tap(data => {
          this.loadingPieOption = false;
          if (data.length > 0) {
            this.noData = false;
            this.buildPieChart(data);
          } else {
            this.noData = true;
          }
          this.loaded.emit();
        }));
  }

  buildPieChart(data: PieResponseType) {
    if (data.value !== null) {
      this.pieOption = {
        animation: true,
        color: this.colors ? this.colors : UTM_COLOR_THEME,
        tooltip: new Tooltip('item', '{b} : {c} ({d}%)', null, 5),
        legend: new Legend(true, data.data, 'scroll', 'horizontal', 'bottom', 'center'),
        series: [
          new SeriesPie(
            this.colorsMap ? this.processMapColor(data) : data.value,
            'pie',
            '',
            ['50%', '80%'],
            ['50%', '50%'],
            new ItemStyle({
                borderWidth: 1,
                borderColor: '#fff',
                label: {
                  show: false
                },
                labelLine: {
                  show: true
                }
              },
              {
                label: {
                  show: true,
                  position: 'center',
                  textStyle: {
                    fontSize: '12',
                    fontWeight: '500'
                  }
                }
              }),
            null
          )
        ]
      };
    }
  }

  chartClick($event) {
    this.params['@timestamp'] = ElasticOperatorsEnum.IS_BETWEEN + '->' + this.queryParams.from + ',' + this.queryParams.to;
    this.params[this.paramClick] = $event.name;
    this.spinner.show('loadingSpinner');
    this.router.navigate([this.navigateUrl], {
      queryParams: this.params
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  getColorByName(name: string): string {
    const indexColor = this.colorsMap.findIndex(map => map.value === name);
    if (indexColor !== -1) {
      return this.colorsMap[indexColor].color;
    } else {
      return UTM_COLOR_THEME[Math.random() * UTM_COLOR_THEME.length];
    }
  }

  private processMapColor(chart: PieResponseType): { name: string, itemStyle: any, value: number }[] {
    let config: { name: string, itemStyle: any, value: number }[] = [];
    for (const val of chart.value) {
      config.push({
        name: val.name,
        itemStyle: {
          color: this.getColorByName(val.name),
        },
        value: val.value
      });
    }
    config = Array.from(new Set(config)).sort((a, b) => {
      return a.name > b.name ? 1 : -1;
    });

    return config;
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
