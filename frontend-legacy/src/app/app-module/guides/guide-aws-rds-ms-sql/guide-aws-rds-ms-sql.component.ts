import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-aws-rds-ms-sql',
  templateUrl: './guide-aws-rds-ms-sql.component.html',
  styleUrls: ['./guide-aws-rds-ms-sql.component.scss']
})
export class GuideAwsRdsMsSqlComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module= UtmModulesEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
