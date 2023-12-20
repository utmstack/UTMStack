import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ModalService} from '../../../core/modal/modal.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {EncryptService} from '../../../shared/services/util/encrypt.service';
import {IntCreateGroupComponent} from '../../shared/components/int-create-group/int-create-group.component';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupService} from '../../shared/services/utm-module-group.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';
import {UtmModuleGroupType} from '../../shared/type/utm-module-group.type';

@Component({
  selector: 'app-int-generic-group-config',
  templateUrl: './int-generic-group-config.component.html',
  styleUrls: ['./int-generic-group-config.component.css']
})
export class IntGenericGroupConfigComponent implements OnInit {
  @Input() serverId: number;
  @Input() moduleId: number;
  @Input() allowAdd = true;
  @Input() editable = true;
  @Output() configValidChange = new EventEmitter<boolean>();
  groups: UtmModuleGroupType[];
  loading = true;
  pendingChanges = false;
  changes: { keys: UtmModuleGroupConfType[], moduleId: number };
  savingConfig = false;

  constructor(private utmModuleGroupService: UtmModuleGroupService,
              private toast: UtmToastService,
              private encryptService: EncryptService,
              private utmModuleGroupConfService: UtmModuleGroupConfService,
              private modalService: ModalService) {
  }

  ngOnInit() {
    this.getGroups();
    this.changes = {keys: [], moduleId: this.moduleId};
  }

  getGroups() {
    this.utmModuleGroupService.query({moduleId: this.moduleId}).subscribe(response => {
      this.groups = response.body;
      this.loading = false;
      this.configValidChange.emit(this.tenantGroupConfigValid());
    });
  }

  createGroup() {
    const modal = this.modalService.open(IntCreateGroupComponent, {centered: true});
    modal.componentInstance.moduleId = this.moduleId;
    modal.componentInstance.groupChange.subscribe(group => {
      this.getGroups();
    });
  }

  editGroup(group: UtmModuleGroupType) {
    const modal = this.modalService.open(IntCreateGroupComponent, {centered: true});
    modal.componentInstance.moduleId = this.moduleId;
    modal.componentInstance.group = group;
    modal.componentInstance.groupChange.subscribe(groupResponse => {
      this.getGroups();
    });
  }

  deleteGroup(group: UtmModuleGroupType) {
    const deleteModal = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModal.componentInstance.header = 'Delete tenant';
    deleteModal.componentInstance.message = 'By deleting ' + group.groupName + ' tenant UTMStack no longer receive' +
      ' logs from this source.' + ' Are you sure that you want to delete this tenant?';
    deleteModal.componentInstance.confirmBtnText = 'Delete';
    deleteModal.componentInstance.confirmBtnIcon = 'icon-stack-cancel';
    deleteModal.componentInstance.confirmBtnType = 'delete';
    deleteModal.result.then(() => {
      this.deleteAction(group);
    });
  }

  deleteAction(group: UtmModuleGroupType) {
    this.utmModuleGroupService.delete(group.id).subscribe(response => {
      this.toast.showSuccessBottom('Tenant group saved successfully');
      this.getGroups();
    }, error => {
      this.toast.showError('Error deleting tenant',
        'Error while trying to delete tenant ' + group.groupName + ' , please try again');
    });
  }

  saveConfig() {
    this.savingConfig = true;
    this.utmModuleGroupConfService.update(this.changes).subscribe(response => {
      this.savingConfig = false;
      this.pendingChanges = false;
      this.changes = {keys: [], moduleId: this.moduleId};
      this.toast.showSuccessBottom('Configuration saved successfully');
    }, () => {
      this.toast.showError('Error saving configuration',
        'Error while trying to save tenant configuration , please try again');
    });
  }

  tenantIsConfigValid(group: UtmModuleGroupType): boolean {
    const required = group.moduleGroupConfigurations.filter(value => value.confRequired === true);
    const valid = required.filter(value => value.confValue !== null && value.confValue);
    return valid.length === required.length;
  }

  tenantGroupConfigValid(): boolean {
    let required = [];
    let valid = [];
    this.groups.forEach((group) => {
      required = group.moduleGroupConfigurations.filter(value => value.confRequired === true);
      valid = required.filter(value => value.confValue !== null && value.confValue);
    });
    return valid.length === required.length;
  }

  checkConfigValue(config: string): boolean {
    return config !== null && config !== '' && config !== undefined;
  }

  addChange(integrationConfig: UtmModuleGroupConfType) {
    this.pendingChanges = true;
    const index = this.changes.keys.findIndex(value => value.confName === integrationConfig.confName);
    if (index === -1) {
      this.changes.keys.push(integrationConfig);
    } else {
      this.changes.keys[index].confValue = integrationConfig.confValue;
    }
    this.configValidChange.emit(this.tenantGroupConfigValid());
  }

  cancelConfig(group: UtmModuleGroupType) {
    this.utmModuleGroupService.find(group.id).subscribe(response => {
      const index = this.groups.indexOf(group);
      if (index !== -1) {
        this.groups[index].moduleGroupConfigurations = response.body.moduleGroupConfigurations;
      }
      // tslint:disable-next-line:prefer-for-of
      for (let i = 0; i < this.changes.keys.length; i++) {
        if (this.changes.keys[i].groupId === group.id) {
          this.changes.keys.splice(i, 1);
        }
      }
      this.configValidChange.emit(this.tenantGroupConfigValid());
    });
  }

  pendingChangesForGroup(group: UtmModuleGroupType): boolean {
    return this.changes.keys.filter(value => value.groupId === group.id).length > 0;
  }

  onFileUpload($event: any[], group: UtmModuleGroupType, integrationConfig: UtmModuleGroupConfType) {
    if (integrationConfig.confKey === 'jsonKey') {
      this.encryptService.encrypt(JSON.stringify($event[0])).subscribe(response => {
        const indexGroup = this.groups.findIndex(value => value.id === group.id);
        const indexConf = this.groups[indexGroup].moduleGroupConfigurations.findIndex(value => value.id === integrationConfig.id);
        this.groups[indexGroup].moduleGroupConfigurations[indexConf].confValue = response.body;
        this.addChange(integrationConfig);
      });
    }
  }
}
