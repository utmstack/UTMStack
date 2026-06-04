import {Component, Input, OnInit} from '@angular/core';
import {AgentStatusEnum, AgentType} from '../../../../shared/types/agent/agent.type';

@Component({
  selector: 'app-agent-info',
  templateUrl: './agent-info.component.html',
  styleUrls: ['./agent-info.component.scss']
})
export class AgentInfoComponent implements OnInit {
  @Input() agent: AgentType;

  constructor() { }

  ngOnInit() {
  }

  statusClass(status: AgentStatusEnum): string {
    switch (status) {
      case AgentStatusEnum.ONLINE:
        return 'badge bg-success';
      case AgentStatusEnum.OFFLINE:
        return 'badge bg-secondary';
      default:
        return 'badge bg-warning';
    }
  }

  platformIcon(platform: string): string {
    switch ((platform || '').toLowerCase()) {
      case 'windows': return 'icon-windows';
      case 'linux': return 'icon-linux';
      case 'darwin': return 'icon-apple'; // macOS
      default: return 'icon-server';
    }
  }

  platformLabel(platform: string): string {
    return platform === 'darwin' ? 'macOS' : (platform || 'Unknown');
  }

}
