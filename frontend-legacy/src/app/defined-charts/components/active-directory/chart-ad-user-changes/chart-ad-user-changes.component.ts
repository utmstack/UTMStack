import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {SerieValue} from '../../../../shared/chart/types/response/multiline-response.type';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-ad-user-changes',
  templateUrl: './chart-ad-user-changes.component.html',
  styleUrls: ['./chart-ad-user-changes.component.scss']
})
export class ChartAdUserChangesComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  @Input() top: number;
  @Input() change: 'made' | 'received';
  interval: any;
  echartEnum = ChartTypeEnum;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  time: TimeFilterType;
  loadingTopUserChange = true;
  topUserMoreChange: any;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router) {
  }

  ngOnInit() {
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadUserMoreChanging(
          {
            from: this.time.timeFrom,
            to: this.time.timeTo,
            top: this.top,
            changesMade: this.change === 'made'
          });
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  topUserMostChangedDateChange($event: TimeFilterType) {
    this.time = $event;
    const req = {
      from: $event.timeFrom,
      to: $event.timeTo,
      top: this.top,
      changesMade: this.change === 'made'
    };
    this.loadUserMoreChanging(req);
  }

  loadUserMoreChanging(req) {
    this.adDashboardService.getTopUserMostChanged(req).subscribe(mostChanged => {
      this.buildBar(mostChanged.body).then(bar => {
        this.loadingTopUserChange = false;
        this.topUserMoreChange = bar;
      });
    });
  }

  buildBar(data: SerieValue[]): Promise<any> {
    return new Promise<any>(resolve => {
      if (!data || data.length === 0) {
        resolve(null);
      }
      const bar = {
        color: UTM_COLOR_THEME,
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(0,75,139,0.85)',
          padding: 10,
          textStyle: {
            fontSize: 13,
            fontFamily: 'Roboto, sans-serif'
          }
        },
        grid: {
          left: 15,
          top: 0,
          right: 15,
          bottom: '20px'
        },
        calculable: true,
        xAxis: [{
          type: 'value',
          boundaryGap: [0, 0.01],
          axisLabel: {
            color: '#333'
          },
          axisLine: {
            lineStyle: {
              color: '#999'
            }
          },
          splitLine: {
            show: true,
            lineStyle: {
              color: '#eee',
              type: 'dashed'
            }
          }
        }],
        yAxis: [
          {
            type: 'category',
            data: this.buildBarCategory(data),
            axisLabel: {
              inside: true,
              textStyle: {
                color: '#000',
              }
            },
            label: {fontSize: '10px'},
            z: 10
          }
        ],
        series: [
          {
            name: 'Events',
            type: 'bar',
            barMaxWidth: '25px',
            data: this.buildBarData(data)
          },
        ]
      };
      resolve(bar);
    });
  }

  buildBarCategory(data: SerieValue[]): string[] {
    const category: string[] = [];
    for (const d of data) {
      category.push(d.serie);
    }
    return category;
  }

  buildBarData(data: SerieValue[]): number[] {
    const value: number[] = [];
    for (const d of data) {
      value.push(d.value);
    }
    return value;
  }

  navigateToUsersChanges($event: any) {
    this.router.navigate(['/active-directory/detail/users'], {
      queryParams: {
        userChange: $event.name.toLowerCase()
      }
    });
  }
}
