import {
  extractMetricLabel,
  getBucketLabel
} from '../../../../../graphic-builder/chart-builder/chart-property-builder/shared/functions/visualization-util';
import {UtmTagCloudOptionType} from '../../../types/charts/tag-cloud/utm-tag-cloud-option.type';
import {PieBuilderResponseType} from '../../../types/response/pie-builder-response.type';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartBuildInterface} from '../chart-build.interface';
import {ChartOption} from '../chart-option';

export class TagCloud implements ChartBuildInterface {

  constructor() {
  }

  buildChart(data?: any[], visualization?: VisualizationType): ChartOption {
    const tagOptions: UtmTagCloudOptionType = visualization.chartConfig;
    tagOptions.series[0].data = this.extractTagValues(data, visualization);
    tagOptions.series[0].name = getBucketLabel(0, visualization);
    return tagOptions;
  }

  extractTagValues(data: PieBuilderResponseType[], visualization: VisualizationType):
    { name: string, value: number, textStyle: any }[] {
    const values: { name: string, value: number, textStyle: any }[] = [];
    const tagOptions: UtmTagCloudOptionType = visualization.chartConfig;
    if (data) {
      for (const dat of data) {
        values.push(
          {
            name: dat.bucketKey ? dat.bucketKey : extractMetricLabel(dat.metricId, visualization),
            value: Number(dat.value.toFixed(2)),
            textStyle: this.createRandomItemStyle(tagOptions.series[0].color),
          }
        );
      }
    } else {
      values.push(
        {
          name: 'All',
          value: 100,
          // value: Number(dat.value.toFixed(2)),
          textStyle: this.createRandomItemStyle(tagOptions.series[0].color),
        }
      );
    }
    return values;
  }

  createRandomItemStyle(colors: string[]) {
    return {
      normal: {
        fontFamily: 'Roboto, sans-serif',
        color: colors[this.randomIndex(0, colors.length)]
      }
    };
  }

//   'rgb(' + [
//     Math.round(Math.random() * 79),
//   Math.round(Math.random() * 79),
//   Math.round(Math.random() * 139)
// ].join(',') + ')'
  randomIndex(min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min);
  }

}
