import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-webroot',
  templateUrl: './guide-webroot.component.html',
  styleUrls: ['./guide-webroot.component.scss']
})
export class GuideWebrootComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  configValidity: boolean;

  constructor() {
  }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

}
