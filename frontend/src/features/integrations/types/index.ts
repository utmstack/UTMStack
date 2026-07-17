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

// ── Remote collector (Forwarder) integration control ────────────────────────
// Mirrors backend/modules/integrations/dto/collector.go.

export interface ForwarderCollector {
  id: number;
  hostname: string;
  ip: string;
  version: string;
  lastSeen?: string;
  status: 'online' | 'offline';
}

export interface SetDataTypeConfigRequest {
  enabled: boolean;
  proto: string;
  port?: string;
  tls?: boolean;
  auth?: string;
  path?: string;
  signatureHeader?: string;
}

export interface ConfigKnowledgeResponse {
  accepted: boolean;
  requestId?: string;
  errorMessage?: string;
  /** Only set once, the first time a bearer/hmac token is generated. */
  generatedSecret?: string;
}

export interface GetDataTypeConfigResponse {
  configured: boolean;
  enabled?: boolean;
  proto?: string;
  port?: string;
  tls?: boolean;
  auth?: string;
  path?: string;
  signatureHeader?: string;
  configStatus?: string;
  lastError?: string;
}

export interface SetForwarderCertificatesRequest {
  certPem: string;
  keyPem: string;
  caPem?: string;
}

export interface TLSStatusResponse {
  available: boolean;
  certExists: boolean;
  keyExists: boolean;
  caExists: boolean;
  valid: boolean;
  error?: string;
}
