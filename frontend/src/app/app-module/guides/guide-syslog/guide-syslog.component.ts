import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-syslog',
  templateUrl: './guide-syslog.component.html',
  styleUrls: ['./guide-syslog.component.scss']
})
export class GuideSyslogComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  @Input() guideName: string;
  @Input() moduleEnum: UtmModulesEnum;
  module = UtmModulesEnum;
  moduleImages: SyslogModuleImages[] = [
    {module: UtmModulesEnum.FORTIGATE, img: 'fortigate.png'},
    {module: UtmModulesEnum.UFW, img: 'ufw.png'},
    {module: UtmModulesEnum.MIKROTIK, img: 'mikrotik.png'},
    {module: UtmModulesEnum.PALO_ALTO, img: 'paloalto.png'},
    {module: UtmModulesEnum.SONIC_WALL, img: 'sonicwall.png'},
    {module: UtmModulesEnum.DECEPTIVE_BYTES, img: 'deceptivebytes.png'},
    {module: UtmModulesEnum.SOPHOS_XG, img: 'sophosxg.png'}
  ];
  syslogPaths: SyslogOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  syslogPorts: SyslogModulePorts[] = [
    {module: UtmModulesEnum.FORTIGATE, port: '7005 TCP'},
    {module: UtmModulesEnum.FORTIGATE, port: '7005 UDP'},
    {module: UtmModulesEnum.FORTIGATE, port: '7055 TCP (TLS)'},

    {module: UtmModulesEnum.UFW, port: '7011 TCP'},
    {module: UtmModulesEnum.UFW, port: '7011 UDP'},
    {module: UtmModulesEnum.UFW, port: '7061 TCP (TLS)'},

    {module: UtmModulesEnum.MIKROTIK, port: '7007 TCP'},
    {module: UtmModulesEnum.MIKROTIK, port: '7007 UDP'},
    {module: UtmModulesEnum.MIKROTIK, port: '7057 TCP (TLS)'},

    {module: UtmModulesEnum.PALO_ALTO, port: '7006 TCP'},
    {module: UtmModulesEnum.PALO_ALTO, port: '7006 UDP'},
    {module: UtmModulesEnum.PALO_ALTO, port: '7056 TCP (TLS)'},

    {module: UtmModulesEnum.SONIC_WALL, port: '7009 TCP'},
    {module: UtmModulesEnum.SONIC_WALL, port: '7009 UDP'},
    {module: UtmModulesEnum.SONIC_WALL, port: '7059 TCP (TLS)'},

    {module: UtmModulesEnum.DECEPTIVE_BYTES, port: '7010 TCP'},
    {module: UtmModulesEnum.DECEPTIVE_BYTES, port: '7010 UDP'},
    {module: UtmModulesEnum.DECEPTIVE_BYTES, port: '7060 TCP (TLS)'},

    {module: UtmModulesEnum.SOPHOS_XG, port: '7008 TCP'},
    {module: UtmModulesEnum.SOPHOS_XG, port: '7008 UDP'},
    {module: UtmModulesEnum.SOPHOS_XG, port: '7058 TCP (TLS)'}
  ];

  constructor() {
  }

  ngOnInit() {
  }
  getImage(): SyslogModuleImages {
    return this.moduleImages.filter(value => value.module === this.moduleEnum)[0];
  }
  getPorts(): SyslogModulePorts[] {
    return this.syslogPorts.filter(value => value.module === this.moduleEnum);
  }
}

export class SyslogModuleImages {
  module: UtmModulesEnum;
  img: string;
}

export class SyslogModulePorts {
  module: UtmModulesEnum;
  port: string;
}

export class SyslogOSPaths {
  os: string;
  path: string;
}
