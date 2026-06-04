import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {SyslogModulePorts} from '../guide-syslog/guide-syslog.component';
import {PLATFORMS} from '../shared/constant';
import {Step} from '../shared/step';
import {ESET_STEPS} from './eset-steps';

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
  dataType = 'antivirus-esmc-eset';
  platforms = PLATFORMS;
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
}
