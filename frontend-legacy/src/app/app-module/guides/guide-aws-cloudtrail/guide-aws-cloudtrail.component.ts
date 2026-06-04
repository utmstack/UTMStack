import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-aws-cloudtrail',
  templateUrl: './guide-aws-cloudtrail.component.html',
  styleUrls: ['./guide-aws-cloudtrail.component.scss']
})
export class GuideAwsCloudtrailComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
