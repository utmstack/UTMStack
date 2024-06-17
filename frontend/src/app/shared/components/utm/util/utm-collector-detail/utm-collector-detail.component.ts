import {Component, Input, OnInit} from '@angular/core';
import {NetScanType} from '../../../../../assets-discover/shared/types/net-scan.type';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';
import {UtmModuleCollectorType} from "../../../../../app-module/shared/type/utm-module-collector.type";

@Component({
  selector: 'app-utm-collector-detail',
  templateUrl: './utm-collector-detail.component.html',
  styleUrls: ['./utm-collector-detail.component.scss']
})
export class UtmCollectorDetailComponent implements OnInit {
  @Input() agent: UtmModuleCollectorType;
  @Input() hosts: string[];
  agentStatusEnum = AgentStatusEnum;

  constructor() {
  }

  ngOnInit() {}
}
