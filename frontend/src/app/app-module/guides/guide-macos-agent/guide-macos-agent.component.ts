import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {Step} from '../shared/step';
import {MAC_STEPS} from './mac.steps';

@Component({
  selector: 'app-guide-macos-agent',
  templateUrl: './guide-macos-agent.component.html',
  styleUrls: ['./guide-macos-agent.component.css']
})
export class GuideMacosAgentComponent implements OnInit {
  @Input() integrationId: number;
  @Input() filebeatModule: UtmModulesEnum;
  @Input() filebeatModuleName: string;
  module = UtmModulesEnum;
  @Input() serverId: number;

  steps: Step[] = MAC_STEPS;
  constructor() { }

  ngOnInit() {
    console.log(this.steps);
  }
}
