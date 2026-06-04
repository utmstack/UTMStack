import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-office365',
  templateUrl: './guide-office365.component.html',
  styleUrls: ['./guide-office365.component.scss']
})
export class GuideOffice365Component implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  configValidity = false;

  constructor() {
  }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }
}
