import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-vmware-syslog',
  templateUrl: './guide-vmware-syslog.component.html',
  styleUrls: ['./guide-vmware-syslog.component.scss']
})
export class GuideVmwareSyslogComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  vmwarePaths: VmwareOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  constructor() {
  }

  ngOnInit() {
  }

}
export class VmwareOSPaths {
  os: string;
  path: string;
}
