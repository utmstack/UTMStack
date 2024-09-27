import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {SyslogModulePorts} from '../guide-syslog/guide-syslog.component';
import {Step} from '../shared/step';
import {KASP_STEPS} from './kasp-steps';
import {PLATFORMS} from "../shared/constant";

@Component({
  selector: 'app-guide-kaspersky',
  templateUrl: './guide-kaspersky.component.html',
  styleUrls: ['./guide-kaspersky.component.css']
})
export class GuideKasperskyComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  @Input() dataType: string;
  steps: Step[] = KASP_STEPS;
  platforms = PLATFORMS;

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
}
