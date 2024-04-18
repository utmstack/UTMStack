import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {UtmModulesEnum} from '../../../shared/enum/utm-module.enum';
import {ModalService} from '../../../../core/modal/modal.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';

@Component({
  selector: 'app-log-colletor',
  template: `
    <div class="flex-container mt-2 mb-3">
      <ng-select [items]="platforms"
                 bindLabel="name"
                 placeholder="Select platform"
                 [(ngModel)]="selectedPlatform"
                 class="flex-item">
      </ng-select>
      <ng-select [items]="protocols"
                 bindLabel="name"
                 placeholder="Select Protocol"
                 [(ngModel)]="selectedProtocol"
                 class="flex-item">
      </ng-select>
      <ng-select [items]="actions"
                 (change)="onChangeAction($event)"
                 bindLabel="name"
                 placeholder="Select Action"
                 [(ngModel)]="selectedAction"
                 class="flex-item">
      </ng-select>
    </div>
    <ng-container *ngIf="selectedProtocol && selectedPlatform && selectedAction">
      <span class="font-weight-semibold mb-2">{{selectedPlatform.shell}}</span>
      <app-utm-code-view class="" [code]=command></app-utm-code-view>
    </ng-container>
  `,
  styles: [`
    .flex-container {
      display: flex;
    }

    .flex-item {
      flex-grow: 1;
      margin-right: 10px;
    }
  `],
  changeDetection: ChangeDetectionStrategy.OnPush
})

export class LogCollectorComponent {

  @Input() protocols = [
    {id: 1, name: 'TCP'},
    {id: 2, name: 'UDP'}
  ];

  actions = [
    {id: 1, name: 'ENABLE', action: 'enable-integration'},
    {id: 2, name: 'DISABLE', action: 'disable-integration'}
  ];

  platforms = [
    {
      id: 1, name: 'WINDOWS',
      command: 'Start-Process "C:\\Program Files\\UTMStack\\UTMStack Agent\\utmstack_agent_service.exe" -ArgumentList \'ACTION\', \'AGENTNAME\', \'PORT\' -NoNewWindow -Wait\n',
      shell: 'Windows Powershell terminal as “ADMINISTRATOR”'
    },
    {
      id: 2,
      name: 'LINUX', command: 'sudo bash -c "/opt/utmstack-linux-agent/utmstack_agent_service ACTION AGENTNAME PORT"',
      shell: 'Linux bash terminal'
    }
  ];

  @Input() agent: string;

  _selectedProtocol: any;
  _selectedPlatform: any;
  _selectedAction: any;
  module = UtmModulesEnum;

  constructor(private modalService: ModalService) {
  }

  get command() {
    this.selectedPlatform.command.replace(/PORT/gi, this.selectedProtocol.name.toLowerCase());
    return this.replaceAll(this.selectedPlatform.command, {
      PORT: this.selectedProtocol.name.toLowerCase(),
      AGENTNAME: this.agentName(),
      ACTION: this.selectedAction.action
    });
  }

  get selectedPlatform() {
    return this._selectedPlatform;
  }

  set selectedPlatform(platform) {
    this._selectedPlatform = platform;
  }

  get selectedProtocol() {
    return this._selectedProtocol;
  }

  set selectedProtocol(protocol) {
    this._selectedProtocol = protocol;
  }

  get selectedAction() {
    return this._selectedAction;
  }

  set selectedAction(action) {
    this._selectedAction = action;
  }

  agentName() {
    switch (this.agent) {

      case UtmModulesEnum.VMWARE:
        return 'vmware';

      case UtmModulesEnum.SYSLOG:
        return 'syslog';

      case UtmModulesEnum.SONIC_WALL:
        return 'firewall_sonicwall';

      case UtmModulesEnum.SOPHOS_XG:
        return 'firewall_sophos';

      case UtmModulesEnum.PFSENSE:
        return 'firewall_pfsense';

      case UtmModulesEnum.PALO_ALTO:
        return 'firewall_paloalto';

      case UtmModulesEnum.MIKROTIK:
        return 'firewall_mikrotik';

      case UtmModulesEnum.FORTIGATE:
        return 'firewall_fortinet';

      case UtmModulesEnum.SENTINEL_ONE:
        return 'antivirus_sentinel_one';

      case UtmModulesEnum.FORTIWEB:
        return 'firewall_fortiweb';

      case UtmModulesEnum.AIX:
        return 'aix';

      case UtmModulesEnum.ESET:
        return 'antivirus_eset';

      case UtmModulesEnum.KASPERSKY:
        return 'antivirus_kaspersky';

      case UtmModulesEnum.MACOS:
        return 'macos_logs';

      case UtmModulesEnum.DECEPTIVE_BYTES:
        return 'antivirus_deceptivebytes';

      case UtmModulesEnum.NETFLOW:
        return 'netflow';

      case UtmModulesEnum.FIRE_POWER:
      case UtmModulesEnum.CISCO:
      case UtmModulesEnum.CISCO_SWITCH:
      case UtmModulesEnum.MERAKI:
        return 'cisco';
    }
  }

  replaceAll(command, wordsToReplace) {
    return Object.keys(wordsToReplace).reduce(
      (f, s, i) =>
        `${f}`.replace(new RegExp(s, 'ig'), wordsToReplace[s]),
      command
    );
  }

  onChangeAction(action: any) {
    if (this.selectedPlatform && this.selectedProtocol && action.name === 'DISABLE') {
      this.openModal();
    }
  }

  openModal() {
    const modalSource = this.modalService.open(ModalConfirmationComponent, {centered: true});

    modalSource.componentInstance.header = 'Disable integration command';
    modalSource.componentInstance.message = 'The following command is to disable the integration. ' +
                                             'Running this command could cause irreversible damage to your infrastructure. ' +
                                             'Only continue if you are sure what you are doing and really want to disable the integration.';
    modalSource.componentInstance.confirmBtnText = 'Accept';
    modalSource.componentInstance.confirmBtnIcon = 'icon-cog3';
    modalSource.componentInstance.confirmBtnType = 'default';
    modalSource.componentInstance.hideBtnCancel = true;
    modalSource.result.then(() => {

    });
  }
}
