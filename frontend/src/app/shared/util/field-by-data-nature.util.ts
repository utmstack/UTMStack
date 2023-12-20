import {DataNatureTypeEnum, NatureDataPrefixEnum} from '../enums/nature-data.enum';
import {ElasticSearchFieldInfoType} from '../types/elasticsearch/elastic-search-field-info.type';

/**
 *
 * @param fields Field of type ElasticsearchFieldInfoType[] that want filter by data data nature
 * @param type Type of data nature EVENT,ALERT,VULNERABILITY
 */
export function fieldByDataNature(fields: ElasticSearchFieldInfoType[], type: DataNatureTypeEnum): ElasticSearchFieldInfoType[] {
  switch (type) {
    case DataNatureTypeEnum.ALERT:
      return fields.filter(value => value.name.startsWith(NatureDataPrefixEnum.ALERT) ||
        value.name.startsWith(NatureDataPrefixEnum.CORRELATION) ||
        value.name.startsWith(NatureDataPrefixEnum.GLOBAL) ||
        value.name.startsWith(NatureDataPrefixEnum.TIMESTAMP));
    case DataNatureTypeEnum.EVENT:
      return fields.filter(value => !(value.name.startsWith(NatureDataPrefixEnum.ALERT) ||
        value.name.startsWith(NatureDataPrefixEnum.CORRELATION) ||
        value.name.startsWith(NatureDataPrefixEnum.VULNERABILITY)) ||
        value.name.startsWith(NatureDataPrefixEnum.GLOBAL) ||
        value.name.startsWith(NatureDataPrefixEnum.TIMESTAMP));
    case DataNatureTypeEnum.VULNERABILITY:
      return fields.filter(value => value.name.startsWith(NatureDataPrefixEnum.VULNERABILITY) ||
        value.name.startsWith(NatureDataPrefixEnum.GLOBAL) ||
        value.name.startsWith(NatureDataPrefixEnum.TIMESTAMP));
  }
}
