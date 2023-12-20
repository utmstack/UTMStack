import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ChartModel} from '../../../../shared/chart/types/chart.model';
import {UTM_COLOR_THEME} from '../../../../shared/constants/utm-color.const';
import {OpenvasOptionModel} from '../../../shared/model/assets/openvas-chart/openvas-option.model';

@Injectable({
  providedIn: 'root'
})
export class WordCloudDef {

  constructor() {
  }

  public buildChartWordCloud(chart: OpenvasOptionModel) {
    let words: any = {};
    this.processWordCloud(chart).subscribe(wordOption => {
      words = {
        color: UTM_COLOR_THEME,
        tooltip: {
          show: true,
          trigger: 'item',
          // formatter: '{a}',
          backgroundColor: 'rgba(0,75,139,0.85)',
          padding: 10,
          textStyle: {
            fontSize: 13,
            fontFamily: 'Roboto, sans-serif'
          }
        },
        grid: {
          top: '0px',
          bottom: '0px',
          right: '0px',
          left: '0px'
        },
        series: [{
          name: 'Vulnerability',
          type: 'wordCloud',
          size: ['100%', '100%'],
          textRotation: [0, 45, 90, -45],
          textPadding: 0,
          autoSize: {
            enable: true,
            minSize: 14
          },
          data: wordOption.series
        }]
      };
    });
    return words;
  }

  createRandomItemStyle() {
    return {
      normal: {
        fontFamily: 'Roboto, sans-serif',
        color: 'rgb(' + [
          Math.round(Math.random() * 79),
          Math.round(Math.random() * 79),
          Math.round(Math.random() * 139)
        ].join(',') + ')'
      }
    };
  }

  private processWordCloud(chart: OpenvasOptionModel): Observable<ChartModel> {
    return new Observable<ChartModel>((data) => {
      const config: ChartModel = {series: [], values: [], legend: []};
      for (const group of chart.groups) {
        const word = group.value;
        const value = Number.parseFloat(group.count);
        if (value >= 2) {
          config.series.push({
            name: word,
            textStyle: this.createRandomItemStyle(),
            value
          });
        }
      }
      data.next(config);
    });
  }

}
