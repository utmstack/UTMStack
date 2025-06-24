import {Component, Input, OnInit} from '@angular/core';
import {FederationConnectionService} from '../../../app-management/connection-key/shared/services/federation-connection.service';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-winlogbeat',
  templateUrl: './guide-winlogbeat.component.html',
  styleUrls: ['./guide-winlogbeat.component.scss']
})
export class GuideWinlogbeatComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  token: string;
  @Input() version: string;

  architectures = [];

  constructor(private federationConnectionService: FederationConnectionService) {
  }

  ngOnInit() {
    this.getToken();
  }


  getToken() {
    this.federationConnectionService.getToken().subscribe(response => {
      if (response.body !== null && response.body !== '') {
        this.token = response.body;
      } else {
        this.token = '';
      }
      this.loadArchitectures();
    });
  }

  loadArchitectures(){
    this.architectures = [
      {
        id: 1, name: 'AMD64',
        install: this.getCommand('utmstack_agent_service.exe'),
        uninstall: this.getUninstallCommand('utmstack_agent_service.exe'),
        shell: 'Windows Powershell terminal as “ADMINISTRATOR”'
      },
      {
        id: 2, name: 'ARM64',
        install: this.getCommand('utmstack_agent_service_arm64.exe'),
        uninstall: this.getUninstallCommand('utmstack_agent_service_arm64.exe'),
        shell: 'Windows Powershell terminal as “ADMINISTRATOR”'
      }
    ];
  }

  getCommand(arch: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `New-Item -ItemType Directory -Force -Path "C:\\Program Files\\UTMStack\\UTMStack Agent"; ` +
      `& curl.exe -k -H "connection-key: <secret>${this.token}</secret>" ` +
      `-o "C:\\Program Files\\UTMStack\\UTMStack Agent\\${arch}" ` +
      `"https://${ip}:9001/private/dependencies/agent/${arch}"; ` +
      `Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\${arch}" ` +
      `-ArgumentList 'install', '${ip}', '<secret>${this.token}</secret>', 'yes' -NoNewWindow -Wait`;
  }

  getUninstallCommand(arch: string): string {
    return `Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\${arch}" ` +
      `-ArgumentList 'uninstall' -NoNewWindow -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Start-Process -FilePath "sc.exe" -ArgumentList 'stop','UTMStackAgent' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackAgent' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Start-Process -FilePath "sc.exe" -ArgumentList 'stop','UTMStackWindowsLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackWindowsLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Start-Process -FilePath "sc.exe" -ArgumentList 'stop','UTMStackModulesLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Start-Process -FilePath "sc.exe" -ArgumentList 'delete','UTMStackModulesLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Write-Host "Removing UTMStack Agent dependencies..."; ` +
      `Start-Sleep -Seconds 10; ` +
      `Remove-Item 'C:\\Program Files\\UTMStack\\UTMStack Agent' -Recurse -Force -ErrorAction Stop; ` +
      `Write-Host "UTMStack Agent removed successfully."`;
  }


}
