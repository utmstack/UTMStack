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

export interface Rule {
    id: number;
    data_types: string[];
    name: string;
    impact: Impact;
    category: string;
    technique: string;
    description: string;
    references: string[];
    definition: Definition;
}
