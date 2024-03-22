import {Component, Input, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {VMWARESTEPS} from './vmware.steps';

@Component({
  selector: 'app-guide-vmware-syslog',
  templateUrl: './guide-vmware-syslog.component.html',
  styleUrls: ['./guide-vmware-syslog.component.scss']
})
export class GuideVmwareSyslogComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  @ViewChild('stepContent6') stepContent6: TemplateRef<any>;
  @ViewChild('stepContent10') stepCommandContent: TemplateRef<any>;
  module = UtmModulesEnum;
  vmwarePaths: VmwareOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];

  steps: Step[] = VMWARESTEPS;
  constructor() {
  }

  ngOnInit() {}

  getContentStep(id: string){
    switch (id) {
      case 'stepContent6':
        return this.stepContent6;
      case 'stepContent10':
        return this.stepCommandContent;
    }
  }
}
export class VmwareOSPaths {
  os: string;
  path: string;
}
