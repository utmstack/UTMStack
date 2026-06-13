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
  ingestType?: string;
  isSystem: boolean;
}

export interface DataTypeOption {
  dataType: string;
  name: string;
  moduleName: string;
  active: boolean;
  isSystem: boolean;
}

export type DeployKind = 'agents & syslog' | 'device' | 'antivirus' | 'other' | 'custom' | 'cloud' | 'utmstack modules'
export type Status = 'configured' | 'available'
export type Tab = 'all' | 'agents' | 'collectors' | 'cloud' |'custom'

export interface Integration {
  id: string;
  name: string;
  moduleName?: string;
  dataType?: string;
  ingestType?: string;
  kind: DeployKind;
  status: Status;
  description: string;
  category: string;
  logo: string;
  logoDark?: string;      // per-theme icon for system modules (falls back to logo)
  isSystem?: boolean;     // false → custom (uses the catalog row's icon/description)
  darkInvert?: boolean;
  defaultPort?: string;
  cloudFields?: { label: string; placeholder: string; secret?: boolean }[];
  events24h?: number;
  rate?: number;
}

export interface TenantResponse {
  name: string;
  config: Record<string, string>;
}
