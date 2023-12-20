export class IncidentAssetType {
  id?: number;
  hostname?: string;
  mac?: string;
  ip?: string;
  filebeat?: string;
  winlogbeat?: string;
  hids?: string;
  agentVersion?: number;
  lastSeen?: Date;
  label?: string;
}
