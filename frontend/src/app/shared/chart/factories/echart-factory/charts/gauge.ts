import {extractMetricLabel} from '../../../../../graphic-builder/chart-builder/chart-property-builder/shared/functions/visualization-util';
import {SeriesGauge} from '../../../types/charts/chart-properties/series/gauge/series-gauge';
import {UtmGaugeOptionType} from '../../../types/charts/gauge/utm-gauge-option.type';
import {GaugeBuilderResponseType} from '../../../types/response/gauge-builder-response.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartBuildInterface} from '../chart-build.interface';
import {ChartOption} from '../chart-option';

export class Gauge implements ChartBuildInterface {
  constructor() {
  }

  buildChart(data: GaugeBuilderResponseType[], visualization?: VisualizationType): ChartOption {
    const gaugeOptions: UtmGaugeOptionType = visualization.chartConfig;
    return {
      toolbox: gaugeOptions.toolbox,
      calculable: true,
      grid: gaugeOptions.grid,
      series: this.buildGaugeSerie(data, visualization),
    };
  }

  buildGaugeSerie(data: GaugeBuilderResponseType[], visualization?: VisualizationType): SeriesGauge[] {
    const options: UtmGaugeOptionType = visualization.chartConfig;
    let offsetLeft = 15;
    let offsetTop = 45;
    const series: SeriesGauge[] = [];
    if (data.length > 1) {
      for (let i = 0; i < data.length; i++) {
        if (i === 3) {
          offsetLeft = 15;
          offsetTop += 45;
        }
        if (data.length === 3) {
          offsetTop = 60;
        }
        const serie: SeriesGauge = options.serie[0];
        serie.data = [
          {
            name: data[i].bucketKey,
            value: data[i].value
          }
        ];
        serie.radius = '50%';
        serie.center = [offsetLeft + '%', offsetTop + '%'];
        serie.title.offsetCenter = [0, -115];
        serie.splitLine.length = serie.axisLine.lineStyle.width + 5;

        const s = new SeriesGauge(
          'gauge',
          serie.startAngle,
          serie.endAngle,
          serie.min,
          serie.max,
          serie.radius,
          serie.center,
          serie.splitNumber,
          serie.axisLine,
          serie.axisTick,
          serie.pointer,
          serie.detail,
          serie.data,
          serie.title,
          serie.splitLine);
        series.push(s);
        offsetLeft += 35;
      }
    } else {
      options.serie[0].title.offsetCenter = [0, -45];
      options.serie[0].splitLine.length = options.serie[0].axisLine.lineStyle.width + 5;
      options.serie[0].data = [
        {
          name: extractMetricLabel(data[0].metricId, visualization),
          value: data[0].value
        }
      ];
      series.push(options.serie[0]);
    }
    return series;
  }
}
