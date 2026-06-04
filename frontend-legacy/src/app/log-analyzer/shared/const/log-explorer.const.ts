import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';

export const LOG_EXPLORER_TIME_FILTER = {field: '@timestamp', operator: ElasticOperatorsEnum.IS_BETWEEN, value: ['now-1d', 'now']};

export const LOG_EXPLORER_TIME_DEFAULT = ['now-1d', 'now'];
