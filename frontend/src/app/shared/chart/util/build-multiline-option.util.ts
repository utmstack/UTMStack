import {UTM_COLOR_THEME} from '../../constants/utm-color.const';
import {Legend} from '../types/charts/chart-properties/legend/legend';
import {MultilineSerie} from '../types/charts/chart-properties/series/line/multiline-serie';
import {MultilineResponse} from '../types/response/multiline-response.type';

export function buildMultilineObject(data: MultilineResponse): Promise<any> {
  return new Promise<any>(resolve => {
    const multiline = {
      animation: true,
      color: UTM_COLOR_THEME,
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(0,75,139,0.85)',
        padding: 10,
        textStyle: {
          fontSize: 13,
          fontFamily: 'Roboto, sans-serif'
        }
      },
      // toolbox: {
      //   show: true,
      //   feature: {
      //     saveAsImage: {
      //       show: true,
      //       type: 'png',
      //       name: 'utm-chart',
      //       title: 'Save as image'
      //     },
      //     restore: {
      //       show: true,
      //       title: 'Restore',
      //     }
      //   }
      // },
      boundaryGap: false,
      grid: {
        top: '40px',
        bottom: '60px',
        left: '70px',
        right: '40px'
      },
      calculable: true,
      legend: new Legend(true, buildMultilineLegend(data),
        'scroll', 'horizontal',
        'top', 'center',
        10, 0, 12, 12),
      xAxis: [
        {
          type: 'category',
          boundaryGap: false,
          axisLabel: {
            show: false,
            rotate: 60,
            textStyle: {
              fontSize: '9px'
            }
          },
          data: formatCategories(data)
        }
      ],
      yAxis: [
        {
          type: 'value',
          axisLabel: {
            formatter: '{value}',
          }
        }
      ],
      series: buildMultilineSeries(data),
      dataZoom: [
        {
          type: 'inside',
          start: 0,
          end: 100
        },
        {
          show: true,
          type: 'slider',
          start: 0,
          end: 100,
          height: 30,
          bottom: 0,
          borderColor: '#ccc',
          fillerColor: 'rgba(40,54,139,0.21)',
          handleStyle: {
            color: '#004b8b'
          }
        }
      ]
    };
    resolve(multiline);
  });
}

export function formatCategories(data: MultilineResponse) {
  const cat: string[] = [];
  for (const category of data.categories) {
    cat.push(category);
  }
  return cat;
}

export function buildMultilineLegend(data: MultilineResponse): string[] {
  const legend: string[] = [];
  for (const leg of data.series) {
    legend.push(leg.serie);
  }
  return legend;
}

export function buildMultilineSeries(data: MultilineResponse): MultilineSerie[] {
  const serie: MultilineSerie[] = [];
  for (const leg of data.series) {
    serie.push(
      {
        name: leg.serie,
        type: 'line',
        smooth: true,
        // itemStyle: {
        //   normal: {
        //     areaStyle: {
        //       type: 'mint',
        //       // normal: {
        //       //   opacity: 0.25
        //       // }
        //     }
        //   }
        // },
        data: leg.values,
      }
    );
  }
  return serie;
}
