import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

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

  esetPaths: EsetOSPaths[] = [
    {os: 'linux', path: '/opt/utmstack-linux-agent/log-collector-config.json'},
    {os: 'windows', path: 'C:\\Program Files\\UTMStack\\UTMStack Agent\\log-collector-config.json'}
  ];
  constructor() {
  }

  ngOnInit() {
  }

}
export class EsetOSPaths {
  os: string;
  path: string;
}
