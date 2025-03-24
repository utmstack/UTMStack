import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {ModalService} from '../../../../core/modal/modal.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {UtmModulesEnum} from '../../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-agent-install-selector',
  template: `
    <div class="flex-container mt-2 mb-3">
      <ng-select [items]="platforms"
                 bindLabel="name"
                 placeholder="Select architecture"
                 [(ngModel)]="selectedPlatform"
                 class="flex-item">
      </ng-select>
      <ng-select *ngIf="protocols.length > 0"
                [items]="protocols"
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
    <ng-container *ngIf="selectedPlatform && selectedAction">
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

export class AgentInstallSelectorComponent {

  @Input() protocols = [];

  actions = [
    {id: 1, name: 'INSTALL', action: 'install'},
    {id: 2, name: 'UNINSTALL', action: 'uninstall'}
  ];

  @Input() platforms = [];

  @Input() agent: string;

  _selectedProtocol: any;
  _selectedPlatform: any;
  _selectedAction: any;
  module = UtmModulesEnum;

  constructor(private modalService: ModalService) {
  }

  get command() {
    return this.selectedPlatform[this.selectedAction.action];
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

  onChangeAction(action: any) {
    console.log(action);
    if (this.selectedPlatform &&  action.name === 'UNINSTALL') {
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
