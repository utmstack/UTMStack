import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {VMWARE_STEPS} from "../guide-vmware-syslog/vmware.steps";
import {Step} from "../shared/step";
import {SYSLOGSTEPS} from "./syslog.steps";

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

  syslogPorts: SyslogModulePorts[] = [
    {module: UtmModulesEnum.FORTIGATE, port: '7005 TCP'},
    {module: UtmModulesEnum.FORTIGATE, port: '7005 UDP'},

    {module: UtmModulesEnum.UFW, port: '7011 TCP'},
    {module: UtmModulesEnum.UFW, port: '7011 UDP'},

    {module: UtmModulesEnum.MIKROTIK, port: '7007 TCP'},
    {module: UtmModulesEnum.MIKROTIK, port: '7007 UDP'},

    {module: UtmModulesEnum.PALO_ALTO, port: '7006 TCP'},
    {module: UtmModulesEnum.PALO_ALTO, port: '7006 UDP'},

    {module: UtmModulesEnum.SONIC_WALL, port: '7009 TCP'},
    {module: UtmModulesEnum.SONIC_WALL, port: '7009 UDP'},

    {module: UtmModulesEnum.DECEPTIVE_BYTES, port: '7010 TCP'},
    {module: UtmModulesEnum.DECEPTIVE_BYTES, port: '7010 UDP'},

    {module: UtmModulesEnum.SOPHOS_XG, port: '7008 TCP'},
    {module: UtmModulesEnum.SOPHOS_XG, port: '7008 UDP'},

    {module: UtmModulesEnum.SYSLOG, port: '7014 TCP'},
    {module: UtmModulesEnum.SYSLOG, port: '7014 UDP'},

    {module: UtmModulesEnum.FIRE_POWER, port: '1470 TCP'},
    {module: UtmModulesEnum.FIRE_POWER, port: '514 UDP'},

    {module: UtmModulesEnum.CISCO, port: '1470 TCP'},
    {module: UtmModulesEnum.CISCO, port: '514 UDP'},

    {module: UtmModulesEnum.MERAKI, port: '1470 TCP'},
    {module: UtmModulesEnum.MERAKI, port: '514 UDP'},

    {module: UtmModulesEnum.CISCO_SWITCH, port: '1470 TCP'},
    {module: UtmModulesEnum.CISCO_SWITCH, port: '514 UDP'},

    {module: UtmModulesEnum.PFSENSE, port: '7017 TCP'},
    {module: UtmModulesEnum.PFSENSE, port: '7017 UDP'},

    {module: UtmModulesEnum.FORTIWEB, port: '7018 TCP'},
    {module: UtmModulesEnum.FORTIWEB, port: '7018 UDP'},

    {module: UtmModulesEnum.NETFLOW, port: '2055 UDP'},

    {module: UtmModulesEnum.AIX, port: '7016 TCP'},
    {module: UtmModulesEnum.AIX, port: '7016 UDP'},
  ];

  steps: Step[] = SYSLOGSTEPS;

  constructor() {
  }

  ngOnInit() {}
  getImage(): SyslogModuleImages {
    return this.moduleImages.filter(value => value.module === this.moduleEnum)[0];
  }
  getPorts(): SyslogModulePorts[] {
    return this.syslogPorts.filter(value => value.module === this.moduleEnum);
  }

  getProtocols(): { name: string; id: number }[] {
    return this.getPorts().map((port, index) => {
      return {
        id: index,
        name: port.port.includes('TLS') ? 'TLS' : port.port.split(' ')[1]
      };
    });
  }

  getSteps() {
    return this.moduleEnum === this.module.SYSLOG
      ? this.steps.slice(0, 2) : this.steps;
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
