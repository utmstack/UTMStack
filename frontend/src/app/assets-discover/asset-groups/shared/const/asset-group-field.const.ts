import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {AssetFieldFilterEnum} from '../../../shared/enums/asset-field-filter.enum';


export const ASSETS_GROUP_FIELDS_FILTERS: UtmFieldType[] = [
  {
    label: 'Source',
    field: AssetFieldFilterEnum.IP,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  {
    label: 'Type',
    field: AssetFieldFilterEnum.TYPE,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  {
    label: 'Data source',
    field: AssetFieldFilterEnum.OS,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
];

