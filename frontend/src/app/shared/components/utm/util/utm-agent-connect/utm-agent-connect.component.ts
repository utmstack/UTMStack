import {Component, Input, OnInit} from '@angular/core';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentType} from '../../../../types/agent/agent.type';
import {IncidentCommandType} from '../../../../types/incident/incident-command.type';
import {NetScanType} from "../../../../../assets-discover/shared/types/net-scan.type";

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
      });
    }
  }

  onAgentSelect($event: AgentType) {
    this.websocketCommand.reason = '';
    this.hasNoReason = true;
    this.agent = $event;
  }
}
