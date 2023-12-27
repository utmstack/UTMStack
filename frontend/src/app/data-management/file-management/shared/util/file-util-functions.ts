import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {FILE_FIELDS} from '../const/file-field.constant';

export function resolveFileFieldNameByFilter(filter: ElasticFilterType): string {
  const field = filter.field.replace('.keyword', '');
  const indexField = FILE_FIELDS.findIndex(value => value.field === field);
  if (indexField !== -1) {
    return FILE_FIELDS[indexField].label;
  } else {
    return field;
  }
}
