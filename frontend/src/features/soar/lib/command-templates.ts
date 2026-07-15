import { isWindowsPlatform } from './alert-fields'

/** Starter snippets for the most common flow response actions, so a new flow
 *  isn't a blank line. Each resolves to real syntax for the agent's actual
 *  shell — cmd and PowerShell are different languages on Windows, not just a
 *  formatting choice — and, where a matching alert field actually exists,
 *  drops it in pre-filled with `$(field.path)`. Kill-process and the Windows
 *  logoff case have no backing alert field today (no PID/session-id anywhere
 *  on the alert document — see the soar backend), so those stay manual
 *  placeholders (`<PID>`, `<SESSION_ID>`) for the user to fill in. */
export type ShellKind = 'cmd' | 'powershell' | 'unix'

/** Windows without an explicit PowerShell shell falls back to cmd (matches
 *  defaultShellForPlatform); everything non-Windows shares one Unix syntax
 *  (bash and sh don't differ for these one-liners). */
export function shellKindFor(platform: string, shell: string): ShellKind {
  if (!isWindowsPlatform(platform)) return 'unix'
  return shell === 'powershell' ? 'powershell' : 'cmd'
}

export interface CommandTemplate {
  id: string
  command: (shell: ShellKind) => string
}

export const COMMAND_TEMPLATES: CommandTemplate[] = [
  {
    id: 'isolate',
    command: (shell) => {
      if (shell === 'powershell') {
        return 'New-NetFirewallRule -DisplayName "UTMStack-Isolate-Out" -Direction Outbound -Action Block -RemoteAddress Any; New-NetFirewallRule -DisplayName "UTMStack-Isolate-In" -Direction Inbound -Action Block -RemoteAddress Any'
      }
      if (shell === 'cmd') {
        return 'netsh advfirewall firewall add rule name="UTMStack-Isolate-Out" dir=out action=block remoteip=any && netsh advfirewall firewall add rule name="UTMStack-Isolate-In" dir=in action=block remoteip=any'
      }
      return 'iptables -A OUTPUT -j DROP && iptables -A INPUT -j DROP'
    },
  },
  {
    id: 'shutdown',
    command: (shell) => {
      if (shell === 'powershell') return 'Stop-Computer -Force'
      if (shell === 'cmd') return 'shutdown /s /f /t 0'
      return 'shutdown -h now'
    },
  },
  {
    id: 'killProcess',
    command: (shell) => {
      if (shell === 'powershell') return 'Stop-Process -Id <PID> -Force'
      if (shell === 'cmd') return 'taskkill /F /PID <PID>'
      return 'kill -9 <PID>'
    },
  },
  {
    id: 'logout',
    command: (shell) => {
      if (shell === 'powershell') return 'logoff.exe <SESSION_ID>'
      if (shell === 'cmd') return 'logoff <SESSION_ID>'
      return 'pkill -KILL -u $(target.user)'
    },
  },
  {
    id: 'disableUser',
    command: (shell) => {
      if (shell === 'powershell') return 'Disable-LocalUser -Name "$(target.user)"'
      if (shell === 'cmd') return 'net user "$(target.user)" /active:no'
      return 'usermod -L $(target.user)'
    },
  },
  {
    id: 'blockIp',
    command: (shell) => {
      if (shell === 'powershell') {
        return 'New-NetFirewallRule -DisplayName "Block-IP" -Direction Inbound -Action Block -RemoteAddress $(adversary.ip)'
      }
      if (shell === 'cmd') {
        return 'netsh advfirewall firewall add rule name="Block-IP" dir=in action=block remoteip=$(adversary.ip)'
      }
      return 'iptables -A INPUT -s $(adversary.ip) -j DROP'
    },
  },
]
