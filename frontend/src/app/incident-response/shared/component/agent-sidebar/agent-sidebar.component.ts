import { Component, OnInit } from '@angular/core';
import {AgentSidebarService} from './agent-sidebar.service';
import {AgentStatusEnum} from "../../../../shared/types/agent/agent.type";

@Component({
  selector: 'app-agent-sidebar',
  templateUrl: './agent-sidebar.component.html',
  styleUrls: ['./agent-sidebar.component.scss']
})
export class AgentSidebarComponent implements OnInit {
  searching: any;
  request = {
    page: 0,
    size: 5,
  };
  AgentStatusEnum = AgentStatusEnum;

  constructor(public agentSidebarService: AgentSidebarService) { }

  ngOnInit() {
    this.agentSidebarService.loadData({
      ...this.request
    });
  }

  searchAgent($event: string | number) {
    this.agentSidebarService.loadData({
      ...this.request,
      page: 0,
      searchQuery: $event.toString()
    });
  }

  onScroll() {

  }

  trackByFn(index: number, item: any) {
    return item.id || index;
  }

  agentDetail(action: any) {
    this.agentSidebarService.selectAgent(action);
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
