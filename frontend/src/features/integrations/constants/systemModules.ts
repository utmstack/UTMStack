// Built-in (system) integration icons. The catalog row (utm_module) no longer
// carries the icon/description for system modules — the frontend owns them so we
// can (1) ship per-theme icons and (2) translate descriptions (see i18n
// integrations.modules.<MODULE_NAME>). Custom modules still use the table.
//
// `icon` is the light/default asset served at /integrations/<icon>; set `iconDark`
// when a dark-theme variant exists (falls back to `icon` otherwise).
export interface SystemModuleMeta {
  icon: string
  iconDark?: string
}

export const SYSTEM_MODULES: Record<string, SystemModuleMeta> = {
  WINDOWS_AGENT: { icon: 'windows.png' },
  LINUX_AGENT: { icon: 'linux.png', iconDark: 'linux-dark.png' },
  MACOS: { icon: 'macos.png', iconDark: 'macos-dark.png' },
  VMWARE: { icon: 'vmware.svg' },
  NETFLOW: { icon: 'netflow.svg' },
  AWS_IAM_USER: { icon: 'aws.png', iconDark: 'aws-dark.png' },
  AZURE: { icon: 'azure.svg' },
  O365: { icon: 'm365.svg' },
  GCP: { icon: 'gcp.svg' },
  KASPERSKY: { icon: 'kaspersky.svg' },
  ESET: { icon: 'eset.svg' },
  BITDEFENDER: { icon: 'bitdefender.svg' },
  SENTINEL_ONE: { icon: 'sentinelone.svg', iconDark: 'sentinelone-dark.svg' },
  SOPHOS: { icon: 'sophos-central.svg' },
  CROWDSTRIKE: { icon: 'crowdstrike.svg' },
  CISCO: { icon: 'cisco.svg' },
  MERAKI: { icon: 'cisco.svg' },
  FIRE_POWER: { icon: 'cisco.svg' },
  CISCO_SWITCH: { icon: 'cisco.svg' },
  FORTIGATE: { icon: 'fortinet.svg', iconDark: 'fortinet-dark.svg' },
  FORTIWEB: { icon: 'fortinet.svg', iconDark: 'fortinet-dark.svg' },
  SOPHOS_XG: { icon: 'sophos-firewall.svg' },
  PALO_ALTO: { icon: 'palo-alto.svg', iconDark: 'palo-alto-dark.svg' },
  SONIC_WALL: { icon: 'sonicwall.svg', iconDark: 'sonicwall-dark.svg' },
  PFSENSE: { icon: 'pfsense.svg', iconDark: 'pfsense-dark.svg' },
  MIKROTIK: { icon: 'mikrotik.svg', iconDark: 'mikrotik-dark.svg' },
  AIX: { icon: 'aix.svg' },
  AS_400: { icon: 'ibm-as-400.svg' },
  ORACLE: { icon: 'oracle.svg' },
  SURICATA: { icon: 'suricata.png', iconDark: 'suricata-dark.png' },
  DECEPTIVE_BYTES: { icon: 'deceptive-bytes.png' },
  GITHUB: { icon: 'github.png', iconDark: 'github-dark.png' },
  UTMSTACK: { icon: 'utmstack.svg' },
}
