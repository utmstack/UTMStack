import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {
    FederationConnectionService
} from '../../../app-management/connection-key/shared/services/federation-connection.service';
import {GroupTypeEnum} from '../../shared/enum/group-type.enum';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {UtmstackSteps} from './utmstack.steps';

@Component({
    selector: 'app-guide-utmstack',
    templateUrl: './guide-utmstack.component.html',
    styleUrls: ['./guide-utmstack.component.css']
})
export class GuideUtmstackComponent implements OnInit {
    @Input() integrationId: number;
    @Input() serverId: number;
    module = UtmModulesEnum;
    serverAS400FormArray: FormGroup;
    configValidity: boolean;
    groupType = GroupTypeEnum.COLLECTOR;
    steps: Step[] = UtmstackSteps;
    token: string;
    ip: string;
    vars: any;
    disablePreAction = false;
    performPreAction = true;
    architectures = [];


    constructor(private formBuilder: FormBuilder,
                private federationConnectionService: FederationConnectionService) {
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

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

  onDisable() {
    this.disablePreAction = true;
  }

  private loadArchitectures() {
    this.architectures = [
      {
        id: 1, name: 'Ubuntu 16/18/20+',
        install: this.getCommandUbuntu('utmstack_collector'),
        uninstall: this.getUninstallCommandUbuntu('utmstack_collector'),
        shell: ''
      },
      {
        id: 2, name: 'CentOS 8+/Red Hat Enterprise Linux',
        install: this.getCommandCentos7RedHat('utmstack_collector'),
        uninstall: this.getUninstallCommandRedHat('utmstack_collector'),
        shell: ''
      },
    ];
  }

  getCommandUbuntu(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c 'apt update -y && \
            apt install wget -y && \
            mkdir -p /opt/utmstack-collector && \
            wget --no-check-certificate -P /opt/utmstack-collector \
            https://${ip}:9001/private/dependencies/collector/${installerName} && \
            chmod -R 777 /opt/utmstack-collector/${installerName} && \
            /opt/utmstack-collector/${installerName} install ${ip} <secret>${this.token}</secret>' yes'`;
  }


  getCommandCentos7RedHat(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "dnf update -y && \
            dnf install wget -y && \
            mkdir -p /opt/utmstack-collector && \
            wget --no-check-certificate -P /opt/utmstack-collector \
            https://${ip}:9001/private/dependencies/collector/${installerName} && \
            chmod -R 777 /opt/utmstack-collector/${installerName} && \
            /opt/utmstack-collector/${installerName} install ${ip} <secret>${this.token}</secret> yes"`;
  }

  getUninstallCommandUbuntu(installerName: string): string {
    return `sudo bash -c "/opt/utmstack-collector/${installerName} uninstall && \
            (systemctl stop UTMStackCollector 2>/dev/null || service UTMStackCollector stop 2>/dev/null || true) && \
            (systemctl disable UTMStackCollector 2>/dev/null || chkconfig UTMStackCollector off 2>/dev/null || true) && \
            rm -rf /opt/utmstack-collector && \
            rm -f /etc/systemd/system/UTMStackCollector.service && \
            rm -f /etc/init.d/UTMStackCollector && \
            (systemctl daemon-reload 2>/dev/null || true) && \
            echo 'UTMStack Collector uninstalled successfully'"`;
  }


  getUninstallCommandRedHat(installerName: string): string {
    return `sudo bash -c "/opt/utmstack-collector/${installerName} uninstall && \
            (systemctl stop UTMStackCollector 2>/dev/null || true) && \
            (systemctl disable UTMStackCollector 2>/dev/null || true) && \
            rm -rf /opt/utmstack-collector && \
            rm -f /etc/systemd/system/UTMStackCollector.service && \
            systemctl daemon-reload && \
            echo 'UTMStack Collector uninstalled successfully'"`;
  }
}
