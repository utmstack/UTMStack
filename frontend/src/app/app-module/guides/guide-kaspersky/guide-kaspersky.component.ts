import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {ESET_STEPS} from '../guide-eset/eset-steps';
import {SyslogModulePorts} from '../guide-syslog/guide-syslog.component';
import {Step} from '../shared/step';
import {KASP_STEPS} from './kasp-steps';

@Component({
  selector: 'app-guide-kaspersky',
  templateUrl: './guide-kaspersky.component.html',
  styleUrls: ['./guide-kaspersky.component.css']
})
export class GuideKasperskyComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  steps: Step[] = KASP_STEPS;

  constructor() {
  }

  ngOnInit() {
  }

  getPorts(): SyslogModulePorts[] {
    return [
      {module: UtmModulesEnum.SENTINEL_ONE, port: '7004 TCP'},
      {module: UtmModulesEnum.SENTINEL_ONE, port: '7004 UDP'}
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

  protected readonly ESET_STEPS = ESET_STEPS;
}
