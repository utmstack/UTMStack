import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {map, tap} from 'rxjs/operators';
import {ModalService} from '../../../core/modal/modal.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {EncryptService} from '../../../shared/services/util/encrypt.service';
import {ModuleChangeStatusBehavior} from '../../shared/behavior/module-change-status.behavior';
import {IntCreateGroupComponent} from '../../shared/components/int-create-group/int-create-group.component';
import {GroupTypeEnum} from '../../shared/enum/group-type.enum';
import {UtmModuleCollectorService} from '../../shared/services/utm-module-collector.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupService} from '../../shared/services/utm-module-group.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';
import {UtmModuleGroupType} from '../../shared/type/utm-module-group.type';
import {CollectorConfiguration} from './int-config-types/collector-configuration';
import {IntegrationConfig} from './int-config-types/integration-config';
import {IntegrationConfigFactory} from './int-config-types/IntegrationConfigFactory';

@Component({
  selector: 'app-int-generic-group-config',
  templateUrl: './int-generic-group-config.component.html',
  styleUrls: ['./int-generic-group-config.component.css']
})
export class IntGenericGroupConfigComponent implements OnInit, OnChanges {
  @Input() serverId: number;
  @Input() moduleId: number;
  @Input() groupType = GroupTypeEnum.TENANT;
  @Input() allowAdd = true;
  @Input() editable = true;
  @Input() disablePreAction = false;
  @Output() configValidChange = new EventEmitter<boolean>();
  @Output() runDisablePreAction = new EventEmitter<boolean>();
  loading = true;
  pendingChanges = false;
  changes: { keys: UtmModuleGroupConfType[], moduleId: number };
  savingConfig = false;
  btnTittle: string;
  GroupTypeEnum = GroupTypeEnum;
  configs: UtmModuleGroupConfType[] = [];
  groupName: string;
  config: IntegrationConfig;

  constructor(private utmModuleGroupService: UtmModuleGroupService,
              private toast: UtmToastService,
              private encryptService: EncryptService,
              private utmModuleGroupConfService: UtmModuleGroupConfService,
              private modalService: ModalService,
              private moduleChangeStatusBehavior: ModuleChangeStatusBehavior,
              private collectorService: UtmModuleCollectorService,
              public configFactory: IntegrationConfigFactory) {
  }

  ngOnInit() {
    this.config = this.configFactory.getConfiguration(this.groupType);
    this.getGroups().subscribe();
    this.changes = {keys: [], moduleId: this.moduleId};
    this.btnTittle = this.groupType === GroupTypeEnum.TENANT ?
      'Add tenant' : 'Add collector';
    this.groupName = this.groupType === GroupTypeEnum.TENANT ? 'tenant' : 'collector';

  }

  get groups() {
    return this.config.groups;
  }

  get collectors() {
    return (this.config as CollectorConfiguration).collectors;
  }

  get collectorList() {
    return (this.config as CollectorConfiguration).collectorList;
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes.disablePreAction && changes.disablePreAction.currentValue) {
      const collectors = [];
      this.collectorList.forEach( c => {
        collectors.push({
          moduleId: this.moduleId,
          keys: this.configs,
          collector: {
            ...c,
            group: null,
          }
        });
      });
      this.collectorService.bulkCreate(collectors)
          .pipe(map(response => response.body.results))
          .subscribe( results => {
            if (results.every( r => r.status === 'success')) {
              this.moduleChangeStatusBehavior.setStatus(false, false);
              this.getGroups().subscribe();
            } else {
              this.toast.showError('Error Removing Collector Configuration',
                  'An error occurred while trying to remove the collector configuration. Please try again.');
            }
          },
          error => console.log(error));
    }
  }

  getGroups() {
    this.loading = true;
    return this.config.getIntegrationConfigs(this.moduleId)
        .pipe(
            tap(() => {
              this.configValidChange.emit(this.tenantGroupConfigValid());
              this.loading = false;
            }));
  }

  createGroup() {
    this.getGroups().subscribe(res => {
      const modal = this.modalService.open(IntCreateGroupComponent, {centered: true});
      modal.componentInstance.moduleId = this.moduleId;
      modal.componentInstance.groupType = this.groupType;
      modal.componentInstance.collectors = this.groupType === GroupTypeEnum.COLLECTOR ?
          this.collectorService.getCollectors(this.groups, this.collectorList) : [];

      modal.componentInstance.groupChange.subscribe(group => {
        this.getGroups().subscribe();
      });
    });
  }

  editGroup(group: UtmModuleGroupType) {
    const modal = this.modalService.open(IntCreateGroupComponent, {centered: true});
    modal.componentInstance.moduleId = this.moduleId;
    modal.componentInstance.group = group;
    modal.componentInstance.groupType = this.groupType;
    modal.componentInstance.collectors = this.groupType === GroupTypeEnum.COLLECTOR ? this.collectors : [];
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

  deleteAction(param: any) {
    this.config.deleteIntegrationConfigs(param)
        .subscribe(response => {
          if (this.groups.length === 1) {
            this.moduleChangeStatusBehavior.setStatus(false, false);
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
      return valid.length === required.length && !((this.config as CollectorConfiguration).validateUniqueHostNameByCollector(group));
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
                            .findIndex(value =>
                              value.confName === integrationConfig.confName && value.groupId === integrationConfig.groupId);
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

  saveCollectorConfig(collector: any, action = 'CREATE') {
    (this.config as CollectorConfiguration).saveCollector(collector)
      .subscribe(response => {
      this.savingConfig = false;
      this.pendingChanges = false;
      this.changes = {keys: [], moduleId: this.moduleId};
      this.configValidChange.emit(this.tenantGroupConfigValid());
      this.toast.showSuccessBottom('Configuration saved successfully');
      this.activeModule();
    }, (err) => {
        if (err.status === 502) {
          this.toast.showWarning('The configuration was saved, but there was an issue sending it to the service. ' +
            'It will be applied once the service is back online.',
            'Configuration saved');
          this.activeModule();
        } else {
          this.toast.showError('Error saving configuration',
            'Error while trying to save collector configuration , please try again');
        }
    });
  }
  addConfig(collector: any) {
    const col = this.collectorList.find(c => c.hostname === collector.collector);
    const name = this.collectorService.generateUniqueName(collector.collector, collector.groups);
    this.utmModuleGroupService.create({
      description: `Description for: ${name}`,
      moduleId: this.moduleId,
      name,
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
      if (collector && collector.collector !== '') {
        const collectorToSave = {
          ...collector,
          groups: []
        };
        this.deleteAction(collectorToSave);
      }
    });
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

  activeModule() {
    this.moduleChangeStatusBehavior.setStatus(!this.moduleChangeStatusBehavior.getLastStatus() ? true : null, true);
  }
}
