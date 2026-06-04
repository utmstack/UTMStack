import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {ElasticPieResponse} from '../../../../active-directory/dashboard/active-directory-dashboard/active-directory-dashboard.component';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {ElasticTimeEnum} from '../../../../shared/enums/elastic-time.enum';
import {ActiveDirectoryDashboardService} from '../../../../shared/services/charts-overview/active-directory-dashboard.service';
import {ElasticFilterCommonType} from '../../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

@Component({
  selector: 'app-chart-ad-admin-vs-user',
  templateUrl: './chart-ad-admin-vs-user.component.html',
  styleUrls: ['./chart-ad-admin-vs-user.component.scss']
})
export class ChartAdAdminVsUserComponent implements OnInit, OnDestroy {
  @Input() refreshInterval;
  pieAdminVsUser: any;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.DAY, last: 7, label: 'last 7 days'};
  loadingAdminVsUser = true;
  echartEnum = ChartTypeEnum;
  time: TimeFilterType;
  interval: any;

  constructor(private adDashboardService: ActiveDirectoryDashboardService,
              private router: Router) {
  }

  ngOnInit() {
    this.loadUserVsAdministrators();
    if (this.refreshInterval) {
      this.interval = setInterval(() => {
        this.loadUserVsAdministrators();
      }, this.refreshInterval);
    }
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  onChangeAdminsVsUserDateChange($event: TimeFilterType) {
    // this.time = $event;
    // this.loadUserVsAdministrators({from: $event.timeFrom, to: $event.timeTo});
  }

  loadUserVsAdministrators() {
    this.adDashboardService.getAmountOfAdminsVsUsers().subscribe(admins => {
      this.buildPieChart(admins.body).then(pie => {
        this.pieAdminVsUser = pie;
        this.loadingAdminVsUser = false;
      });
    });
  }

  buildDataPie(response: ElasticPieResponse[]): { name: string, value: number }[] {
    const data: { name: string, value: number }[] = [];
    for (const res of response) {
      data.push({name: res.bucketKey, value: res.value});
    }
    return data;
  }

  buildPieLegend(response: ElasticPieResponse[]): string[] {
    const data: string[] = [];
    for (const res of response) {
      data.push(res.bucketKey);
    }
    return data;
  }

  buildPieChart(response: ElasticPieResponse[]): Promise<any> {
    return new Promise<any>((resolve) => {
      const pie = {
        color: UTM_COLOR_THEME,
        grid: {
          top: 0,
          right: 0,
          bottom: '50px',
          left: 0,
        },
        tooltip: {
          trigger: 'item',
          formatter: '{a} <br/>{b} : {c} ({d}%)',
          backgroundColor: 'rgba(0,75,139,0.85)',
          padding: 10,
          textStyle: {
            fontSize: 13,
            fontFamily: 'Roboto, sans-serif'
          }
        },
        legend: {
          orient: 'vertical',
          top: 'center',
          left: 0,
          itemHeight: 8,
          itemWidth: 8,
          data: this.buildPieLegend(response),
        },
        series: [
          {
            name: 'User type',
            type: 'pie',
            radius: ['50%', '80%'],
            center: ['57.5%', '50.5%'],
            itemStyle: {
              normal: {
                borderWidth: 1,
                borderColor: '#fff',
                label: {
                  show: false
                },
                labelLine: {
                  show: true
                }
              },
              // emphasis: {
              //   label: {
              //     show: true,
              //     position: 'center',
              //     textStyle: {
              //       fontSize: '15',
              //       fontWeight: '500'
              //     }
              //   }
              // }
            },
            data: this.buildDataPie(response)
          }
        ]
      };
      resolve(pie);
    });
  }

  navigateToUserDetailChart($event: any) {
    this.router.navigate(['/active-directory/detail/users'], {
      queryParams: {
        userType: $event.data.name.toLowerCase()
      }
    });
  }

}
