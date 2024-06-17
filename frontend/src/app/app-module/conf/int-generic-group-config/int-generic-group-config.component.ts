import { HttpResponse } from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {forkJoin, of} from 'rxjs';
import {catchError, finalize, map, switchMap, tap} from 'rxjs/operators';
import {ModalService} from '../../../core/modal/modal.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {EncryptService} from '../../../shared/services/util/encrypt.service';
import {ModuleChangeStatusBehavior} from '../../shared/behavior/module-change-status.behavior';
import {IntCreateGroupComponent} from '../../shared/components/int-create-group/int-create-group.component';
import {GroupTypeEnum} from '../../shared/enum/group-type.enum';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {UtmModuleCollectorService} from '../../shared/services/utm-module-collector.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupService} from '../../shared/services/utm-module-group.service';
import {UtmListCollectorType} from '../../shared/type/utm-list-collector-type';
import {UtmModuleCollectorType} from '../../shared/type/utm-module-collector.type';
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
  @Input() groupType = GroupTypeEnum.TENANT;
  @Input() allowAdd = true;
  @Input() editable = true;
  @Output() configValidChange = new EventEmitter<boolean>();
  groups: UtmModuleGroupType[];
  loading = true;
  pendingChanges = false;
  changes: { keys: UtmModuleGroupConfType[], moduleId: number };
  savingConfig = false;
  btnTittle: string;
  GroupTypeEnum = GroupTypeEnum;
  collectors: any[];
  collectorList: UtmModuleCollectorType[] = [];
  configs: UtmModuleGroupConfType[] = [];
  groupName: string;

  constructor(private utmModuleGroupService: UtmModuleGroupService,
              private toast: UtmToastService,
              private encryptService: EncryptService,
              private utmModuleGroupConfService: UtmModuleGroupConfService,
              private modalService: ModalService,
              private moduleChangeStatusBehavior: ModuleChangeStatusBehavior,
              private collectorService: UtmModuleCollectorService, ) {
  }

  ngOnInit() {
    this.getGroups().subscribe();
    this.changes = {keys: [], moduleId: this.moduleId};
    this.btnTittle = this.groupType === GroupTypeEnum.TENANT ?
      'Add tenant' : 'Add collector';
    this.groupName = this.groupType === GroupTypeEnum.TENANT ? 'tenant' : 'collector';

  }

  getGroups() {
    this.loading = true;
    return this.utmModuleGroupService.query({ moduleId: this.moduleId }).pipe(
        tap(response => {
          this.groups = response.body;
          this.configValidChange.emit(this.tenantGroupConfigValid());
        }),
        switchMap(response => {
          if (this.groupType === GroupTypeEnum.COLLECTOR) {
            return this.collectorService.query({ module: UtmModulesEnum.AS_400 }).pipe(
                map((response: HttpResponse<UtmListCollectorType>) => {
                  response.body.collectors = response.body.collectors.filter(c => c.status === 'ONLINE');
                  return response;
                }),
                tap((response: HttpResponse<UtmListCollectorType>) => {
                  this.collectors = this.collectorService.getCollectorGroupConfig(this.groups, response.body.collectors);
                  this.collectorList = response.body.collectors;
                }),
                catchError(error => {
                  this.toast.showError('Error listing collectors',
                      'An error occurred while trying to list collectors. Please try again.');

                  return of(null);
                })
            );
          } else {
            return of(null);
          }
        }),
        catchError(error => {
          this.toast.showError('Error Listing collector configurations',
              'An error occurred while trying to list collector configurations. Please try again.');
          return of(null);
        }),
        finalize(() => this.loading = false)
    );
  }

  createGroup() {
    this.getGroups().subscribe(res => {
      const modal = this.modalService.open(IntCreateGroupComponent, {centered: true});
      modal.componentInstance.moduleId = this.moduleId;
      modal.componentInstance.groupType = this.groupType;
      modal.componentInstance.collectors = this.collectorService.getCollectors(this.groups, this.collectorList);

      modal.componentInstance.groupChange.subscribe(group => {
        this.getGroups().subscribe();
      });
    });
  }

  editGroup(group: UtmModuleGroupType) {
    if (this.groupType === GroupTypeEnum.COLLECTOR) {
      group.collector = this.collectorList.find( c => c.id === Number(group.collector)).hostname;
    }
    const modal = this.modalService.open(IntCreateGroupComponent, {centered: true});
    modal.componentInstance.moduleId = this.moduleId;
    modal.componentInstance.group = group;
    modal.componentInstance.groupType = this.groupType;
    modal.componentInstance.collectors = this.collectors;
    modal.componentInstance.groupChange.subscribe(groupResponse => {
      this.getGroups().subscribe();
    });
  }

  deleteGroup(group: UtmModuleGroupType) {
    const deleteModal = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModal.componentInstance.header = this.groupType === GroupTypeEnum.TENANT ? 'Delete tenant' : 'Delete collector';
    deleteModal.componentInstance.message = `By deleting ${group.groupName} ` +
        `${this.groupName}, UTMStack will no longer receive logs from this source. ` +
        `${this.groups.length === 1 ? ` Since this is the only ${this.groupName}, the module associated with it will be deactivated.` : ''} ` +
        ` Are you sure that you want to delete this ${this.groupName}`;

    deleteModal.componentInstance.confirmBtnText = 'Delete';
    deleteModal.componentInstance.confirmBtnIcon = 'icon-stack-cancel';
    deleteModal.componentInstance.confirmBtnType = 'delete';
    deleteModal.result.then(() => {
      this.deleteAction(group);
    });
  }

  deleteAction(group: UtmModuleGroupType) {
    this.utmModuleGroupService.delete(group.id).subscribe(response => {
      if (this.groups.length === 1) {
        this.moduleChangeStatusBehavior.setStatus(false);
      } else {
        this.toast.showSuccessBottom(`${this.groupName.toUpperCase()} group deleted successfully`);
      }
      this.getGroups().subscribe();
    }, error => {
      this.toast.showError(`Error deleting ${this.groupName}`,
        `Error while trying to delete ${this.groupName} , please try again`);
    });
  }

  saveConfig(group: UtmModuleGroupType) {
    this.savingConfig = true;
    const configs = this.changes.keys.filter(change => change.groupId === group.id);
    this.utmModuleGroupConfService.update({
      moduleId: group.moduleId,
      keys: configs
    }).subscribe(response => {
      this.savingConfig = false;
      this.pendingChanges = false;
      this.changes = {keys: [], moduleId: this.moduleId};
      this.configValidChange.emit(this.tenantGroupConfigValid());
      this.toast.showSuccessBottom('Configuration saved successfully');
    }, () => {
      this.toast.showError('Error saving configuration',
        'Error while trying to save tenant configuration , please try again');
    });
  }

  tenantIsConfigValid(group: UtmModuleGroupType): boolean {

    const required = group.moduleGroupConfigurations.filter(value => value.confRequired === true);
    const valid = required.filter(value => value.confValue !== null && value.confValue);

    if (this.groupType === GroupTypeEnum.COLLECTOR) {
      return valid.length === required.length && !this.validateUniqueHostNameByCollector(group);
    } else {
      return valid.length === required.length;
    }
  }

  tenantGroupConfigValid(): boolean {
    let required = [];
    let valid = [];
    this.groups.forEach((group) => {
      required = group.moduleGroupConfigurations.filter(value => value.confRequired === true);
      valid = required.filter(value => value.confValue !== null && value.confValue);
    });
    return this.groups.length > 0 && valid.length === required.length;
  }

  checkConfigValue(config: string): boolean {
    return config !== null && config !== '' && config !== undefined;
  }

  addChange(integrationConfig: UtmModuleGroupConfType) {
    this.pendingChanges = true;
    const index = this.changes.keys
      .findIndex(value => value.confName === integrationConfig.confName &&
        value.groupId === integrationConfig.groupId);
    if (index === -1) {
      this.changes.keys.push(integrationConfig);
    } else {
      this.changes.keys[index].confValue = integrationConfig.confValue;
    }
    // this.configValidChange.emit(this.tenantGroupConfigValid());
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

  pendingChangesForCollector(groups: UtmModuleGroupType[]): boolean {
    return groups.some(group => this.pendingChangesForGroup(group));
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

  collectorValid(groups: UtmModuleGroupType[]) {
    return groups.every(group => this.tenantIsConfigValid(group));
  }

  collectorConfigValid(groups: UtmModuleGroupType[]) {
    return groups.some(group => this.tenantIsConfigValid(group));
  }

  cancelCollectorConfig(groups: UtmModuleGroupType []) {
   groups.forEach(group => this.cancelConfig(group));
  }

  saveCollectorConfig(collector: any) {
    const collectorDto = this.collectorList.find(c => c.hostname === collector.collector);
    this.configs = [];
    collector.groups.forEach((item: { moduleGroupConfigurations: any; }) => {
      const configurations = item.moduleGroupConfigurations;
      this.configs.push(...configurations);
    });
    const body = {
      collectorConfig: {
        moduleId: this.moduleId,
        keys: this.configs
      },
      collector: {
        ... collectorDto,
        group: null,
      },
    };
    this.collectorService.create(body).subscribe(response => {
      this.savingConfig = false;
      this.pendingChanges = false;
      this.changes = {keys: [], moduleId: this.moduleId};
      this.configValidChange.emit(this.tenantGroupConfigValid());
      this.toast.showSuccessBottom('Configuration saved successfully');
    }, () => {
      this.toast.showError('Error saving configuration',
          'Error while trying to save collector configuration , please try again');
    });
  }

  addConfig(collector: any) {
    const col = this.collectorList.find(c => c.hostname === collector.collector);
    this.utmModuleGroupService.create({
      description: '',
      moduleId: this.moduleId,
      name: this.collectorService.generateUniqueName(collector.collector, collector.groups),
      collector: col.id
    }).subscribe(response => {
      this.getGroups().subscribe();
    }, () => {
    });
  }

  deleteCollectorConfig(collector: any) {
    const deleteModal = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModal.componentInstance.header = 'Delete collector';
    deleteModal.componentInstance.message = 'By deleting this collector, UTMStack will no longer receive logs from this source.' +
      (this.groups.length === 1 ? ' Since this is the only collector, the module associated with it will be deactivated.' : '') +
      ' Are you sure that you want to delete this collector?';

    deleteModal.componentInstance.confirmBtnText = 'Delete';
    deleteModal.componentInstance.confirmBtnIcon = 'icon-stack-cancel';
    deleteModal.componentInstance.confirmBtnType = 'delete';
    deleteModal.result.then(() => {
        const deleteRequests = collector.groups.map( g => this.utmModuleGroupService.delete(g.id));

        forkJoin(deleteRequests).subscribe(() => {
          this.moduleChangeStatusBehavior.setStatus(false);
          this.getGroups().subscribe();
        },
            (error) => {
              this.toast.showError('Error saving configuration',
                  'Error processing deletion requests , please try again');
            });

    });
  }

  validateUniqueHostNameByCollector(group: UtmModuleGroupType) {
      const configs = [];

      const config = group.moduleGroupConfigurations.find(c => c.confName === 'Hostname');
      const groups = this.groups.filter(g => g.collector === g.collector && g.id !== group.id);

      groups.forEach((item: { moduleGroupConfigurations: any; }) => {
        const configurations = item.moduleGroupConfigurations;
        configs.push(...configurations);
      });

      return configs.some(c => c.confName === 'Hostname' && c.confValue === config.confValue);
  }

  showClose(group: UtmModuleGroupType, groups: UtmModuleGroupType[]) {

    const start = groups.length - 1;
    const index = groups.findIndex(g => g.id === group.id);

    if (group.id === groups[start].id) {
      return true;
    } else if (groups.length === 2) {
      return this.tenantIsConfigValid(groups[index + 1]) && !this.pendingChangesForGroup(groups[index + 1]);
    } else {
      return true;
    }
  }
}
