import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';
import {IncidentCommandType} from '../../../../types/incident/incident-command.type';

@Component({
  selector: 'app-utm-console-check',
  templateUrl: './utm-console-check.component.html',
  styleUrls: ['./utm-console-check.component.css']
})
export class UtmConsoleCheckComponent implements OnInit {
  @Input() hostname: string;
  @Input() websocketCommand: IncidentCommandType;
  @Output() emptyValue = new EventEmitter<boolean>(false);

  connect = false;
  canConnect = false;
  agent: AgentType;
  agentStatusEnum = AgentStatusEnum;

  constructor(private agentManagerService: UtmAgentManagerService,) {
  }

  ngOnInit() {
    this.getAgent(this.hostname).then(value => {
      this.agent = value;
      this.canConnect = this.agent.status === AgentStatusEnum.ONLINE;
      this.emptyValue.emit(!this.canConnect);
    });
  }

  getAgent(hostname: string): Promise<AgentType> {
    return new Promise<AgentType>((resolve, reject) => {
        this.agentManagerService.getAgent(hostname).subscribe(response => {
          resolve(response.body);
        }, () => reject());
      }
    );
  }

  connectConsole() {
    this.connect = true;
  }
}
