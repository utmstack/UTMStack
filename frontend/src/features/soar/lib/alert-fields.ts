/** Alert fields a flow condition can match on (mirrors the legacy
 *  INCIDENT_AUTOMATION_ALERT_FIELDS list). field = the dotted path, label = UI. */
export interface AlertField {
  field: string
  label: string
}

export const ALERT_FIELDS: AlertField[] = [
  { field: 'name', label: 'Alert name' },
  { field: 'severityLabel', label: 'Severity' },
  { field: 'category', label: 'Category' },
  { field: 'statusLabel', label: 'Status' },
  { field: 'dataType', label: 'Data type' },
  { field: 'dataSource', label: 'Data source' },
  { field: 'tactic', label: 'Tactic' },
  { field: 'technique', label: 'Technique' },
  { field: 'protocol', label: 'Protocol' },
  { field: 'action', label: 'Action' },
  { field: 'origin.ip', label: 'Origin IP' },
  { field: 'origin.user', label: 'Origin user' },
  { field: 'origin.host', label: 'Origin host' },
  { field: 'origin.port', label: 'Origin port' },
  { field: 'target.ip', label: 'Target IP' },
  { field: 'target.user', label: 'Target user' },
  { field: 'target.host', label: 'Target host' },
  { field: 'target.port', label: 'Target port' },
  { field: 'adversary.ip', label: 'Adversary IP' },
  { field: 'adversary.user', label: 'Adversary user' },
  { field: 'adversary.host', label: 'Adversary host' },
]

const WINDOWS_SHELLS = ['cmd', 'powershell']
const UNIX_SHELLS = ['bash', 'sh']

export const COMMON_PLATFORMS = ['windows', 'linux', 'ubuntu', 'centos', 'darwin']

export function isWindowsPlatform(platform: string): boolean {
  return platform.toLowerCase().includes('win')
}

/** Shells valid for the selected platform (windows → cmd/powershell; else bash/sh). */
export function shellsForPlatform(platform: string): string[] {
  return isWindowsPlatform(platform) ? WINDOWS_SHELLS : UNIX_SHELLS
}

export function defaultShellForPlatform(platform: string): string {
  return isWindowsPlatform(platform) ? 'cmd' : 'bash'
}
