import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-vulnerabilities',
  templateUrl: './guide-vulnerabilities.component.html',
  styleUrls: ['./guide-vulnerabilities.component.css']
})
export class GuideVulnerabilitiesComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  configValidity: boolean;


  constructor() {
  }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }


}
