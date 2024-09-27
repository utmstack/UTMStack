import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {SyslogModulePorts} from '../guide-syslog/guide-syslog.component';
import {VMWARE_STEPS} from '../guide-vmware-syslog/vmware.steps';
import {Step} from '../shared/step';
import {SENTINELSTEPS} from './sentinel.steps';
import {PLATFORMS} from "../shared/constant";

@Component({
  selector: 'app-guide-sentinel-one',
  templateUrl: './guide-sentinel-one.component.html',
  styleUrls: ['./guide-sentinel-one.component.css']
})
export class GuideSentinelOneComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  steps: Step[] = SENTINELSTEPS;
  @Input() dataType!: string;
  platforms = PLATFORMS;

  constructor() {
  }

  ngOnInit() {
  }

  getPorts(): SyslogModulePorts[] {
    return [
      {module: UtmModulesEnum.SENTINEL_ONE, port: '7012 TCP'},
      {module: UtmModulesEnum.SENTINEL_ONE, port: '7012 UDP'}
    ];
  }

  getProtocols() {
    return this.getPorts().map((port, index) => {
      return {
        id: index,
        name: port.port.split(' ')[1]
      };
    });
  }

}
