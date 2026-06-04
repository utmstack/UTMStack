import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';

export function rebuildVisualizationFilterTime(time: TimeFilterType, elasticFilters: ElasticFilterType[]): Promise<ElasticFilterType[]> {
  return new Promise<ElasticFilterType[]>(resolve => {
    const indexTime = elasticFilters.findIndex(value => value.field === '@timestamp');
    const filter = {
      field: '@timestamp',
      operator: ElasticOperatorsEnum.IS_BETWEEN.toString(),
      value: [time.timeFrom, time.timeTo]
    };
    if (indexTime === -1) {
      elasticFilters.push(filter);
    } else {
      elasticFilters[indexTime].value = [time.timeFrom, time.timeTo];
    }
    resolve(elasticFilters);
  });
}
