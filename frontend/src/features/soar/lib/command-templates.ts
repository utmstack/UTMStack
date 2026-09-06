import { isWindowsPlatform } from './alert-fields'

/** Starter snippets for the most common flow response actions, so a new shell
 *  node isn't a blank line. Each resolves to real syntax for the agent's
 *  actual shell — cmd and PowerShell are different languages on Windows, not
 *  just a formatting choice — and, where a matching alert field actually
 *  exists, drops it in pre-filled with `$(alert.field.path)`. Kill-process
 *  and the Windows logoff case have no backing alert field today, so those
 *  stay manual placeholders (`<PID>`, `<SESSION_ID>`) for the user to fill in. */
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
  label: string
  command: (shell: ShellKind) => string
}

export const COMMAND_TEMPLATES: CommandTemplate[] = [
  {
    id: 'isolate',
    label: 'Isolate host (block all traffic)',
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
    label: 'Shutdown host',
    command: (shell) => {
      if (shell === 'powershell') return 'Stop-Computer -Force'
      if (shell === 'cmd') return 'shutdown /s /f /t 0'
      return 'shutdown -h now'
    },
  },
  {
    id: 'killProcess',
    label: 'Kill process by PID',
    command: (shell) => {
      if (shell === 'powershell') return 'Stop-Process -Id <PID> -Force'
      if (shell === 'cmd') return 'taskkill /F /PID <PID>'
      return 'kill -9 <PID>'
    },
  },
  {
    id: 'logout',
    label: 'Log user out',
    command: (shell) => {
      if (shell === 'powershell') return 'logoff.exe <SESSION_ID>'
      if (shell === 'cmd') return 'logoff <SESSION_ID>'
      return 'pkill -KILL -u $(alert.target.user)'
    },
  },
  {
    id: 'disableUser',
    label: 'Disable target user',
    command: (shell) => {
      if (shell === 'powershell') return 'Disable-LocalUser -Name "$(alert.target.user)"'
      if (shell === 'cmd') return 'net user "$(alert.target.user)" /active:no'
      return 'usermod -L $(alert.target.user)'
    },
  },
  {
    id: 'blockIp',
    label: 'Block adversary IP',
    command: (shell) => {
      if (shell === 'powershell') {
        return 'New-NetFirewallRule -DisplayName "Block-IP" -Direction Inbound -Action Block -RemoteAddress $(alert.adversary.ip)'
      }
      if (shell === 'cmd') {
        return 'netsh advfirewall firewall add rule name="Block-IP" dir=in action=block remoteip=$(alert.adversary.ip)'
      }
      return 'iptables -A INPUT -s $(alert.adversary.ip) -j DROP'
    },
  },
  {
    id: 'restart',
    label: 'Restart host',
    command: (shell) => {
      if (shell === 'powershell') return 'Restart-Computer -Force'
      if (shell === 'cmd') return 'shutdown /r /f /t 0'
      return 'shutdown -r now'
    },
  },
  {
    id: 'uninstall',
    label: 'Uninstall a package',
    command: (shell) => {
      if (shell === 'powershell') {
        return 'Get-Package -Name "<PACKAGE>" | Uninstall-Package -Force'
      }
      if (shell === 'cmd') return 'wmic product where name="<PACKAGE>" call uninstall /nointeractive'
      return 'apt-get remove -y "<PACKAGE>"'
    },
  },
]
