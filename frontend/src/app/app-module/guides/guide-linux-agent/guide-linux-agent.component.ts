import {Component, Input, OnInit} from '@angular/core';
import {FederationConnectionService} from '../../../app-management/connection-key/shared/services/federation-connection.service';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-linux-agent',
  templateUrl: './guide-linux-agent.component.html',
  styleUrls: ['./guide-linux-agent.component.css']
})
export class GuideLinuxAgentComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  @Input() version: string;
  token: string;
  platforms = [];

  constructor(private federationConnectionService: FederationConnectionService) { }

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
      this.loadPlatforms();
    });
  }

  getCommandUbuntu(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "apt update -y && apt install wget -y && mkdir -p /opt/utmstack_agent_service_linux_amd64 && \
    wget --no-check-certificate -P /opt/utmstack_agent_service_linux_amd64 \
    https://${ip}:9001/private/dependencies/agent/${installerName} && \
    chmod -R 755 /opt/utmstack_agent_service_linux_amd64/${installerName} && \
    /opt/utmstack_agent_service_linux_amd64/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getCommandCentos7RedHat(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "yum install wget -y && mkdir -p /opt/utmstack_agent_service_linux_amd64 && \
    wget --no-check-certificate -P /opt/utmstack_agent_service_linux_amd64 \
    https://${ip}:9001/private/dependencies/agent/${installerName} && \
    chmod -R 755 /opt/utmstack_agent_service_linux_amd64/${installerName} && \
    /opt/utmstack_agent_service_linux_amd64/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getCommandCentos8Almalinux(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "dnf install wget -y && mkdir -p /opt/utmstack_agent_service_linux_amd64 && \
    wget --no-check-certificate -P /opt/utmstack_agent_service_linux_amd64 \
    https://${ip}:9001/private/dependencies/agent/${installerName} && \
    chmod -R 755 /opt/utmstack_agent_service_linux_amd64/${installerName} && \
    /opt/utmstack_agent_service_linux_amd64/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getUninstallCommand(installerName: string): string {
    return `sudo bash -c "/opt/utmstack_agent_service_linux_amd64/${installerName} uninstall || true; \
    systemctl stop UTMStackAgent 2>/dev/null || true; systemctl disable UTMStackAgent 2>/dev/null || true; \
    rm -f /etc/systemd/system/UTMStackAgent.service 2>/dev/null || true; \
    systemctl stop UTMStackModulesLogsCollector 2>/dev/null || true; \
    systemctl disable UTMStackModulesLogsCollector 2>/dev/null || true; \
    rm -f /etc/systemd/system/UTMStackModulesLogsCollector.service 2>/dev/null || true; \
    systemctl daemon-reload 2>/dev/null || true; \
    echo 'Removing UTMStack Agent dependencies...' && sleep 10 && rm -rf /opt/utmstack_agent_service_linux_amd64 2>/dev/null || true; \
    echo 'UTMStack Agent dependencies removed successfully.'"`;
  }

  private loadPlatforms() {
    const amd64 = 'utmstack_agent_service_linux_amd64';
    const arm64 = 'utmstack_agent_service_linux_arm64';

    this.platforms = [
      {
        id: 1, name: 'Ubuntu / Debian (AMD64)',
        install: this.getCommandUbuntu(amd64),
        uninstall: this.getUninstallCommand(amd64),
        shell: ''
      },
      {
        id: 2, name: 'Ubuntu / Debian (ARM64)',
        install: this.getCommandUbuntu(arm64),
        uninstall: this.getUninstallCommand(arm64),
        shell: ''
      },
      {
        id: 3, name: 'Fedora / RedHat (AMD64)',
        install: this.getCommandCentos7RedHat(amd64),
        uninstall: this.getUninstallCommand(amd64),
        shell: ''
      },
      {
        id: 4, name: 'Fedora / RedHat (ARM64)',
        install: this.getCommandCentos7RedHat(arm64),
        uninstall: this.getUninstallCommand(arm64),
        shell: ''
      }
    ];
  }
}
