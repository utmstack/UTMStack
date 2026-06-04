import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-aws-rds-postgres',
  templateUrl: './guide-aws-rds-postgres.component.html',
  styleUrls: ['./guide-aws-rds-postgres.component.scss']
})
export class GuideAwsRdsPostgresComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
