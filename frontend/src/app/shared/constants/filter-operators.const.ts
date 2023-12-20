import {ElasticOperatorsEnum} from '../enums/elastic-operators.enum';
import {OperatorsType} from '../types/filter/operators.type';

export const FILTER_OPERATORS: OperatorsType[] = [
  {name: 'is', operator: ElasticOperatorsEnum.IS, inverse: ElasticOperatorsEnum.IS_NOT},
  {name: 'is not', operator: ElasticOperatorsEnum.IS_NOT, inverse: ElasticOperatorsEnum.IS},
  {name: 'is one of', operator: ElasticOperatorsEnum.IS_ONE_OF, inverse: ElasticOperatorsEnum.IS_NOT_ONE_OF},
  {name: 'is not one of', operator: ElasticOperatorsEnum.IS_NOT_ONE_OF, inverse: ElasticOperatorsEnum.IS_ONE_OF},
  {name: 'exists', operator: ElasticOperatorsEnum.EXIST, inverse: ElasticOperatorsEnum.DOES_NOT_EXIST},
  {name: 'does not exist', operator: ElasticOperatorsEnum.DOES_NOT_EXIST, inverse: ElasticOperatorsEnum.EXIST},
  {name: 'is between', operator: ElasticOperatorsEnum.IS_BETWEEN, inverse: ElasticOperatorsEnum.IS_NOT_BETWEEN},
  {name: 'is not between', operator: ElasticOperatorsEnum.IS_NOT_BETWEEN, inverse: ElasticOperatorsEnum.IS_BETWEEN},
  {name: 'contain', operator: ElasticOperatorsEnum.CONTAIN, inverse: ElasticOperatorsEnum.DOES_NOT_CONTAIN},
  {name: 'does not contain', operator: ElasticOperatorsEnum.DOES_NOT_CONTAIN, inverse: ElasticOperatorsEnum.CONTAIN},
  {name: 'start with', operator: ElasticOperatorsEnum.START_WITH, inverse: ElasticOperatorsEnum.NOT_START_WITH},
  {name: 'does not start with', operator: ElasticOperatorsEnum.NOT_START_WITH, inverse: ElasticOperatorsEnum.START_WITH},
  {name: 'end with', operator: ElasticOperatorsEnum.ENDS_WITH, inverse: ElasticOperatorsEnum.NOT_ENDS_WITH},
  {name: 'does not end with', operator: ElasticOperatorsEnum.NOT_ENDS_WITH, inverse: ElasticOperatorsEnum.ENDS_WITH},
  // {name: 'fields contain', operator: ElasticOperatorsEnum.IS_IN_FIELD, inverse: ElasticOperatorsEnum.IS_NOT_FIELD},
  // {name: 'fields does not contain', operator: ElasticOperatorsEnum.IS_NOT_FIELD, inverse: ElasticOperatorsEnum.IS_IN_FIELD}
];
