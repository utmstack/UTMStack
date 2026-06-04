import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-github',
  templateUrl: './guide-github.component.html',
  styleUrls: ['./guide-github.component.css']
})
export class GuideGitHubComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;
  instance = window.location.host;
  constructor() { }

  ngOnInit() {
  }

}
