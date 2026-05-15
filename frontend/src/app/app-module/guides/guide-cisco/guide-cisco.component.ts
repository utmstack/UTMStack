import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-cisco',
  templateUrl: './guide-cisco.component.html',
  styleUrls: ['./guide-cisco.component.scss']
})
export class GuideCiscoComponent implements OnInit {
  @Input() integrationId: number;
  @Input() filebeatModule: UtmModulesEnum;
  @Input() filebeatModuleName: string;
  module = UtmModulesEnum;
  @Input() serverId: number;

  ciscoPaths: CiscoOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  ciscoPorts: CiscoPorts[] = [
    {module: UtmModulesEnum.CISCO, port: '1470  TCP'},
    {module: UtmModulesEnum.CISCO, port: '514 UDP'}
  ];

  constructor() {
  }

  ngOnInit() {
  }
  getPorts(): CiscoPorts[] {
    return this.ciscoPorts;
  }
}

export class CiscoPorts {
  module: UtmModulesEnum;
  port: string;
}

export class CiscoOSPaths {
  os: string;
  path: string;
}
