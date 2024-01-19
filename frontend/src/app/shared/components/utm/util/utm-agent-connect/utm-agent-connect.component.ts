import {Component, Input, OnInit} from '@angular/core';
import {NetScanType} from '../../../../../assets-discover/shared/types/net-scan.type';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';
import {IncidentCommandType} from '../../../../types/incident/incident-command.type';

@Component({
  selector: 'app-utm-agent-connect',
  templateUrl: './utm-agent-connect.component.html',
  styleUrls: ['./utm-agent-connect.component.css']
})
export class UtmAgentConnectComponent implements OnInit {
  @Input() hostname: string;
  @Input() asset: NetScanType;
  @Input() websocketCommand: IncidentCommandType;
  agent: AgentType;
  connectToAgent = false;
  hasNoReason = false;

  constructor(private agentManagerService: UtmAgentManagerService) {
  }

  ngOnInit() {
    this.hasNoReason = this.websocketCommand.reason === '' || !this.websocketCommand.reason;
    if (this.hostname) {
      this.agentManagerService.getAgent(this.hostname).subscribe(response => {
        this.agent = response.body;
      },
      error => this.assetTypeToAgentType());
    }
  }

  onAgentSelect($event: AgentType) {
    this.websocketCommand.reason = '';
    this.hasNoReason = true;
    this.agent = $event;
  }

  assetTypeToAgentType() {
    this.agent = {
      ip: this.asset.assetIp,
      hostname: this.asset.assetName,
      os: this.asset.assetOs,
      status: AgentStatusEnum.OFFLINE,
      platform: this.asset.assetOsPlatform,
      version: this.asset.assetOsMinorVersion,
      agentKey: '',
      id: this.asset.id,
      lastSeen: this.asset.modifiedAt,
      mac: '',
      osMajorVersion: this.asset.assetOsMajorVersion,
      osMinorVersion: this.asset.assetOs,
      aliases: '',
      addresses: ''
    };
  }
}
