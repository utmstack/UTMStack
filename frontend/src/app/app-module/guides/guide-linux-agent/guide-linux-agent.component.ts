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
  architectures = [];

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
      this.loadArchitectures();
    });
  }

  getCommandUbuntu(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "apt update -y && apt install wget -y && mkdir -p /opt/utmstack-linux-agent && \
    wget --no-check-certificate --header='connection-key: <secret>${this.token}</secret>' -P /opt/utmstack-linux-agent \
    https://${ip}:9001/private/dependencies/agent/${installerName} && \
    chmod -R 777 /opt/utmstack-linux-agent/${installerName} && \
    /opt/utmstack-linux-agent/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getCommandCentos7RedHat(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "yum install wget -y && mkdir -p /opt/utmstack-linux-agent && \
    wget --no-check-certificate --header='connection-key: <secret>${this.token}</secret>' -P /opt/utmstack-linux-agent \
    https://${ip}:9001/private/dependencies/agent/${installerName} && \
    chmod -R 777 /opt/utmstack-linux-agent/${installerName} && \
    /opt/utmstack-linux-agent/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getCommandCentos8Almalinux(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "dnf install wget -y && mkdir -p /opt/utmstack-linux-agent && \
    wget --no-check-certificate --header='connection-key: <secret>${this.token}</secret>' -P /opt/utmstack-linux-agent \
    https://${ip}:9001/private/dependencies/agent/${installerName} && \
    chmod -R 777 /opt/utmstack-linux-agent/${installerName} && \
    /opt/utmstack-linux-agent/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getUninstallCommand(installerName: string): string {
    return `sudo bash -c "/opt/utmstack-linux-agent/${installerName} uninstall || true; \
    systemctl stop UTMStackAgent 2>/dev/null || true; systemctl disable UTMStackAgent 2>/dev/null || true; \
    rm -f /etc/systemd/system/UTMStackAgent.service 2>/dev/null || true; \
    systemctl stop UTMStackModulesLogsCollector 2>/dev/null || true; \
    systemctl disable UTMStackModulesLogsCollector 2>/dev/null || true; \
    rm -f /etc/systemd/system/UTMStackModulesLogsCollector.service 2>/dev/null || true; \
    systemctl daemon-reload 2>/dev/null || true; \
    echo 'Removing UTMStack Agent dependencies...' && sleep 10 && rm -rf /opt/utmstack-linux-agent 2>/dev/null || true; \
    echo 'UTMStack Agent dependencies removed successfully.'"`;
  }

  private loadArchitectures() {
    this.architectures = [
      {
        id: 1, name: 'Ubuntu 16/18/20+',
        install: this.getCommandUbuntu('utmstack_agent_service'),
        uninstall: this.getUninstallCommand('utmstack_agent_service'),
        shell: ''
      },
      {
        id: 2, name: 'Centos 7/Red Hat Enterprise Linux',
        install: this.getCommandCentos7RedHat('utmstack_agent_service'),
        uninstall: this.getUninstallCommand('utmstack_agent_service'),
        shell: ''
      },
      {
        id: 3, name: 'Centos 8/AlmaLinux',
        install: this.getCommandCentos8Almalinux('utmstack_agent_service'),
        uninstall: this.getUninstallCommand('utmstack_agent_service'),
        shell: ''
      }
    ];
  }
}
