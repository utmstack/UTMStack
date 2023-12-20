import {
  extractMetricLabel,
  getBucketLabel
} from '../../../../../graphic-builder/chart-builder/chart-property-builder/shared/functions/visualization-util';
import {Legend} from '../../../types/charts/chart-properties/legend/legend';
import {SeriesPie} from '../../../types/charts/chart-properties/series/pie/series-pie';
import {ItemStyle} from '../../../types/charts/chart-properties/style/item-style';
import {Tooltip} from '../../../types/charts/chart-properties/tooltip/tooltip';
import {UtmPieOptionType} from '../../../types/charts/pie/utm-pie-option.type';
import {PieBuilderResponseType} from '../../../types/response/pie-builder-response.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartBuildInterface} from '../chart-build.interface';
import {ChartOption} from '../chart-option';

export class Pie implements ChartBuildInterface {

  constructor() {
  }

  buildChart(data: PieBuilderResponseType[], visualization?: VisualizationType, toExport?: boolean): ChartOption {
    const pieOptions: UtmPieOptionType = visualization.chartConfig;
    const legend: Legend = pieOptions.legend;
    legend.data = this.extractSeries(data, visualization);
    return {
      color: pieOptions.color,
      tooltip: new Tooltip('item', '{a} <br/>{b} : {c} ({d}%)'),
      legend,
      series: new SeriesPie(
        this.extractPieValues(data, visualization),
        'pie',
        this.extractSerieName(data, visualization),
        pieOptions.pieType === 'donut' ? ['50%', '80%'] : null,
        ['50%', '50%'],
        new ItemStyle({
            borderWidth: 1,
            borderColor: '#fff',
            label: {
              show: toExport,
              formatter: toExport ? '{c}({d}%)' : null
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
      ),
      toolbox: pieOptions.toolbox,
      calculable: true,
      grid: pieOptions.grid,
    };
  }

  extractSeries(data: any[], visualization: VisualizationType): string[] {
    const series: string[] = [];
    for (const dat of data) {
      series.push(dat.bucketKey ? dat.bucketKey : extractMetricLabel(dat.metricId, visualization));
    }
    return series;
  }

  extractPieValues(data: PieBuilderResponseType[], visualization: VisualizationType): { name: string, value: number }[] {
    const values: { name: string, value: number }[] = [];
    for (const dat of data) {
      values.push(
        {
          name: dat.bucketKey ? dat.bucketKey : extractMetricLabel(dat.metricId, visualization),
          value: Number(dat.value.toFixed(2))
        }
      );
    }
    return values;
  }

  private extractSerieName(data: PieBuilderResponseType[], visualization: VisualizationType): string {
    return getBucketLabel(0, visualization);
  }

}
