import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {AssetFieldFilterEnum} from '../../../shared/enums/asset-field-filter.enum';
import {CollectorFieldFilterEnum} from "../../../shared/enums/collector-field-filter.enum";


export const COLLECTORS_GROUP_FIELDS_FILTERS: UtmFieldType[] = [
  {
    label: 'Source',
    field: CollectorFieldFilterEnum.COLLECTOR_IP,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  }
];

