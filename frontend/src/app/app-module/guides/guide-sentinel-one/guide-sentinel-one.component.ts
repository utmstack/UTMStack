import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-sentinel-one',
  templateUrl: './guide-sentinel-one.component.html',
  styleUrls: ['./guide-sentinel-one.component.css']
})
export class GuideSentinelOneComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;

  sentinelonePaths: SentinelOneOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  constructor() {
  }

  ngOnInit() {
  }

}
export class SentinelOneOSPaths {
  os: string;
  path: string;
}
