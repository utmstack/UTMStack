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
    dataTypeDescription?: string;
    lastUpdate?: string;
    systemOwner: boolean;
}

export interface RegexPattern {
    id: number;
    patternId: string;
    patternDescription?: string;
    patternDefinition: string;
    systemOwner: boolean;
    lastUpdate?: string;
}


export interface Rule {
    id: number;
    dataTypes: DataType[];
    name: string;
    confidentiality: number;
    integrity: number;
    availability: number;
    category: string;
    technique: string;
    description: string;
    references: string[];
    definition: Definition;
}

export type Mode = 'ADD' | 'EDIT';
