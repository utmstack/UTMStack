import {FILTER_OPERATORS} from '../constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../enums/elastic-operators.enum';
import {ElasticFilterType} from '../types/filter/elastic-filter.type';
import {OperatorsType} from '../types/filter/operators.type';

/**
 * Remove empty elements from filter, if value is null or empty and operator is distinct than exist
 * @param filters ElasticFilterType
 */
export function sanitizeFilters(filters: ElasticFilterType[]) {
  // tslint:disable-next-line:prefer-for-of
  for (let i = 0; i < filters.length; i++) {
    if (filters[i].operator !== ElasticOperatorsEnum.EXIST && filters[i].operator !== ElasticOperatorsEnum.DOES_NOT_EXIST) {
      if (!filters[i].value || filters[i].value === '' || filters[i].value.length === 0) {
        filters.splice(i, 1);
      }
    }
  }
  return filters;
}

/**
 * Merge default filter with params filter, if this a current filter will be override by params filter
 * @param filterToMerge Filter to merge in Merged
 * @param filterMerged Filter merged
 */
export function mergeParams(filterToMerge: ElasticFilterType[], filterMerged: ElasticFilterType[]): Promise<ElasticFilterType[]> {
  return new Promise<ElasticFilterType[]>(resolve => {
    for (const fil of filterToMerge) {
      const indexParam = filterMerged.findIndex(value => value.field === fil.field);
      if (indexParam === -1) {
        filterMerged.push(fil);
      } else {
        filterMerged[indexParam].value = fil.value;
        filterMerged[indexParam].operator = fil.operator;
      }
    }
    resolve(filterMerged);
  });
}

/**
 * Return inverse operator of a given Elastic filter
 * @param filter filter to invert
 */
export function invertFilter(filter: ElasticFilterType): ElasticFilterType {
  const operator: OperatorsType = FILTER_OPERATORS[FILTER_OPERATORS.findIndex(value => value.operator === filter.operator)];
  filter.operator = operator.inverse;
  return filter;
}
