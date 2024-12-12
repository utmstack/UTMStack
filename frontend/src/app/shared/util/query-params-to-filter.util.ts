import {ElasticOperatorsEnum} from '../enums/elastic-operators.enum';
import {ElasticFilterType} from '../types/filter/elastic-filter.type';

/**
 * Take query params and return ElasticFiltersType object
 * @param queryParams Object params from route snapshot
 */
export function parseQueryParamsToFilter(queryParams: object): Promise<ElasticFilterType[]> {
  return new Promise<ElasticFilterType[]>(resolve => {
    const filters: ElasticFilterType[] = [];
    for (const key of Object.keys(queryParams)) {
      if (key !== 'patternId' && key !== 'indexPattern' && key !== 'dataNature' && key !== 'mode'
        && key !== 'queryId' && key !== 'queryName' && key !== 'alertType') {
        const index = filters.findIndex(filter => filter.field === key);
        if (index !== -1) {
          filters[index].value = getValue(queryParams[key]);
          filters[index].operator = getOperator(queryParams[key]);
        } else {
          filters.push({
            field: key,
            value: getValue(queryParams[key]),
            operator: getOperator(queryParams[key])
          });
        }
      }
    }
    resolve(filters);
  });
}

/**
 * Return operator to add to filter
 * @param param Value of the param in queryParams
 */
export function getOperator(param: string): ElasticOperatorsEnum {
  if (param.includes('->')) {
    return ElasticOperatorsEnum[param.split('->')[0]];
  } else {
    return ElasticOperatorsEnum.IS;
  }
}

/**
 * Return value based on operator used in param
 * @param param Value of the param in queryParams
 */
export function getValue(param: string) {
  if (param.includes('->')) {
    const operator = ElasticOperatorsEnum[param.split('->')[0]];
    let value: string | string[] = param.split('->')[1];
    if (ElasticOperatorsEnum.IS_BETWEEN === operator || ElasticOperatorsEnum.IS_NOT_BETWEEN === operator) {
      return param.split('->')[1].split(',');
    } else if (ElasticOperatorsEnum.IS_ONE_OF === operator || ElasticOperatorsEnum.IS_NOT_ONE_OF === operator) {
      if (typeof value === 'string') {
        value = value.split(',');
      }
      return value;
    } else if (ElasticOperatorsEnum.EXIST === operator || ElasticOperatorsEnum.DOES_NOT_EXIST === operator) {
      return null;
    } else {
      return value;
    }
  } else {
    return param;
  }
}

/**
 * Return string of query params
 * @param filters ElasticFilterType to convert to string params
 */
export function filtersToStringParam(filters: ElasticFilterType[]): Promise<string> {
  return new Promise<string>(resolve => {
    let queryString = '';
    /**
     * Add all filters to string
     */
    filters.forEach(value => {
      queryString += value.field + '=' + value.operator + '->' + value.value + '&';
    });
    // remove last &
    queryString = queryString.substring(0, queryString.length - 1);
    resolve(queryString);
  });
}

/**
 * Convert url string to query params to pass to Angualar route in extra data;
 * @param queryString Query string to convert to queryParams
 */
export function stringParamToQueryParams(queryString: string): Promise<object> {
  return new Promise<object>(resolve => {
    const queryParams = {};
    const params = queryString.split('&');
    for (const param of params) {
      const queryObject = param.split('=');
      queryParams[queryObject[0]] = queryObject[1];
    }
    resolve(queryParams);
  });
}

/**
 * Return string of query params
 * @param filters ElasticFilterType to convert to string params
 */
export function filtersWithPatternToStringParam(filters: ElasticFilterType[]): Promise<string> {
  return new Promise<string>(resolve => {
    let queryString = '';
    /**
     * Add all filters to string
     */
    filters.forEach(value => {
      if (value.pattern) {
        queryString += value.field + '=' + value.operator + '->' + value.value + '->' + value.pattern + '&';
      } else {
        queryString += value.field + '=' + value.operator + '->' + value.value + '&';
      }
    });
    // remove last &
    queryString = queryString.substring(0, queryString.length - 1);
    resolve(queryString);
  });
}

/**
 * Take query params and return ElasticFiltersType object
 * @param queryParams Object params from route snapshot
 */
export function parseQueryParamsToFilterWithPattern(queryParams: object): Promise<ElasticFilterType[]> {
  return new Promise<ElasticFilterType[]>(resolve => {
    const filters: ElasticFilterType[] = [];
    for (const key of Object.keys(queryParams)) {
      if (key !== 'patternId' && key !== 'indexPattern' && key !== 'dataNature' && key !== 'mode'
        && key !== 'queryId' && key !== 'queryName' && key !== 'alertType') {
        const index = filters.findIndex(filter => filter.field === key);
        if (index !== -1) {
          filters[index].value = getValue(queryParams[key]);
          filters[index].operator = getOperator(queryParams[key]);
          filters[index].pattern = getIndexPattern(queryParams[key]);
        } else {
          filters.push({
            field: key,
            value: getValue(queryParams[key]),
            operator: getOperator(queryParams[key]),
            pattern: getIndexPattern(queryParams[key])
          });
        }
      }
    }
    resolve(filters);
  });
}

/**
 * Return indexPattern to add to filter
 * @param param Value of the param in queryParams
 */
export function getIndexPattern(param: string) {
  if (param.includes('->')) {
    return param.split('->').length > 2 ? param.split('->')[2] : '';
  }
  return '';
}
