export type IngestType = 'agent' | 'collector' | 'forwarder' | 'plugin'

export interface CreateIntegrationRequest {
  name: string;
  dataType: string;
  ingestType: IngestType;
  description?: string;
  icon?: string;
}

export interface UpdateIntegrationRequest {
  description?: string;
  icon?: string;
}

export interface IntegrationResponse {
  id: string;
  name: string;
  dataType?: string;
  ingestType?: IngestType;
  description?: string;
  icon?: string;
  systemOwner: boolean;
  /** Whether this tenant has any configuration group for it (pullers only). */
  configured: boolean;
}

export interface DataTypeOption {
  dataType: string;
  name: string;
  systemOwner: boolean;
}

export type DeployKind = 'agents & syslog' | 'device' | 'antivirus' | 'other' | 'custom' | 'cloud' | 'utmstack modules'
export type Status = 'configured' | 'available'
export type Tab = 'all' | 'agents' | 'collectors' | 'cloud' |'custom'

export interface Integration {
  id: string;
  /** Display label, translated (i18n integrations.modules.<moduleName>). */
  name: string;
  /** The catalog identity the API is addressed by, e.g. AWS_IAM_USER. */
  moduleName?: string;
  dataType?: string;
  ingestType?: string;
  kind: DeployKind;
  status: Status;
  description: string;
  category: string;
  logo: string;
  logoDark?: string;      // per-theme icon for system modules (falls back to logo)
  systemOwner?: boolean;  // false → custom (uses the catalog row's icon/description)
  darkInvert?: boolean;
  defaultPort?: string;
  cloudFields?: { label: string; placeholder: string; secret?: boolean }[];
  events24h?: number;
  rate?: number;
}

export interface ConfigGroupResponse {
  name: string;
  description?: string;
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
