import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ModalService} from '../../../../core/modal/modal.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {replaceCommandTokens} from '../../../../shared/util/replace-command-tokens.util';
import {UtmModulesEnum} from '../../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-agent-action-command',
  template: `
    <div class="flex-container mt-2 mb-3">
      <ng-select [items]="platforms"
                 (change)="platformEmitter.emit($event)"
                 bindLabel="name"
                 placeholder="Select platform"
                 [(ngModel)]="selectedPlatform"
                 class="flex-item">
      </ng-select>
      <ng-select *ngIf="!hideProtocols"
                [items]="protocols"
                 bindLabel="name"
                 placeholder="Select Protocol"
                 [(ngModel)]="selectedProtocol"
                 class="flex-item">
      </ng-select>
      <ng-select *ngIf="!hideActions"
                 [items]="actions"
                 (change)="onChangeAction($event)"
                 bindLabel="name"
                 placeholder="Select Action"
                 [(ngModel)]="selectedAction"
                 class="flex-item">
      </ng-select>
    </div>
    <div *ngIf="this.selectedProtocol && this.selectedProtocol.name === 'TCP/TLS' && selectedAction"
         class="alert alert-info alert-styled-right mt-2">
      After the TLS certificates have been successfully loaded into the system,
      it is not necessary to repeat the certificate loading process when enabling
      additional integrations that use TLS. The system will automatically apply the
      previously configured certificates to ensure secure communication.
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

export class AgentActionCommandComponent implements OnInit{
  @Output() platformEmitter = new EventEmitter();
  @Input() platforms: any[];
  @Input() agent: string;
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

  constructor(private modalService: ModalService) {
  }

  ngOnInit(): void {}

  get commands() {

    const protocol = this.selectedProtocol && this.selectedProtocol.name === 'TCP/TLS' ? 'tcp' : this.selectedProtocol.name.toLowerCase();

    const command = replaceCommandTokens(this.selectedPlatform.command, {
        ACTION: this.selectedAction && this.selectedAction.action || '',
        AGENT_NAME: this.agent || '',
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
