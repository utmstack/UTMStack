import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {VMWARE_STEPS} from "../guide-vmware-syslog/vmware.steps";
import {Step} from "../shared/step";
import {ESET_STEPS} from "./eset-steps";
import {SyslogModulePorts} from "../guide-syslog/guide-syslog.component";

@Component({
  selector: 'app-guide-eset',
  templateUrl: './guide-eset.component.html',
  styleUrls: ['./guide-eset.component.css']
})
export class GuideEsetComponent implements OnInit {
  @Input() integrationId: number;
  @Input() filebeatModule: UtmModulesEnum;
  @Input() filebeatModuleName: string;
  module = UtmModulesEnum;
  @Input() serverId: number;
  steps: Step[] = ESET_STEPS;
  constructor() {
  }

  ngOnInit() {
  }

  getPorts(): SyslogModulePorts[] {
    return [
      {module: UtmModulesEnum.SENTINEL_ONE, port: '7003 TCP'},
      {module: UtmModulesEnum.SENTINEL_ONE, port: '7003 UDP'}
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
