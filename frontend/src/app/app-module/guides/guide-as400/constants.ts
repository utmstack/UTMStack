
export const PLATFORM = [
    {
        id: 1,
        name: 'WINDOWS',
        install: `New-Item -ItemType Directory -Force -Path "C:\\Program Files\\UTMStack\\UTMStack Collectors\\AS400"; ` +
                 `cd "C:\\Program Files\\UTMStack\\UTMStack Collectors\\AS400"; ` +
                 `Invoke-WebRequest -Uri "https://storage.googleapis.com/utmstack-updates/collectors/windows-as400-collector.zip" ` +
                 `-OutFile ".\\windows-as400-collector.zip"; Expand-Archive -Path ".\\windows-as400-collector.zip" ` +
                 `-DestinationPath "."; Remove-Item ".\\windows-as400-collector.zip"; Start-Process ".\\utmstack_collectors_installer.exe" ` +
                 `-ArgumentList 'install', 'as400', 'V_IP', '<secret>V_TOKEN</secret>' -NoNewWindow -Wait`,

        uninstall: `cd "C:\\Program Files\\UTMStack\\UTMStack Collectors\\AS400"; ` +
                   `Start-Process ".\\utmstack_collectors_installer.exe" -ArgumentList ` +
                   ` 'uninstall', 'as400' -NoNewWindow -Wait -ErrorAction SilentlyContinue ` +
                   `| Out-Null; Start-Process -FilePath "sc.exe" -ArgumentList 'stop', ` +
                   `'UTMStackAS400Collector' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
                   `Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackAS400Collector' ` +
                   `-Wait -ErrorAction SilentlyContinue | Out-Null; Write-Host "Removing UTMStack AS400 Collector dependencies..."; ` +
                   `Get-Process | Where-Object { $_.Path -like "*java.exe*" } | Stop-Process -Force; ` +
                   `cd "C:\\Program Files\\UTMStack\\UTMStack Collectors"; Start-Sleep -Seconds 5; ` +
                   `cd "C:\\Program Files"; Remove-Item 'C:\\Program Files\\UTMStack\\UTMStack Collectors\\AS400' ` +
                   `-Recurse -Force -ErrorAction Stop; Write-Host "UTMStack AS400 Collector removed successfully."`,

        shell: 'Open Windows Powershell terminal as “ADMINISTRATOR”'
    },
    {
        id: 2,
        name: 'LINUX UBUNTU',
        install: `sudo bash -c "apt update -y && apt install wget unzip -y && mkdir -p ` +
                 `/opt/utmstack-linux-collectors/as400 && cd /opt/utmstack-linux-collectors/as400 && ` +
                 `wget https://storage.googleapis.com/utmstack-updates/collectors/linux-as400-collector.zip ` +
                 `&& unzip linux-as400-collector.zip && rm linux-as400-collector.zip && chmod -R 777 ` +
                 `utmstack_collectors_installer && ./utmstack_collectors_installer install as400 ` +
                 `V_IP <secret>V_TOKEN</secret>"`,

        uninstall: `sudo bash -c " cd /opt/utmstack-linux-collectors/as400 && ./utmstack_collectors_installer ` +
                   `uninstall as400 && echo 'Removing UTMStack AS400 Collector dependencies...' && sleep 5 && rm ` +
                   `-rf /opt/utmstack-linux-collectors/as400 && echo 'UTMStack AS400 Collector removed successfully.'"`,

        shell: 'Linux bash terminal'
    }
];

export const ACTIONS = [
    {id: 1, name: 'INSTALL'},
    {id: 2, name: 'UNINSTALL'}
];
