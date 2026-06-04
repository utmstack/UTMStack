import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {CROWDSTRIKE_STEPS} from './crowdstrike.steps';

@Component({
  selector: 'app-guide-crowdstrike',
  templateUrl: './guide-crowdstrike.component.html',
  styleUrls: ['./guide-crowdstrike.component.css']
})
export class GuideCrowdstrikeComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  configValidity = false;
  steps: Step[] = CROWDSTRIKE_STEPS;

  constructor() {
  }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

}
