import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ChartModel} from '../../../../shared/chart/types/chart.model';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {VsSeverityResolverService} from '../../../../shared/services/scan/vs-severity-resolver.service';
import {ScannerSeverityEnum} from '../../../shared/enums/scanner-severity.enum';
import {OpenvasOptionModel} from '../../../shared/model/assets/openvas-chart/openvas-option.model';

@Injectable({
  providedIn: 'root'
})
export class MultilineSeverityDef {
  legend = [ScannerSeverityEnum.HIGH,
    ScannerSeverityEnum.MEDIUM,
    ScannerSeverityEnum.LOW,
    ScannerSeverityEnum.UNKNOWN];

  constructor(private assetSeverityResolverService: VsSeverityResolverService) {
  }

  public buildChartHostsByModificationTime(chart: OpenvasOptionModel) {
    let multiline: any = {};
    this.processHostsByModificationTime(chart).subscribe(multilineOption => {
      multiline = {
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
          left: 0,
          right: '40px',
          top: '35px',
          bottom: '50px',
          containLabel: true
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          axisLabel: {
            color: '#333'
          },
          axisLine: {
            lineStyle: {
              color: '#999'
            }
          },
          data: multilineOption.legend
        },
        yAxis: [{
          type: 'value',
          axisLabel: {
            formatter: '{value} ',
            color: '#333'
          },
          axisLine: {
            lineStyle: {
              color: '#999'
            }
          },
          splitLine: {
            lineStyle: {
              color: ['#eee']
            }
          },
          splitArea: {
            show: true,
            areaStyle: {
              color: ['rgba(250,250,250,0.1)', 'rgba(0,0,0,0.01)']
            }
          }
        }],
        series: [{
          type: 'line',
          smooth: true,
          // itemStyle: {
          //   normal: {
          //     areaStyle: {
          //       type: 'mint'
          //     }
          //   }
          // },
          data: multilineOption.values,
        }],
        dataZoom: [
          {
            type: 'inside',
            start: 30,
            end: 70
          },
          {
            show: true,
            type: 'slider',
            start: 30,
            end: 70,
            height: 40,
            bottom: 0,
            borderColor: '#ccc',
            fillerColor: 'rgba(40,54,139,0.21)',
            handleStyle: {
              color: '#004b8b'
            }
          }
        ]
      };
    });
    return multiline;
  }

  processHostsByModificationTime(chart: OpenvasOptionModel): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {series: [], values: [], legend: []};
      if (chart[0].groups !== null) {
        for (const group of chart[0].groups) {
          const valueName = group.value.replace('T00:00:00Z', '');
          const valueCount = Number.parseFloat(group.count);
          const valueIndex = config.values.findIndex(value => value === valueName);
          if (valueIndex === -1) {
            config.legend.push(valueName);
            config.values.push(valueCount);
          }
        }
        data.next(config);
      } else {
        data.next(null);
      }
    });
  }
}
