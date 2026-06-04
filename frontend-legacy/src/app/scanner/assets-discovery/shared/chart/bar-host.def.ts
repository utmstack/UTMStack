import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ChartModel} from '../../../../shared/chart/types/chart.model';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {VsSeverityResolverService} from '../../../../shared/services/scan/vs-severity-resolver.service';
import {AssetModel} from '../../../shared/model/assets/asset.model';
import {OpenvasOptionModel} from '../../../shared/model/assets/openvas-chart/openvas-option.model';

@Injectable({
  providedIn: 'root'
})
export class BarHostDef {

  constructor(private assetSeverityResolverService: VsSeverityResolverService) {
  }

  public buildCharMostVulnerableHost(chart: OpenvasOptionModel, assets: AssetModel[]) {
    let bar: any = {};
    this.processMostVulnerableHost(chart, assets).subscribe(barOption => {
      bar = {
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
          top: '35px',
          right: 15,
          bottom: '35px'
        },
        calculable: true,
        yAxis: [{
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
        xAxis: [
          {
            type: 'category',
            data: barOption.legend
          }
        ],
        series: [
          {
            name: 'Severity',
            type: 'bar',
            barMaxWidth: '25px',
            data: barOption.values
          },
        ]
      };
    });
    return bar;
  }

  // SO build option
  public buildCharOperatingSystemsByVulnerabilityScore(chart: OpenvasOptionModel, assets: AssetModel[]) {
    let barSo: any = {};
    this.processMostVulnerableSo(chart, assets).subscribe(barOption => {
      barSo = {
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
          left: '180px',
          top: '35px',
          right: 15,
          bottom: 0
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
            data: barOption.legend
          }
        ],
        series: [
          {
            name: 'Severity',
            type: 'bar',
            itemStyle: {
              // normal: {
              //   color: '#ff705a'
              // }
            },
            data: barOption.values
          },
        ]
      };
    });
    return barSo;
  }

  private processMostVulnerableHost(chart: OpenvasOptionModel, assets: AssetModel[]): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {
        series: [], values: [], legend: []
      };
      if (chart.groups !== null && assets !== null) {
        chart.groups = Array.from(new Set(chart.groups)).sort((a, b) => {
          return Number.parseFloat(a.stats[0].max) < Number.parseFloat(b.stats[0].max) ? 1 : -1;
        });
        for (let i = 0; i < (chart.groups.length < 10 ? chart.groups.length : 10); i++) {
          if (chart.groups[i].stats[0].max !== '0') {
            const index = assets.findIndex(value => value.uuid === chart.groups[i].value);
            if (index !== -1) {
              config.legend.push(assets[index].name);
              config.values.push({
                value: Number.parseFloat(chart.groups[i].stats[0].max),
                itemStyle: {
                  color: this.assetSeverityResolverService.resolveColor(chart.groups[i].stats[0].max)
                }
              });
            }
          }
        }
        data.next(config);
      } else {
        data.next(null);
      }
    });
  }

  private processMostVulnerableSo(chart: OpenvasOptionModel, assets: AssetModel[]): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {
        series: [], values: [], legend: []
      };
      if (chart.groups !== null && assets !== null) {
        chart.groups = Array.from(new Set(chart.groups)).sort((a, b) => {
          return Number.parseFloat(a.stats[0].max) < Number.parseFloat(b.stats[0].max) ? 1 : -1;
        });
        for (let i = 0; i < (chart.groups.length < 10 ? chart.groups.length : 10); i++) {
          const statIndex = chart.groups[i].stats.findIndex(value => value.column === 'average_severity');
          if (chart.groups[i].stats[statIndex].max !== '0' && statIndex !== -1) {
            const assetIndex = assets.findIndex(value => value.uuid === chart.groups[i].value);
            if (assetIndex !== -1) {
              config.legend.push(assets[assetIndex].name);
              config.values.push({
                type: 'bar',
                name: assets[assetIndex].name,
                value: Number.parseFloat(chart.groups[i].stats[statIndex].max),
                itemStyle: {
                  color: this.assetSeverityResolverService.resolveColor(chart.groups[i].stats[statIndex].max)
                }
              });
            }
          }
        }
        data.next(config);
      } else {
        data.next(null);
      }
    });
  }
}
