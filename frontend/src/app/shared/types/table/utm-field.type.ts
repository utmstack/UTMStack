import {ElasticDataTypesEnum} from '../../enums/elastic-data-types.enum';

export class UtmFieldType {
  label?: string;
  field?: string;
  filterField?: string;
  type?: ElasticDataTypesEnum;
  visible?: boolean;
  customStyle?: string;
  filter?: boolean;
  width?: string;
}
