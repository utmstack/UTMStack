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
    });
  }

  getCommand(): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;
    return `New-Item -ItemType Directory -Force -Path "C:\\Program Files\\UTMStack\\UTMStack Agent"; ` +
      `Invoke-WebRequest -Uri "https://cdn.utmstack.com/agent_updates/release/installer/v${this.version}/utmstack_agent_installer.exe" ` +
      `-OutFile "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_installer.exe"; ` +
      `Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_installer.exe" ` +
      `-ArgumentList 'install', '` + ip  + `', '<secret>` + this.token + `</secret>', 'yes' -NoNewWindow -Wait`;
  }
  getUninstallCommand(): string {
    return `Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_installer.exe" -ArgumentList ` +
      `'uninstall' -NoNewWindow -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'stop','UTMStackAgent' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'delete','UTMStackAgent' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'stop','UTMStackRedline' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'delete','UTMStackRedline' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'stop','UTMStackUpdater' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'delete','UTMStackUpdater' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'stop','UTMStackWindowsLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'delete','UTMStackWindowsLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'stop','UTMStackModulesLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; Start-Process -FilePath "sc.exe" ` +
      `-ArgumentList 'delete','UTMStackModulesLogsCollector' -Wait -ErrorAction SilentlyContinue | Out-Null; ` +
      `Write-Host "Removing UTMStack Agent dependencies..."; Start-Sleep -Seconds 10; Remove-Item 'C:\\Program Files\\UTMStack' ` +
      `-Recurse -Force -ErrorAction Stop; Write-Host "UTMStack Agent removed successfully."`;
  }

}
