export interface Impact {
    confidentiality: number;
    integrity: number;
    availability: number;
}

export interface Variable {
    get: string;
    as: string;
    of_type: string;
}

export interface Definition {
    variables: Variable[];
    expression: string;
}

export interface DataType {
    id: number;
    name: string;
}

export interface Rule {
    id: number;
    data_types: DataType[];
    name: string;
    impact: Impact;
    category: string;
    technique: string;
    description: string;
    references: string[];
    definition: Definition;
}

export type Mode = 'ADD' | 'EDIT';
