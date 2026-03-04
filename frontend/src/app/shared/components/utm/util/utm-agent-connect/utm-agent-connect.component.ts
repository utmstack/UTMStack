import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {NetScanType} from '../../../../../assets-discover/shared/types/net-scan.type';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';
import {IncidentCommandType} from '../../../../types/incident/incident-command.type';
import {Observable, of} from "rxjs";
import {catchError, map} from "rxjs/operators";

@Component({
  selector: 'app-utm-agent-connect',
  templateUrl: './utm-agent-connect.component.html',
  styleUrls: ['./utm-agent-connect.component.css']
})
export class UtmAgentConnectComponent implements OnInit, OnChanges {
  @Input() hostname: string;
  @Input() asset: NetScanType;
  @Input() websocketCommand: IncidentCommandType;
  agent$: Observable<AgentType>;
  connectToAgent = false;
  hasNoReason = false;

  constructor(private agentManagerService: UtmAgentManagerService) {
  }

  ngOnInit() {
    this.hasNoReason = this.websocketCommand.reason === '' || !this.websocketCommand.reason;
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (this.hostname) {
    this.agent$ =  this.agentManagerService.getAgent(this.hostname)
      .pipe(
        map(response => response.body),
        catchError(() => this.assetTypeToAgentType() )
      );
    }
  }

  onAgentSelect($event: AgentType) {
    this.websocketCommand.reason = '';
    this.hasNoReason = true;
    this.agent$ = of($event);
  }

  assetTypeToAgentType() {
    return of({
      ip: this.asset.assetIp,
      hostname: this.asset.assetName,
      os: this.asset.assetOs,
      status: !this.asset.status ? AgentStatusEnum.ONLINE : AgentStatusEnum.OFFLINE,
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
    });
  }
}
