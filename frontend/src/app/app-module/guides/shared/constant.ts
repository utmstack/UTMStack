const WINDOWS_SHELL =
  'Run the following PowerShell script as “ADMINISTRATOR” on a server with the UTMStack agent installed.';

const LINUX_SHELL =
  'Run the following Bash script as “ADMINISTRATOR” on a server with the UTMStack agent installed.';

export interface Platform {
  id: number;
  name: string;
  command: string;
  shell: string;
  path?: string;
  restart?: string;
  extraCommands?: string[];
}

function createPlatform(
  id: number,
  name: string,
  command: string,
  shell: string,
  path?: string,
  restart?: string,
  extraCommands?: string[]): Platform {
  return { id, name, command, shell, path, restart, extraCommands };
}

export const createPlatforms = (
  windowsCommandAMD64: string,
  windowsCommandARM64: string,
  linuxCommand: string,
  windowsPath?: string,
  windowsRestart?: string,
  linuxPath?: string,
  linuxRestart?: string): Platform[] => [
  createPlatform(
    1,
    'WINDOWS (AMD64)',
    windowsCommandAMD64,
    WINDOWS_SHELL,
    windowsPath,
    windowsRestart,[
      'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service.exe" ' +
      '-ArgumentList \'load-tls-certs\', \'[YOUR_CERT_PATH]\', \'[YOUR_KEY_PATH]\' ' +
      '-NoNewWindow -Wait'
    ]
  ),
  createPlatform(
    2,
    'WINDOWS (ARM64)',
    windowsCommandARM64,
    WINDOWS_SHELL,
    windowsPath,
    windowsRestart,
    [
      'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service_arm64.exe" ' +
      '-ArgumentList \'load-tls-certs\', \'[YOUR_CERT_PATH]\', \'[YOUR_KEY_PATH]\' ' +
      '-NoNewWindow -Wait'
    ]
  ),
  createPlatform(
    3,
    'LINUX',
    linuxCommand,
    LINUX_SHELL,
    linuxPath,
    linuxRestart,
    [
      `sudo bash -c "/opt/utmstack-linux-agent/utmstack_agent_service load-tls-certs [YOUR_CERT_PATH] [YOUR_KEY_PATH]"`
    ]
  )
];

export const createFileBeatsPlatforms = (
  windowsCommand: string,
  linuxCommand: string,
  windowsPath?: string,
  windowsRestart?: string,
  linuxPath?: string,
  linuxRestart?: string): Platform[] => [
  createPlatform(
    1,
    'WINDOWS',
    windowsCommand,
    WINDOWS_SHELL,
    windowsPath,
    windowsRestart
  ),
  createPlatform(
    3,
    'LINUX',
    linuxCommand,
    LINUX_SHELL,
    linuxPath,
    linuxRestart
  )
];

export const PLATFORMS = createPlatforms(
  'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service.exe" -ArgumentList \'ACTION\', \'AGENT_NAME\', \'PROTOCOL\', \'TLS\' -NoNewWindow -Wait\n',
  'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service_arm64.exe" -ArgumentList \'ACTION\', \'AGENT_NAME\', \'PROTOCOL\, \'TLS\' -NoNewWindow -Wait\n',
  'sudo bash -c "/opt/utmstack-linux-agent/utmstack_agent_service ACTION AGENT_NAME PROTOCOL TLS"'
);

export const FILEBEAT_PLATFORMS = createFileBeatsPlatforms(
  'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\"; Start-Process "filebeat.exe" -ArgumentList "modules", "enable", "AGENT_NAME"',
  'cd /opt/utmstack-linux-agent/beats/filebeat/ && ./filebeat modules enable AGENT_NAME',
  'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
  'Stop-Service -Name UTMStackModulesLogsCollector; Start-Sleep -Seconds 5; Start-Service -Name UTMStackModulesLogsCollector',
  '/opt/utmstack-linux-agent/beats/filebeat/modules.d/',
  'sudo systemctl restart UTMStackModulesLogsCollector'
);
