import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {ModalService} from '../../../../core/modal/modal.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {replaceCommandTokens} from '../../../../shared/util/replace-command-tokens.util';
import {UtmModulesEnum} from '../../../shared/enum/utm-module.enum';
import {PLATFORMS} from '../constant';

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
      <app-utm-code-view *ngFor="let command of commands" class="" [code]=command></app-utm-code-view>
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

  @Input() agent: string;
  @Input() platforms: any[] = PLATFORMS;
  @Input() hideActions = false;
  @Input() hideProtocols = false;
  @Input() protocols = [
    {id: 1, name: 'TCP'},
    {id: 2, name: 'TCP/TLS'},
    {id: 3, name: 'UDP'}
  ];

  actions = [
    {id: 1, name: 'ENABLE', action: 'enable-integration'},
    {id: 2, name: 'DISABLE', action: 'disable-integration'}
  ];

  _selectedProtocol: any;
  _selectedPlatform: any;
  _selectedAction: any;
  module = UtmModulesEnum;

  constructor(private modalService: ModalService) {}

  get commands() {

    const protocol = this.selectedProtocol && this.selectedProtocol.name === 'TCP/TLS' ? 'tcp' : this.selectedProtocol.name.toLowerCase();

    const command = replaceCommandTokens(this.selectedPlatform.command, {
        ACTION: this.selectedAction && this.selectedAction.action || '',
        AGENT_NAME: this.agentName(),
        PROTOCOL: protocol,
        TLS: this.selectedProtocol && this.selectedProtocol.name === 'TCP/TLS' &&
          this.selectedAction.name === 'ENABLE' ? `--tls` : ''
      });

    if (this.selectedProtocol && this.selectedProtocol.name === 'TCP/TLS' &&
      this.selectedAction.name === 'ENABLE') {
      const extras = this.selectedPlatform.extraCommands ? this.selectedPlatform.extraCommands : [];
      return [...extras, command];
    }

    return [command];
  }

  get selectedPlatform() {
    return this._selectedPlatform;
  }

  @Input()
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
        return 'ibm_aix';

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

      case UtmModulesEnum.SURICATA:
        return 'suricata';

      case UtmModulesEnum.FIRE_POWER:
      case UtmModulesEnum.CISCO:
      case UtmModulesEnum.CISCO_SWITCH:
      case UtmModulesEnum.MERAKI:
        return 'cisco';
    }
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
