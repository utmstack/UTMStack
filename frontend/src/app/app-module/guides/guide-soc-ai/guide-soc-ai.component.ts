import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-soc-ai',
  templateUrl: './guide-soc-ai.component.html',
  styleUrls: ['./guide-soc-ai.component.css']
})
export class GuideSocAiComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  configValidity = false;

  constructor() {
  }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

}
