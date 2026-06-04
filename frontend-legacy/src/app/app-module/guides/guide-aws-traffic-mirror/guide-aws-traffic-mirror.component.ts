import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-aws-traffic-mirror',
  templateUrl: './guide-aws-traffic-mirror.component.html',
  styleUrls: ['./guide-aws-traffic-mirror.component.scss']
})
export class GuideAwsTrafficMirrorComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
