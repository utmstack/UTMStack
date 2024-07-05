import {ElasticDataTypesEnum} from '../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../shared/types/table/utm-field.type';

export const RULE_NAME_FIELD = 'name';
export const RULE_DATA_TYPES = 'dataTypes';
export const RULE_CATEGORY = 'category';
export const RULE_TECHNIQUE = 'technique';
export const RULE_REFERENCES = 'references';

// FILTERS FIELDS

export const RULE_FILTER_NAME_FIELD = 'name';
export const RULE_FILTER_DATA_TYPES = 'RULE_DATA_TYPES';
export const RULE_FILTER_CATEGORY = 'RULE_CATEGORY';
export const RULE_FILTER_TECHNIQUE = 'RULE_TECHNIQUE';
export const RULE_FILTER_REFERENCES = 'references';


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
        filter: true,
        filterField: RULE_FILTER_DATA_TYPES
    },
    {
        label: 'Category',
        field: RULE_CATEGORY,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true,
        filterField: RULE_FILTER_CATEGORY
    },
    {
        label: 'Technique',
        field: RULE_TECHNIQUE,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true,
        filterField: RULE_FILTER_TECHNIQUE
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

export interface RuleFilterType {
    ruleName?: string;
    ruleConfidentiality?: number[];
    ruleIntegrity?: number[];
    ruleAvailability?: number[];
    ruleCategory?: string[];
    ruleTechnique?: string[];
    ruleInitDate?: Date;
    ruleEndDate?: Date;
    ruleActive?: boolean[];
    systemOwner?: boolean[];
    dataTypes?: number[];
    page: number;
    size: number;
    sort: string;
}

