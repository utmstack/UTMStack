// Built-in (system) integration presentation. The catalog row (utm_module) no longer
// carries the icon/description/category for system modules — the frontend owns them so we
// can (1) ship per-theme icons and (2) translate descriptions (see i18n
// integrations.modules.<MODULE_NAME>). Custom modules still use the table.
//
// `icon` is the light/default asset served at /integrations/<icon>; set `iconDark`
// when a dark-theme variant exists (falls back to `icon` otherwise).
export interface SystemModuleMeta {
  icon: string
  iconDark?: string
  /** Grid grouping. The catalog row no longer carries it — see the note above. */
  category: string
}

export const SYSTEM_MODULES: Record<string, SystemModuleMeta> = {
  WINDOWS_AGENT: { icon: 'windows.png', category: 'operating-system' },
  LINUX_AGENT: { icon: 'linux.png', iconDark: 'linux-dark.png', category: 'operating-system' },
  MACOS: { icon: 'macos.png', iconDark: 'macos-dark.png', category: 'operating-system' },
  VMWARE: { icon: 'vmware.svg', category: 'virtualization' },
  NETFLOW: { icon: 'netflow.svg', category: 'network' },
  AWS_IAM_USER: { icon: 'aws.png', iconDark: 'aws-dark.png', category: 'cloud' },
  AZURE: { icon: 'azure.svg', category: 'cloud' },
  O365: { icon: 'm365.svg', category: 'cloud' },
  GCP: { icon: 'gcp.svg', category: 'cloud' },
  KASPERSKY: { icon: 'kaspersky.svg', category: 'antivirus' },
  ESET: { icon: 'eset.svg', category: 'antivirus' },
  BITDEFENDER: { icon: 'bitdefender.svg', category: 'antivirus' },
  SENTINEL_ONE: { icon: 'sentinelone.svg', iconDark: 'sentinelone-dark.svg', category: 'xdr' },
  SOPHOS: { icon: 'sophos-central.svg', category: 'xdr' },
  CROWDSTRIKE: { icon: 'crowdstrike.svg', category: 'xdr' },
  CISCO: { icon: 'cisco.svg', category: 'firewall' },
  MERAKI: { icon: 'cisco.svg', category: 'firewall' },
  FIRE_POWER: { icon: 'cisco.svg', category: 'firewall' },
  CISCO_SWITCH: { icon: 'cisco.svg', category: 'network' },
  FORTIGATE: { icon: 'fortinet.svg', iconDark: 'fortinet-dark.svg', category: 'firewall' },
  FORTIWEB: { icon: 'fortinet.svg', iconDark: 'fortinet-dark.svg', category: 'firewall' },
  SOPHOS_XG: { icon: 'sophos-firewall.svg', category: 'firewall' },
  PALO_ALTO: { icon: 'palo-alto.svg', iconDark: 'palo-alto-dark.svg', category: 'firewall' },
  SONIC_WALL: { icon: 'sonicwall.svg', iconDark: 'sonicwall-dark.svg', category: 'firewall' },
  PFSENSE: { icon: 'pfsense.svg', iconDark: 'pfsense-dark.svg', category: 'firewall' },
  MIKROTIK: { icon: 'mikrotik.svg', iconDark: 'mikrotik-dark.svg', category: 'firewall' },
  AIX: { icon: 'aix.svg', category: 'operating-system' },
  AS_400: { icon: 'ibm-as-400.svg', category: 'operating-system' },
  ORACLE: { icon: 'oracle.svg', category: 'database' },
  SURICATA: { icon: 'suricata.png', iconDark: 'suricata-dark.png', category: 'ids' },
  DECEPTIVE_BYTES: { icon: 'deceptive-bytes.png', category: 'deception' },
  GITHUB: { icon: 'github.png', iconDark: 'github-dark.png', category: 'devops' },
  UTMSTACK: { icon: 'utmstack.svg', category: 'siem' },
}
