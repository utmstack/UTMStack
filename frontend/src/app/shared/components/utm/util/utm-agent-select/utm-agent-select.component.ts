import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';

@Component({
  selector: 'app-utm-agent-select',
  templateUrl: './utm-agent-select.component.html',
  styleUrls: ['./utm-agent-select.component.css']
})
export class UtmAgentSelectComponent implements OnInit {
  @Input() onlyWithCommands = false;
  @Output() agentSelect = new EventEmitter<AgentType>();
  agents: AgentType[];
  agentStatusEnum = AgentStatusEnum;
  agent: any;

  constructor(private utmAgentManagerService: UtmAgentManagerService) {
  }

  ngOnInit() {
    if (this.onlyWithCommands) {
      this.getAgentsWithCommands();
    } else {
      this.getAgents();
    }
  }


  getAgents() {
    this.utmAgentManagerService.getAgents({pageNumber: 1, pageSize: 100000, searchQuery: '', sortBy: ''})
      .subscribe(response => {
        this.agents = response.body.filter(value => value.status === this.agentStatusEnum.ONLINE);
      });
  }

  getAgentsWithCommands() {
    this.utmAgentManagerService.getAgentsWithCommands({pageNumber: 1, pageSize: 100000, searchQuery: '', sortBy: ''})
      .subscribe(response => {
        this.agents = response.body;
      });
  }


  selectAgent($event: AgentType | any) {
    this.agent = `${$event.hostname} (${$event.os})`;
    this.agentSelect.emit($event);
  }
}
