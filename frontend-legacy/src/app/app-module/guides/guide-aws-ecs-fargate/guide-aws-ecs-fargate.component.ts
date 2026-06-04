import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-aws-ecs-fargate',
  templateUrl: './guide-aws-ecs-fargate.component.html',
  styleUrls: ['./guide-aws-ecs-fargate.component.scss']
})
export class GuideAwsEcsFargateComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
