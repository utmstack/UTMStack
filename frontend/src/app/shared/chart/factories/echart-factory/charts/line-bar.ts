import {extractMetricLabel} from '../../../../../graphic-builder/chart-builder/chart-property-builder/shared/functions/visualization-util';
import {ChartTypeEnum} from '../../../../enums/chart-type.enum';
import {ChartValueSeparator} from '../../../../enums/chart-value-separator';
import {SeriesLine} from '../../../types/charts/chart-properties/series/line/series-line';
import {ItemStyle} from '../../../types/charts/chart-properties/style/item-style';
import {UtmLineBarOptionType} from '../../../types/charts/line/utm-line-bar-option.type';
import {BarLineResponseType} from '../../../types/response/bar-line-response.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartBuildInterface} from '../chart-build.interface';
import {ChartOption} from '../chart-option';

export class LineBar implements ChartBuildInterface {
  chartEnumType = ChartTypeEnum;

  constructor() {
  }

  buildChart(data?: BarLineResponseType[], visualization?: VisualizationType, toExport?: boolean): ChartOption {
    /**
     * Condition to run build object of echart
     * - data exist, data is the response of run visualization,
     * - chartConfig.seriesOption exist, this is user defined metrics by user
     * -data[0].series size is greater than 0 mean that category can exist but series does not have data,
     * so return null
     */
    if (data && visualization &&
      visualization.chartConfig.seriesOption &&
      visualization.chartConfig.seriesOption.length > 0 &&
      data[0] && data[0].series &&
      data[0].series.length > 0) {
      const lineOptions: UtmLineBarOptionType = visualization.chartConfig;
      /**
       * Condition to build single or complex object for bar chart
       * -!visualization.aggregationType.bucket, mean that no have any bucket, with this show only one bar with total
       * -visualization.aggregationType.bucket.subBucket, mean tha have a hierarchy with more than on level, on this
       * case build complex
       * -data[0].series.length mean that have only one metric
       * - visualization type for single must have bar type
       */
      if (!visualization.aggregationType.bucket ||
        visualization.aggregationType.bucket.subBucket ||
        data[0].series.length > 1 ||
        (visualization.chartType === this.chartEnumType.LINE_CHART ||
          visualization.chartType === this.chartEnumType.AREA_LINE_CHART)) {
        this.resolveSerie(data, visualization, toExport).then(series => {
          lineOptions.series = series;
        });
        if (this.chartEnumType.BAR_HORIZONTAL_CHART === visualization.chartType) {
          lineOptions.yAxis.data = data[0].categories;
        } else {
          lineOptions.xAxis.data = data[0].categories;
        }
        lineOptions.legend.data = this.resolveLegend(data, visualization);
      } else {
        lineOptions.series = this.resolveSimpleBarSerie(data, visualization, toExport);
        lineOptions.legend.data = data[0].categories;
        if (this.chartEnumType.BAR_HORIZONTAL_CHART === visualization.chartType) {
          lineOptions.yAxis.data = this.getCategoriesSingleBar(data, visualization);
        } else {
          lineOptions.xAxis.data = this.getCategoriesSingleBar(data, visualization);
        }
        lineOptions.legend.data = data[0].categories;
      }
      return lineOptions;
    } else {
      return null;
    }
  }

  resolveLegendSingleBar(data?: BarLineResponseType[], visualization?: VisualizationType): string[] {
    const legendSingle: string[] = [];
    const metricId = data[0].series[0].metricId;
    for (const series of data[0].series) {
      const leg = series;
      legendSingle.push(leg.name);
    }
    return legendSingle;
  }

  getCategoriesSingleBar(data?: BarLineResponseType[], visualization?: VisualizationType): string[] {
    const categories: string[] = [];
    for (const serie of data[0].series) {
      const cat = serie.name === '' ? extractMetricLabel(serie.metricId, visualization) : serie.name;
      categories.push(cat);
    }
    return categories;
  }

  resolveLegend(data?: BarLineResponseType[], visualization?: VisualizationType): string[] {
    const legend: string[] = [];

    for (const series of data[0].series) {
      const ser = series.name === '' ? extractMetricLabel(series.metricId, visualization) : series.name;
      legend.push(ser);
    }
    return legend;
  }

  resolveSimpleBarSerie(data?: BarLineResponseType[], visualization?: VisualizationType, toExport?: boolean) {
    const serie: SeriesLine[] = visualization.chartConfig.seriesOption;
    const serieReturn: SeriesLine[] = [];
    // tslint:disable-next-line:prefer-for-of
    for (let i = 0; i < data[0].categories.length; i++) {
      // tslint:disable-next-line:prefer-for-of
      for (let j = 0; j < data[0].series.length; j++) {
        const metricId = Number(data[0].series[j].metricId);
        const index = serie.findIndex(value => Number(value.metricId) === metricId);
        const metricLabel = extractMetricLabel(data[0].series[j].metricId, visualization);
        const stackBy = data[0].series[j].name === '' ? metricLabel : data[0].series[j].name;
        if (index !== -1) {
          const ser = {
            name: data[0].categories[i],
            type: serie[index].type,
            data: [{value: data[0].series[j].data[i], name: metricLabel}],
            markPoint: serie[index].markPoint,
            markLine: serie[index].markLine,
            smooth: true,
            itemStyle: this.resolveItemStyle(serie[index], toExport),
          };
          serieReturn.push(ser);
        }
      }
    }
    return serieReturn;
  }

  resolveSerie(data: BarLineResponseType[], visualization?: VisualizationType, toExport?: boolean): Promise<SeriesLine[]> {
    return new Promise<SeriesLine[]>(resolve => {
      const serie: SeriesLine[] = visualization.chartConfig.seriesOption;
      const serieReturn: SeriesLine[] = [];
      for (const dat of data[0].series) {
        const metricId = Number(dat.metricId);
        const index = serie.findIndex(value => Number(value.metricId) === metricId);
        if (index !== -1) {
          const ser = {
            name: dat.name === '' ? extractMetricLabel(metricId, visualization) : dat.name,
            type: serie[index].type,
            data: dat.data,
            markPoint: serie[index].markPoint,
            markLine: serie[index].markLine,
            smooth: true,
            itemStyle: this.resolveItemStyle(serie[index], toExport, visualization.chartType),
          };
          serieReturn.push(ser);
        }
      }
      resolve(serieReturn);
    });
  }

  resolveItemStyle(serie: SeriesLine, exportTo?: boolean, chartType?: ChartTypeEnum): ItemStyle {
    if (serie.type === 'line' && chartType === this.chartEnumType.AREA_LINE_CHART) {
      return new ItemStyle({
        areaStyle: {
          type: 'mint',
          normal: {
            opacity: 0.25
          }
        }
      });
    } else {
      if (exportTo) {
        return new ItemStyle({
          label: {
            show: true,
            position: ['45%', '70%'],
            rotate: chartType === ChartTypeEnum.BAR_HORIZONTAL_CHART ? 0 : 90,
            fontSize: 13,
            fontWeight: 500,
            color: '#000'
          }
        });
      }
    }
  }

  getNameOfCategory(category: string, visualization: VisualizationType) {
    return visualization.aggregationType.bucket.field + ChartValueSeparator.VALUE_SEPARATOR + category;
  }

}
