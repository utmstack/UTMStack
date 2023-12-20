import {ElasticDataTypesEnum} from '../../../../../shared/enums/elastic-data-types.enum';
import {ElasticSearchFieldInfoType} from '../../../../../shared/types/elasticsearch/elastic-search-field-info.type';

export function filterFieldAgg(fields: ElasticSearchFieldInfoType[]): ElasticSearchFieldInfoType[] {
  return fields.filter(value => {
    if (value.type !== ElasticDataTypesEnum.TEXT) {
      return true;
    } else {
      return value.name.includes('.keyword');
    }
  });
}
