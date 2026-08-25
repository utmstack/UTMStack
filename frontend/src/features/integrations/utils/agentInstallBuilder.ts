export type AgentAction = 'install' | 'uninstall'

export interface AgentActionOption {
  id: AgentAction
  labelKey: string
}

export interface AgentPlatform {
  id: string
  name: string
  installerName: string
  flavor?: 'ubuntu' | 'fedora'
}

export interface AgentInstallConfig {
  actions: AgentActionOption[]
  platforms: AgentPlatform[]
  getCommand: (action: AgentAction, platformId: string) => string
}

export interface AgentBuilderContext {
  host: string
  token: string
}

const ACTIONS: AgentActionOption[] = [
  { id: 'install', labelKey: 'integrations.setup.agent.actions.install' },
  { id: 'uninstall', labelKey: 'integrations.setup.agent.actions.uninstall' },
]

const LINUX_AMD64 = 'utmstack_agent_service_linux_amd64'
const LINUX_ARM64 = 'utmstack_agent_service_linux_arm64'
const DARWIN_ARM64 = 'utmstack_agent_service_darwin_arm64'
const WIN_AMD64 = 'utmstack_agent_service_windows_amd64.exe'
const WIN_ARM64 = 'utmstack_agent_service_windows_arm64.exe'

const LINUX_PLATFORMS: AgentPlatform[] = [
  { id: 'ubuntu-amd64', name: 'Ubuntu / Debian (AMD64)', installerName: LINUX_AMD64, flavor: 'ubuntu' },
  { id: 'ubuntu-arm64', name: 'Ubuntu / Debian (ARM64)', installerName: LINUX_ARM64, flavor: 'ubuntu' },
  { id: 'fedora-amd64', name: 'Fedora / RedHat (AMD64)', installerName: LINUX_AMD64, flavor: 'fedora' },
  { id: 'fedora-arm64', name: 'Fedora / RedHat (ARM64)', installerName: LINUX_ARM64, flavor: 'fedora' },
]

const MACOS_PLATFORMS: AgentPlatform[] = [
  { id: 'darwin-arm64', name: 'ARM64 (Apple Silicon)', installerName: DARWIN_ARM64 },
]

const WINDOWS_PLATFORMS: AgentPlatform[] = [
  { id: 'windows-amd64', name: 'AMD64', installerName: WIN_AMD64 },
  { id: 'windows-arm64', name: 'ARM64', installerName: WIN_ARM64 },
]

function linuxInstall(host: string, token: string, installer: string, flavor: 'ubuntu' | 'fedora'): string {
  const pkgInstall = flavor === 'ubuntu'
    ? 'apt update -y && apt install wget -y'
    : 'yum install wget -y'

  return `sudo bash -c "${pkgInstall} && mkdir -p /opt/utmstack-linux-agent && \\
    wget --no-check-certificate -P /opt/utmstack-linux-agent \\
    https://${host}/private/dependencies/agent/${installer} && \\
    chmod -R 755 /opt/utmstack-linux-agent/${installer} && \\
    /opt/utmstack-linux-agent/${installer} install ${host} <secret>${token}</secret> yes"`
}

function linuxUninstall(installer: string): string {
  return `sudo bash -c "/opt/utmstack-linux-agent/${installer} uninstall || true; \\
    systemctl stop UTMStackAgent 2>/dev/null || true; systemctl disable UTMStackAgent 2>/dev/null || true; \\
    rm -f /etc/systemd/system/UTMStackAgent.service 2>/dev/null || true; \\
    systemctl stop UTMStackModulesLogsCollector 2>/dev/null || true; \\
    systemctl disable UTMStackModulesLogsCollector 2>/dev/null || true; \\
    rm -f /etc/systemd/system/UTMStackModulesLogsCollector.service 2>/dev/null || true; \\
    systemctl daemon-reload 2>/dev/null || true; \\
    echo 'Removing UTMStack Agent dependencies...' && sleep 10 && rm -rf /opt/utmstack-linux-agent 2>/dev/null || true; \\
    echo 'UTMStack Agent dependencies removed successfully.'"`
}

function macosInstall(host: string, token: string, installer: string): string {
  return `sudo bash -c "mkdir -p /opt/utmstack-macos-agent && \\
    curl -k -o /opt/utmstack-macos-agent/${installer} \\
    https://${host}/private/dependencies/agent/${installer} && \\
    chmod +x /opt/utmstack-macos-agent/${installer} && \\
    /opt/utmstack-macos-agent/${installer} install ${host} <secret>${token}</secret> yes"`
}

function macosUninstall(installer: string): string {
  return `sudo bash -c "/opt/utmstack-macos-agent/${installer} uninstall || true; \\
    launchctl bootout system /Library/LaunchDaemons/UTMStackAgent.plist 2>/dev/null || true; \\
    launchctl bootout system /Library/LaunchDaemons/UTMStackUpdater.plist 2>/dev/null || true; \\
    rm -f /Library/LaunchDaemons/UTMStackAgent.plist 2>/dev/null || true; \\
    rm -f /Library/LaunchDaemons/UTMStackUpdater.plist 2>/dev/null || true; \\
    echo 'Removing UTMStack Agent dependencies...' && sleep 10 && \\
    rm -rf /opt/utmstack-macos-agent 2>/dev/null || true; \\
    echo 'UTMStack Agent removed successfully.'"`
}

function windowsInstall(host: string, token: string, installer: string): string {
  const dir = 'C:\\Program Files\\UTMStack\\UTMStack Agent'
  // One statement per line (backtick = PowerShell line continuation) so the block
  // wraps top-to-bottom like the bash guides instead of one long horizontal line.
  return `New-Item -ItemType Directory -Force -Path "${dir}"
& curl.exe -k -o "${dir}\\${installer}" \`
  "https://${host}/private/dependencies/agent/${installer}"
Start-Process "${dir}\\${installer}" \`
  -ArgumentList 'install', '${host}', '<secret>${token}</secret>', 'yes' \`
  -NoNewWindow -Wait`
}

function windowsUninstall(installer: string): string {
  const dir = 'C:\\Program Files\\UTMStack\\UTMStack Agent'
  // One statement per line so the block flows top-to-bottom (no horizontal scroll).
  return `Start-Process "${dir}\\${installer}" -ArgumentList 'uninstall' -NoNewWindow -Wait -ErrorAction SilentlyContinue | Out-Null
Start-Process -FilePath "sc.exe" -ArgumentList 'stop','UTMStackAgent' -Wait -ErrorAction SilentlyContinue | Out-Null
Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackAgent' -Wait -ErrorAction SilentlyContinue | Out-Null
Start-Process -FilePath "sc.exe" -ArgumentList 'stop','UTMStackWindowsLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null
Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackWindowsLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null
Start-Process -FilePath "sc.exe" -ArgumentList 'stop','UTMStackModulesLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null
Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackModulesLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null
Write-Host "Removing UTMStack Agent dependencies..."
Start-Sleep -Seconds 10
Remove-Item '${dir}' -Recurse -Force -ErrorAction Stop
Write-Host "UTMStack Agent removed successfully."`
}

export function buildAgentInstall(agentId: string, ctx: AgentBuilderContext): AgentInstallConfig {
  const { host, token } = ctx

  if (agentId.toLowerCase().startsWith('linux')) {
    return {
      actions: ACTIONS,
      platforms: LINUX_PLATFORMS,
      getCommand: (action, platformId) => {
        const platform = LINUX_PLATFORMS.find(p => p.id === platformId)
        if (!platform) return ''
        return action === 'install'
          ? linuxInstall(host, token, platform.installerName, platform.flavor ?? 'ubuntu')
          : linuxUninstall(platform.installerName)
      },
    }
  }

  if (agentId.toLowerCase().startsWith('macos')) {
    return {
      actions: ACTIONS,
      platforms: MACOS_PLATFORMS,
      getCommand: (action, platformId) => {
        const platform = MACOS_PLATFORMS.find(p => p.id === platformId)
        if (!platform) return ''
        return action === 'install'
          ? macosInstall(host, token, platform.installerName)
          : macosUninstall(platform.installerName)
      },
    }
  }

  if (agentId.toLowerCase().startsWith('windows')) {
    return {
      actions: ACTIONS,
      platforms: WINDOWS_PLATFORMS,
      getCommand: (action, platformId) => {
        const platform = WINDOWS_PLATFORMS.find(p => p.id === platformId)
        if (!platform) return ''
        return action === 'install'
          ? windowsInstall(host, token, platform.installerName)
          : windowsUninstall(platform.installerName)
      },
    }
  }

  return {
    actions: ACTIONS,
    platforms: [],
    getCommand: () => '',
  }
}
