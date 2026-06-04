import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ChartModel} from '../../../../shared/chart/types/chart.model';
import {VsSeverityResolverService} from '../../../../shared/services/scan/vs-severity-resolver.service';
import {OpenvasOptionModel} from '../../../shared/model/assets/openvas-chart/openvas-option.model';

@Injectable({
  providedIn: 'root'
})
export class PieSeverityClassDef {

  constructor(private severityResolver: VsSeverityResolverService) {
  }

  public buildChartBySeverityClass(chart: OpenvasOptionModel) {
    let pie: any = {};
    this.processBySeverityClass(chart).subscribe(pieOption => {
      pie = {
        grid: {
          top: '35px',
          right: 0,
          bottom: 0,
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
          data: pieOption.legend,
        },
        series: [
          {
            name: 'Severity',
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
              emphasis: {
                label: {
                  show: true,
                  position: 'center',
                  textStyle: {
                    fontSize: '15',
                    fontWeight: '500'
                  }
                }
              }
            },
            data: pieOption.values
          }
        ]
      };
    });
    return pie;
  }

  public buildChartResultsByCVSS(chart: OpenvasOptionModel) {
    let bar: any = {};
    this.processResultsByCVSS(chart).subscribe(barOption => {
      bar = {
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
          left: '0',
          top: '35px',
          right: '0',
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
            name: 'CVSS',
            type: 'bar',
            data: barOption.values
          },
        ]
      };
    });
    return bar;
  }

  public buildChartOperatingSystemsBySeverityClass(chart: OpenvasOptionModel) {
    let pie: any = {};
    this.processOperatingSystemsBySeverityClass(chart).subscribe(pieOption => {
      pie = {
        grid: {
          top: '0px',
          right: 0,
          bottom: 0,
          left: '35px',
          position: 'center'
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
          data: pieOption.legend,
        },
        series: [
          {
            name: 'Severity',
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
              emphasis: {
                label: {
                  show: true,
                  position: 'center',
                  textStyle: {
                    fontSize: '15',
                    fontWeight: '500'
                  }
                }
              }
            },
            data: pieOption.values
          }
        ]
      };
    });
    return pie;
  }

  private processBySeverityClass(chart: OpenvasOptionModel): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {series: [], values: [], legend: []};
      if (chart.groups !== null) {
        for (const group of chart.groups) {
          const severityName = this.severityResolver.resolveSeverityLabel(group.value);
          config.legend.push(severityName);
          const valuesIndex = config.values.findIndex(value => value.name === severityName);
          if (valuesIndex !== -1) {
            config.values[valuesIndex].value += Number.parseFloat(group.count);
          } else {
            config.values.push({
              name: severityName,
              itemStyle: {
                color: this.severityResolver.resolveColorByName(severityName),
              },
              value: Number.parseFloat(group.count)
            });
          }
        }
        config.legend = Array.from(new Set(config.legend)).sort();
        config.values = Array.from(new Set(config.values)).sort((a, b) => {
          return a.name > b.name ? 1 : -1;
        });
        data.next(config);
      } else {
        data.next(null);
      }
    });
  }

  private processResultsByCVSS(chart: OpenvasOptionModel): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {series: [], values: [], legend: []};
      if (chart.groups !== null) {
        for (const group of chart.groups) {
          const value = this.severityResolver.resolveSeverityFixed(group.value);
          config.legend.push(value);
          config.values.push({
            name: value,
            itemStyle: {
              color: this.severityResolver.resolveColor(value),
            },
            value: Number.parseFloat(group.count)
          });
        }
        config.legend = Array.from(new Set(config.legend)).sort();
        config.values = Array.from(new Set(config.values)).sort((a, b) => {
          return a.name > b.name ? 1 : -1;
        });
        data.next(config);
      } else {
        data.next(null);
      }
    });
  }

  private processOperatingSystemsBySeverityClass(chart: OpenvasOptionModel): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {series: [], values: [], legend: []};
      for (const group of chart.groups) {
        const severityName = this.severityResolver.resolveSeverityLabel(group.value);
        config.legend.push(severityName);
        const valuesIndex = config.values.findIndex(value => value.name === severityName);
        if (valuesIndex > -1) {
          config.values[valuesIndex].value += Number.parseFloat(group.count);
        } else {
          config.values.push({
            name: severityName,
            itemStyle: {
              color: this.severityResolver.resolveColorByName(severityName),
            },
            value: Number.parseFloat(group.count)
          });
        }
      }
      config.legend = Array.from(new Set(config.legend)).sort();
      config.values = Array.from(new Set(config.values)).sort((a, b) => {
        return a.name > b.name ? 1 : -1;
      });
      data.next(config);
    });
  }

}
