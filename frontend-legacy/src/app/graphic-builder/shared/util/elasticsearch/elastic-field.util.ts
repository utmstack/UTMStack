import {Observable} from 'rxjs';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';

/**
 * Extract field based on data type
 * @param fields Fields of index
 * @param type Field type to filter
 */
export function filterFieldDataType(fields: ElasticSearchFieldInfoType[], type: string | string[]):
  Observable<ElasticSearchFieldInfoType[]> {
  return new Observable<ElasticSearchFieldInfoType[]>(emit => {
    emit.next(fields.filter(value => type.includes(value.type)));
  });
}
