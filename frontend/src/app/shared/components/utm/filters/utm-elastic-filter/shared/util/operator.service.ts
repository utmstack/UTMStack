import {Injectable} from '@angular/core';
import {FILTER_OPERATORS} from '../../../../../../constants/filter-operators.const';
import {ElasticDataTypesEnum} from '../../../../../../enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../../../enums/elastic-operators.enum';
import {ElasticSearchFieldInfoType} from '../../../../../../types/elasticsearch/elastic-search-field-info.type';
import {OperatorsType} from '../../../../../../types/filter/operators.type';

@Injectable({
  providedIn: 'root'
})
export class OperatorService {

  getOperators(field: ElasticSearchFieldInfoType, operators: OperatorsType[]) {
      if (field.type === ElasticDataTypesEnum.TEXT || field.type === ElasticDataTypesEnum.STRING ||
        field.type === ElasticDataTypesEnum.KEYWORD) {
        if (!field.name.includes('.keyword')) {
            operators = FILTER_OPERATORS.filter(value =>
            value.operator !== ElasticOperatorsEnum.IS_BETWEEN &&
            value.operator !== ElasticOperatorsEnum.IS_NOT_BETWEEN);
        } else {
            operators = FILTER_OPERATORS.filter(value =>
            value.operator !== ElasticOperatorsEnum.IS_BETWEEN &&
            value.operator !== ElasticOperatorsEnum.CONTAIN &&
            value.operator !== ElasticOperatorsEnum.DOES_NOT_CONTAIN &&
            value.operator !== ElasticOperatorsEnum.IS_NOT_BETWEEN &&
            value.operator !== ElasticOperatorsEnum.ENDS_WITH &&
            value.operator !== ElasticOperatorsEnum.NOT_ENDS_WITH &&
            value.operator !== ElasticOperatorsEnum.START_WITH &&
            value.operator !== ElasticOperatorsEnum.NOT_START_WITH);
        }

      } else if (field.type === ElasticDataTypesEnum.LONG ||
        field.type === ElasticDataTypesEnum.NUMBER || field.type === ElasticDataTypesEnum.DATE) {
          operators = FILTER_OPERATORS.filter(value =>
          value.operator !== ElasticOperatorsEnum.CONTAIN &&
          value.operator !== ElasticOperatorsEnum.DOES_NOT_CONTAIN &&
          value.operator !== ElasticOperatorsEnum.START_WITH &&
          value.operator !== ElasticOperatorsEnum.NOT_START_WITH);
      } else {
        operators = FILTER_OPERATORS;
      }

      return operators;
    }
}
