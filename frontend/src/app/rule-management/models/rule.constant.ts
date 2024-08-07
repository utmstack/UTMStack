import {ElasticDataTypesEnum} from '../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../shared/types/table/utm-field.type';

export const RULE_NAME_FIELD = 'name';
export const RULE_DATA_TYPES = 'dataTypes';
export const RULE_CATEGORY = 'category';
export const RULE_TECHNIQUE = 'technique';
export const RULE_CONFIDENTIALITY = 'confidentiality';
export const RULE_INTEGRITY = 'integrity';
export const RULE_AVAILABILITY = 'availability';

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
        filter: false,
        width: '15%'
    },
    {
        label: 'Types',
        field: RULE_DATA_TYPES,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true,
        filterField: RULE_FILTER_DATA_TYPES,
        width: '30%'
    },
    {
        label: 'Category',
        field: RULE_CATEGORY,
        type: ElasticDataTypesEnum.OBJECT,
        visible: true,
        filter: true,
        filterField: RULE_FILTER_CATEGORY,
        width: '10%'
    },
    {
        label: 'Technique',
        field: RULE_TECHNIQUE,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
        filter: true,
        filterField: RULE_FILTER_TECHNIQUE,
        width: '10%'
    },
    {
        label: 'Confidentiality',
        field: RULE_CONFIDENTIALITY,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
        filter: false,
        width: '10%'
    },
    {
      label: 'Integrity',
      field: RULE_INTEGRITY,
      type: ElasticDataTypesEnum.STRING,
      visible: true,
      filter: false,
      width: '10%'
    },
    {
      label: 'Availability',
      field: RULE_AVAILABILITY,
      type: ElasticDataTypesEnum.STRING,
      visible: true,
      filter: false,
      width: '10%'
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

export const VariableDataType = [
  { label: 'String', value: 'cel.StringType' },
  { label: 'Integer', value: 'cel.IntType' },
  { label: 'Double', value: 'cel.DoubleType' },
  { label: 'Boolean', value: 'cel.BoolType' },
  { label: 'Bytes', value: 'cel.BytesType' },
  { label: 'Unsigned Integer', value: 'cel.UintType' },
  { label: 'Timestamp', value: 'cel.TimestampType' },
  { label: 'Duration', value: 'cel.DurationType' },
  { label: 'Type', value: 'cel.TypeType' },
  { label: 'Null', value: 'cel.NullType' },
  { label: 'Any', value: 'cel.AnyType' }
];
