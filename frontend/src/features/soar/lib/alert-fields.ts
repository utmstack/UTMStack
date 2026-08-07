/** Alert fields a flow condition can match on.
 *
 *  These are matched against the alert the engine hands the SOAR plugin — the
 *  protobuf, not the row that is later stored — so the paths are the protobuf's.
 *  That is why there is no status here: a flow runs as the alert is raised,
 *  before anyone has triaged it, and status does not exist yet.
 *
 *  Anything the event carried rather than the alert lives under lastEvent. */
export interface AlertField {
  field: string
  label: string
}

export const ALERT_FIELDS: AlertField[] = [
  { field: 'name', label: 'Alert name' },
  { field: 'severity', label: 'Severity' },
  { field: 'category', label: 'Category' },
  { field: 'technique', label: 'Technique' },
  { field: 'dataType', label: 'Data type' },
  { field: 'dataSource', label: 'Data source' },
  { field: 'impactScore', label: 'Risk' },

  { field: 'adversary.ip', label: 'Adversary IP' },
  { field: 'adversary.user', label: 'Adversary user' },
  { field: 'adversary.host', label: 'Adversary host' },

  { field: 'target.ip', label: 'Target IP' },
  { field: 'target.user', label: 'Target user' },
  { field: 'target.host', label: 'Target host' },
  { field: 'target.port', label: 'Target port' },

  { field: 'lastEvent.protocol', label: 'Protocol' },
  { field: 'lastEvent.action', label: 'Action' },
  { field: 'lastEvent.origin.ip', label: 'Origin IP' },
  { field: 'lastEvent.origin.user', label: 'Origin user' },
  { field: 'lastEvent.origin.host', label: 'Origin host' },
  { field: 'lastEvent.origin.port', label: 'Origin port' },
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
