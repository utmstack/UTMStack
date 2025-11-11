import {Component, Input, OnInit} from '@angular/core';
import {EMPTY} from 'rxjs';
import {catchError, map, tap} from 'rxjs/operators';
import {NetScanType} from '../../../../../assets-discover/shared/types/net-scan.type';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {UtmAgentManagerService} from '../../../../services/agent/utm-agent-manager.service';
import {AgentStatusEnum, AgentType} from '../../../../types/agent/agent.type';

@Component({
  selector: 'app-utm-agent-detail',
  templateUrl: './utm-agent-detail.component.html',
  styleUrls: ['./utm-agent-detail.component.scss']
})
export class UtmAgentDetailComponent implements OnInit {
  @Input() agent: AgentType;
  @Input() asset: NetScanType;
  agentStatusEnum = AgentStatusEnum;
  ips: string[];
  macs: string[];
  agentIp: string;
  loading = false;

  constructor(private agentManagerService: UtmAgentManagerService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.agentIp = this.agent.ip;
    this.ips = this.agent.addresses !== '' ? this.agent.addresses.split(',') : [];
    this.macs = this.agent.mac !== '' ? this.agent.mac.split(',') : [];
  }

  public getAgentIcon(): string {
    if (this.agent && this.agent.os) {
      if (this.agent.os.toLowerCase().includes('windows')) {
        return 'icon-windows8';
      } else if (this.agent.os.toLowerCase().includes('linux')) {
        return 'icon-tux';
      } else if (this.agent.os.toLowerCase().includes('mac')) {
        return 'icon-apple2';
      } else if (this.agent.os.toLowerCase().includes('possible conflict')) {
        return 'icon-power2';
      }
    }
    return 'icon-question3';
  }

  getAssets() {

  }

  changeIp(event: string) {
    this.loading = true;
    const agent = {
      hostname: this.agent.hostname,
      ip: this.agentIp,
    };

    this.agentManagerService.updateAgent(agent)
      .pipe(
        map(response => response.body),
        tap(response => this.loading = false),
        catchError(error => {
          this.utmToastService.showError('Error',
            'An error occurred while updating the agent ip. Please try again later.');
          this.loading = false;
          this.agentIp = this.agent.ip;
          return EMPTY;
        }))
      .subscribe(() => this.agent.ip = this.agentIp);
  }
}
