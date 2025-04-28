export const RULE_REQUEST: {size: number, page: number, sort: string, search?: string | number} = {
  size: 25,
  page: 0,
  sort: 'ruleLastUpdate,DESC',
};
export interface Impact {
    confidentiality: number;
    integrity: number;
    availability: number;
}
export const itemsPerPage = 25;

export interface Variable {
    get: string;
    as: string;
    of_type: string;
}

export interface Definition {
    ruleVariables: Variable[];
    ruleExpression: string;
}

export interface DataType {
    id: number;
    dataType: string;
    dataTypeName: string;
    dataTypeDescription?: string;
    lastUpdate?: string;
    systemOwner: boolean;
    included: boolean;
}

export interface RegexPattern {
    id: number;
    patternId: string;
    patternDescription?: string;
    patternDefinition: string;
    systemOwner: boolean;
    lastUpdate?: string;
}

export interface Asset {
  id?: number;
  assetName: string;
  assetHostnameList?: string[];
  assetIpList?: string[];
  assetConfidentiality: number;
  assetIntegrity: number;
  assetAvailability: number;
  lastUpdate?: string;
}



export interface Rule {
    id: number;
    dataTypes: DataType[];
    name: string;
    adversary?: string;
    confidentiality: number;
    integrity: number;
    availability: number;
    category: string;
    technique: string;
    description: string;
    references: string[];
    definition: Definition;
    systemOwner: boolean;
    ruleActive: boolean;
    valid?: boolean;
    showDetail?: boolean;
    status?: Status;
    isLoading: boolean;
    afterEvents: SearchRequest[];
    deduplicateBy?: string[];
}

export interface Expression {
  field: string;
  operator: string;
  value: any;
}

export interface SearchRequest {
  indexPattern: string;
  with: Expression[];
  or?: SearchRequest[];
  within?: string;
  count?: number;
}

export type Mode = 'ADD' | 'EDIT' | 'IMPORT' | 'ERROR';
export type Status = 'valid' | 'saved' | 'error';

export enum AddRuleStepEnum {
  STEP0,
  STEP1,
  STEP2,
  STEP3,
  STEP4,
}
