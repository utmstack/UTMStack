import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-netflow',
  templateUrl: './guide-netflow.component.html',
  styleUrls: ['./guide-netflow.component.css']
})
export class GuideNetflowComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  netflowPaths: NetflowOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  constructor() {
  }

  ngOnInit() {
  }

}
export class NetflowOSPaths {
  os: string;
  path: string;
}
