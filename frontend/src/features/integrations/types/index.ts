export interface ModuleActivationRequest {
  moduleName: string;
  activationStatus: boolean;
}

export interface CreateModuleRequest {
  moduleName: string;
  dataType: string;
  prettyName?: string;
  moduleDescription?: string;
  moduleIcon?: string;
  moduleCategory?: string;
}

export interface UpdateModuleRequest {
  prettyName?: string;
  moduleDescription?: string;
  moduleIcon?: string;
  moduleCategory?: string;
}

export interface ModuleResponse {
  id: number;
  dataType?: string;
  moduleName: string;
  prettyName?: string;
  moduleDescription?: string;
  moduleActive: boolean;
  moduleIcon?: string;
  moduleCategory?: string;
  isSystem: boolean;
}

export interface DataTypeOption {
  dataType: string;
  name: string;
  moduleName: string;
  active: boolean;
  isSystem: boolean;
}
