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
  { label: 'String', value: 'string' },
  { label: 'Integer', value: 'int' },
  { label: 'Double', value: 'double' },
  { label: 'Boolean', value: 'bool' },
  { label: 'Bytes', value: 'bytes' },
  { label: 'Unsigned Integer', value: 'uint' },
  { label: 'Timestamp', value: 'timestamp' },
  { label: 'Duration', value: 'duration' },
  { label: 'Type', value: 'type' },
  { label: 'Null', value: 'null' },
  { label: 'Any', value: 'any' },
  { label: 'Array of string', value: '[]string' },
  { label: 'Array of integer', value: '[]int' },
  { label: 'Array of double', value: '[]double' },
  { label: 'Array of boolean', value: '[]bool' },
  { label: 'Array of bytes', value: '[]bytes' },
  { label: 'Array of unsigned integer', value: '[]uint' },
  { label: 'Array of timestamp', value: '[]timestamp' },
  { label: 'Array of duration', value: '[]duration' },
  { label: 'Array of type', value: '[]type' },
  { label: 'Array of null', value: '[]null' },
  { label: 'Array of any', value: '[]any' },
  { label: 'Map: String, String', value: 'map[string]string' },
  { label: 'Map: String, Integer', value: 'map[string]int' },
  { label: 'Map: String, Double', value: 'map[string]double' },
  { label: 'Map: String, Boolean', value: 'map[string]bool' },
  { label: 'Map: String, Bytes', value: 'map[string]bytes' },
  { label: 'Map: String, Unsigned Integer', value: 'map[string]uint' },
  { label: 'Map: String, Timestamp', value: 'map[string]timestamp' },
  { label: 'Map: String, Duration', value: 'map[string]duration' },
  { label: 'Map: String, Type', value: 'map[string]type' },
  { label: 'Map: String, Null', value: 'map[string]null' },
  { label: 'Map: String, Any', value: 'map[string]any' }
];

