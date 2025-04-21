import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {MAC_STEPS} from './mac.steps';
import {
  FederationConnectionService
} from "../../../app-management/connection-key/shared/services/federation-connection.service";

@Component({
  selector: 'app-guide-macos-agent',
  templateUrl: './guide-macos-agent.component.html',
  styleUrls: ['./guide-macos-agent.component.css']
})
export class GuideMacosAgentComponent implements OnInit {
  @Input() integrationId: number;
  @Input() filebeatModule: UtmModulesEnum;
  @Input() filebeatModuleName: string;
  module = UtmModulesEnum;
  @Input() serverId: number;

  steps: Step[] = MAC_STEPS;
  architectures = [];
  token: string;
  constructor(private federationConnectionService: FederationConnectionService,) { }

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

  getCommandARM(installerName: string): string {
    const ip = window.location.host.includes(':') ? window.location.host.split(':')[0] : window.location.host;

    return `sudo bash -c "./${installerName} ${ip} <secret>${this.token}</secret> yes"`;
  }


  getUninstallCommand(installerName: string): string {
    return `sudo bash -c "./utmstack_agent_service uninstall"`;
  }

  private loadArchitectures() {
    this.architectures = [
      {
        id: 1, name: 'ARM64',
        install: this.getCommandARM('utmstack_agent_service install'),
        uninstall: this.getUninstallCommand('utmstack_agent_service install'),
        shell: ''
      },
    ];
  }
}
