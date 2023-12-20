import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';


@Component({
  selector: 'app-guide-aws-lambda',
  templateUrl: './guide-aws-lambda.component.html',
  styleUrls: ['./guide-aws-lambda.component.scss']
})
export class GuideAwsLambdaComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
