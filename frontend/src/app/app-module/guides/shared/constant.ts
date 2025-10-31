export interface Platform {
  id: number;
  name: string;
  command: string;
  shell: string;
  path: string;
  restart: string;
}

export const createPlatforms = (windowsCommandAMD64: string,
                                windowsCommandARM64: string,
                                linuxCommand: string,
                                windowsPath?: string,
                                windowsRestart?: string,
                                linuxPath?: string,
                                linuxRestart?: string) => [
  {
    id: 1,
    name: 'WINDOWS (AMD64)',
    command: windowsCommandAMD64,
    shell: 'Run the following powershell script as “ADMINISTRATOR” in a Server with the UTMStack agent Installed.',
    path: windowsPath,
    restart: windowsRestart
  },
  {
    id: 2,
    name: 'WINDOWS (ARM64)',
    command: windowsCommandARM64,
    shell: 'Run the following powershell script as “ADMINISTRATOR” in a Server with the UTMStack agent Installed.',
    path: windowsPath,
    restart: windowsRestart
  },
  {
    id: 3,
    name: 'LINUX',
    command: linuxCommand,
    shell: 'Run the following bash script as “ADMINISTRATOR” in a Server with the UTMStack agent Installed.',
    path: linuxPath,
    restart: linuxRestart
  }
];

export const createFileBeatsPlatforms = (windowsCommand: string,
                                         linuxCommand: string,
                                         windowsPath?: string,
                                         windowsRestart?: string,
                                         linuxPath?: string,
                                         linuxRestart?: string) => [
  {
    id: 1,
    name: 'WINDOWS',
    command: windowsCommand,
    shell: 'Run the following powershell script as “ADMINISTRATOR” in a Server with the UTMStack agent Installed.',
    path: windowsPath,
    restart: windowsRestart
  },
  {
    id: 3,
    name: 'LINUX',
    command: linuxCommand,
    shell: 'Run the following bash script as “ADMINISTRATOR” in a Server with the UTMStack agent Installed.',
    path: linuxPath,
    restart: linuxRestart
  }
];


export const PLATFORMS = createPlatforms(
  'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service.exe" -ArgumentList \'ACTION\',' +
  ' \'AGENT_NAME\', \'PORT\' -NoNewWindow -Wait\n',
  'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service_arm64.exe" -ArgumentList \'ACTION\',' +
  ' \'AGENT_NAME\', \'PORT\' -NoNewWindow -Wait\n',
  'sudo bash -c "/opt/utmstack-linux-agent/utmstack_agent_service ACTION AGENT_NAME PORT"'
);


export const FILEBEAT_PLATFORMS = createFileBeatsPlatforms(
  'cd "C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\"; Start-Process "filebeat.exe" -ArgumentList "modules", "enable", \"AGENT_NAME\"',
  'cd /opt/utmstack-linux-agent/beats/filebeat/ && ./filebeat modules enable AGENT_NAME',
  'C:\\Program Files\\UTMStack\\UTMStack Agent\\beats\\filebeat\\modules.d\\',
  'Stop-Service -Name UTMStackModulesLogsCollector; Start-Sleep -Seconds 5; Start-Service -Name UTMStackModulesLogsCollector',
  '/opt/utmstack-linux-agent/beats/filebeat/modules.d/',
  'sudo systemctl restart UTMStackModulesLogsCollector'
);
