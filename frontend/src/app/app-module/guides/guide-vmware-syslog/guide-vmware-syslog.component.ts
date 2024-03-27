import {Component, Input, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {VMWARE_STEPS} from './vmware.steps';

@Component({
  selector: 'app-guide-vmware-syslog',
  templateUrl: './guide-vmware-syslog.component.html',
  styleUrls: ['./guide-vmware-syslog.component.scss']
})
export class GuideVmwareSyslogComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  steps: Step[] = VMWARE_STEPS;

  constructor() {
  }

  ngOnInit() {}

}
export class VmwareOSPaths {
  os: string;
  path: string;
}
