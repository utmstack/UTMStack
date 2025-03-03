import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {AS400STEPS} from "../guide-as400/as400.steps";
import {ACTIONS, PLATFORM} from "../guide-as400/constants";
import {Step} from "../shared/step";
import {SOPHOS_STEPS} from "./sophos.steps";

@Component({
  selector: 'app-guide-sophos',
  templateUrl: './guide-sophos.component.html',
  styleUrls: ['./guide-sophos.component.css']
})
export class GuideSophosComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  configValidity: boolean;
  steps = SOPHOS_STEPS;
  platforms = PLATFORM;
  actions = ACTIONS;

  constructor() {
  }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }
}
