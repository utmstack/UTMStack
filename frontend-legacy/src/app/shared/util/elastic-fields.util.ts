import {ElasticDataTypesEnum} from '../enums/elastic-data-types.enum';
import {ElasticSearchFieldInfoType} from '../types/elasticsearch/elastic-search-field-info.type';

export function resolveIcon(field: ElasticSearchFieldInfoType): string {
  switch (field.type) {
    case ElasticDataTypesEnum.TEXT:
      return 'icon-text-color';
    case ElasticDataTypesEnum.LONG:
      return 'icon-hash';
    case ElasticDataTypesEnum.FLOAT:
      return 'icon-hash';
    case ElasticDataTypesEnum.DATE:
      return 'icon-calendar52';
    case ElasticDataTypesEnum.OBJECT:
      return 'icon-circle-css';
    case ElasticDataTypesEnum.BOOLEAN:
      return 'icon-split';
    default:
      return 'icon-question3';
  }
}
