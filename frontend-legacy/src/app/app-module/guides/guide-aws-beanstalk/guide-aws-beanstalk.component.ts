import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-aws-beanstalk',
  templateUrl: './guide-aws-beanstalk.component.html',
  styleUrls: ['./guide-aws-beanstalk.component.scss']
})
export class GuideAwsBeanstalkComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
