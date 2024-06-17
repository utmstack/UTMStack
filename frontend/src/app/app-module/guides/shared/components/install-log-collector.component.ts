import {ChangeDetectionStrategy, Component, Input} from '@angular/core';
import {ModalService} from '../../../../core/modal/modal.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {UtmModulesEnum} from '../../../shared/enum/utm-module.enum';
import {generate} from "rxjs";

@Component({
  selector: 'app-install-log-collector',
  template: `
    <div class="flex-container mt-2 mb-3">
      <ng-select [items]="platforms"
                 bindLabel="name"
                 placeholder="Select platform"
                 [(ngModel)]="selectedPlatform"
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
      <app-utm-code-view class="" [code]="generateCommand()"></app-utm-code-view>
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

export class InstallLogCollectorComponent {

  @Input() platforms: [];
  @Input() actions: [];
  @Input() vars;

  _selectedProtocol: any;
  _selectedPlatform: any;
  _selectedAction: any;
  module = UtmModulesEnum;

  constructor(private modalService: ModalService) {
  }

  generateCommand() {
    return this.replaceAll(this.getCommand());
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

  getCommand() {
    return this.selectedAction.name === 'INSTALL' ? this.selectedPlatform.install : this.selectedPlatform.uninstall;
  }

  replaceAll(command: string) {
    return Object.keys(this.vars).reduce(
      (f, s, i) =>
        `${f}`.replace(new RegExp(s, 'ig'), this.vars[s]),
      command
    );
  }

  onChangeAction(action: any) {
    if (this.selectedPlatform && action.name === 'UNINSTALL') {
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

  protected readonly generate = generate;
}
