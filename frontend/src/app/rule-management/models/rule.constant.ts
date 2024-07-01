import {ElasticDataTypesEnum} from '../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../shared/types/table/utm-field.type';

export const RULE_NAME_FIELD = 'name';
export const RULE_DATA_TYPES = 'data_types';
export const RULE_CATEGORY = 'category';
export const RULE_TECHNIQUE = 'technique';
export const RULE_REFERENCES = 'references';

export const RULE_FIELDS: UtmFieldType[] = [
    {
        label: 'Rule name',
        field: RULE_NAME_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
        filter: false
    },
    {
        label: 'Types',
        field: RULE_DATA_TYPES,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true
    },
    {
        label: 'Category',
        field: RULE_CATEGORY,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true
    },
    {
        label: 'Technique',
        field: RULE_TECHNIQUE,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true
    },
    {
        label: 'References',
        field: RULE_REFERENCES,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true
    },
];

export const FILTER_RULE_FIELDS = RULE_FIELDS.filter(f => f.filter);
