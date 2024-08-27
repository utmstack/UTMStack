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
    });
  }

  getCommandUbuntu(): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "apt update -y && apt install curl jq -y && mkdir -p /opt/utmstack-linux-agent && ` +
    `curl -k -L -H 'connection-key: <secret>${this.token}</secret>' 'http://${ip}/dependencies/agent?version=0&os=linux&type=installer&partIndex=1&partSize=20' ` +
    `| jq -r '.fileContent' | base64 -d > /opt/utmstack-linux-agent/utmstack_agent_installer && ` +
    `chmod -R 777 /opt/utmstack-linux-agent/utmstack_agent_installer && ` +
    `/opt/utmstack-linux-agent/utmstack_agent_installer install ${ip} <secret>${this.token}</secret> yes"`;
  }
  getCommandCentos7RedHat(): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "yum install curl jq -y && mkdir -p /opt/utmstack-linux-agent && ` +
      `curl -k -L -H 'connection-key: <secret>${this.token}</secret>' ` +
      `'http://${ip}/dependencies/agent?version=0&os=linux&type=installer&partIndex=1&partSize=20' ` +
      `| jq -r '.fileContent' | base64 -d > /opt/utmstack-linux-agent/utmstack_agent_installer && ` +
      `chmod -R 777 /opt/utmstack-linux-agent/utmstack_agent_installer && ` +
      `/opt/utmstack-linux-agent/utmstack_agent_installer install ${ip} <secret>${this.token}</secret> yes"`;
  }
  getCommandCentos8Almalinux(): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "dnf install curl jq -y && mkdir -p /opt/utmstack-linux-agent && ` +
      `curl -k -L -H 'connection-key: <secret>${this.token}</secret>' ` +
      `'http://${ip}/dependencies/agent?version=0&os=linux&type=installer&partIndex=1&partSize=20' ` +
      `| jq -r '.fileContent' | base64 -d > /opt/utmstack-linux-agent/utmstack_agent_installer && ` +
      `chmod -R 777 /opt/utmstack-linux-agent/utmstack_agent_installer && ` +
      `/opt/utmstack-linux-agent/utmstack_agent_installer install ${ip} <secret>${this.token}</secret> yes"`;
  }
  getUninstallCommand(): string {
    return `sudo bash -c "/opt/utmstack-linux-agent/utmstack_agent_installer uninstall || true; ` +
      `systemctl stop UTMStackAgent 2>/dev/null || true; ` +
      `systemctl disable UTMStackAgent 2>/dev/null || true; ` +
      `rm /etc/systemd/system/UTMStackAgent.service 2>/dev/null || true; ` +
      `systemctl stop UTMStackModulesLogsCollector 2>/dev/null || true; ` +
      `systemctl disable UTMStackModulesLogsCollector 2>/dev/null || true; ` +
      `rm /etc/systemd/system/UTMStackModulesLogsCollector.service 2>/dev/null || true; ` +
      `systemctl daemon-reload 2>/dev/null || true; ` +
      `echo 'Removing UTMStack Agent dependencies...' && sleep 10 && rm -rf /opt/utmstack-linux-agent && ` +
      `echo 'UTMStack Agent dependencies removed successfully.'"`;
  }
}
