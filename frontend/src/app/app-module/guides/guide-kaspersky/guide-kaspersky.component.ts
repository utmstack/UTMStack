import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-kaspersky',
  templateUrl: './guide-kaspersky.component.html',
  styleUrls: ['./guide-kaspersky.component.css']
})
export class GuideKasperskyComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;

  kavPaths: KavOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  constructor() {
  }

  ngOnInit() {
  }

}
export class KavOSPaths {
  os: string;
  path: string;
}
