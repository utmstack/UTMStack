import {ElasticOperatorsEnum} from '../../../../enums/elastic-operators.enum';
import {IndexPatternSystemEnumName} from '../../../../enums/index-pattern-system.enum';
import {EchartClickAction} from '../../../types/action/echart-click-action';
import {VisualizationType} from '../../../types/visualization.type';
import {ChartClickInterface} from '../chart-click.interface';

export class SingleBucketChartClick implements ChartClickInterface {
  buildClickParams(visualization: VisualizationType, chartValue: EchartClickAction): object {
    const queryParams: any = {};
    let field: string;
    visualization.filterType.forEach(value => {
      if (value.operator === ElasticOperatorsEnum.EXIST || value.operator === ElasticOperatorsEnum.DOES_NOT_EXIST) {
        queryParams[value.field] = value.operator + '->';
      } else {
        queryParams[value.field] = value.operator + '->' + value.value;
      }
    });
    if (visualization.aggregationType.bucket) {
      field = visualization.aggregationType.bucket.field;
      queryParams[field] = chartValue.data.name;
    }
    if (visualization.pattern.pattern === IndexPatternSystemEnumName.ALERT) {
      queryParams.alertType = 'ALERT';
    } else {
      queryParams.patternId = visualization.pattern.id;
      queryParams.indexPattern = visualization.pattern.pattern;
    }

    return Object.keys(queryParams).length === 0 ? null : queryParams;
  }
}
