import {ChangeDetectorRef, Component, EventEmitter, Input, OnChanges, OnDestroy, OnInit, Output, SimpleChanges} from '@angular/core';
import {Subject} from 'rxjs';
import {debounceTime, finalize, takeUntil, tap} from 'rxjs/operators';
import {debounceTime, finalize, map, takeUntil, tap} from 'rxjs/operators';
import {ModalService} from '../../../core/modal/modal.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
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
export class IntGenericGroupConfigComponent implements OnInit, OnDestroy {
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
  inputChange$ = new Subject<{id: number, value: string}>();
  destroy$ = new Subject<void>();
  uniqueConfigNameConstrain = false;
  invalidDomainOrIp = false;
  savingConfigs = new Map<number, boolean>();

  constructor(private utmModuleGroupService: UtmModuleGroupService,
              private toast: UtmToastService,
              private utmModuleGroupConfService: UtmModuleGroupConfService,
              private modalService: ModalService,
              private moduleChangeStatusBehavior: ModuleChangeStatusBehavior,
              private collectorService: UtmModuleCollectorService,
              public configFactory: IntegrationConfigFactory,
              private cdr: ChangeDetectorRef) {
  }

  ngOnInit() {
    this.config = this.configFactory.getConfiguration(this.groupType);
    this.getGroups().subscribe();
    this.changes = {keys: [], moduleId: this.moduleId};
    this.btnTittle = this.groupType === GroupTypeEnum.TENANT ?
      'Add tenant' : 'Add collector';
    this.groupName = this.groupType === GroupTypeEnum.TENANT ? 'tenant' : 'collector';

    this.inputChange$
      .pipe(
        takeUntil(this.destroy$),
        debounceTime(400))
      .subscribe(event => {
        this.uniqueConfigNameConstrain = this.uniqueConfigNameConstrain = this.groups
          .filter(g => g.id !== event.id)
          .some(group =>
              group.moduleGroupConfigurations.some(config =>
               config.confOptions === 'unique' && config.confValue.toLowerCase() === event.value.toLowerCase()
              )
        );
      });

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
    this.savingConfigs.set(group.id, true);
    const configs = this.changes.keys.filter(change => change.groupId === group.id);

    this.utmModuleGroupConfService.update({
      moduleId: group.moduleId,
      keys: configs
    }).pipe(
      finalize(() => {
        this.savingConfigs.set(group.id, false);
        this.cdr.detectChanges();
      })
    ).subscribe({
      next: () => {
        this.pendingChanges = false;
        this.changes = { keys: [], moduleId: this.moduleId };
        this.configValidChange.emit(this.tenantGroupConfigValid());
        this.toast.showSuccessBottom('Configuration saved successfully');
      },
      error: err => {
        if (err.status === 400) {
          this.toast.showError('Invalid Configuration',
            'The configuration data is invalid. Please check your inputs and try again.');
        } else {
          this.toast.showError('Error saving configuration',
            'Error while trying to save tenant configuration, please try again.');
        }
      }
    });
  }

  tenantIsConfigValid(group: UtmModuleGroupType): boolean {

    const required = group.moduleGroupConfigurations.filter(value => value.confRequired === true && this.isVisible(value));
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
    if (integrationConfig.confKey === 'ipAddress') {
      this.invalidDomainOrIp = !this.isValidDomainOrIp(integrationConfig.confValue);
    }
    this.pendingChanges = true;
    const index = this.changes.keys
                            .findIndex(value =>
                              value.confName === integrationConfig.confName && value.groupId === integrationConfig.groupId);

    const config = {
      ...integrationConfig,
      confOptions: integrationConfig.confOptions ? JSON.stringify(integrationConfig.confOptions) : integrationConfig.confOptions,
      confVisibility: integrationConfig.confVisibility ? JSON.stringify(integrationConfig.confVisibility) :
        integrationConfig.confVisibility,
    };

    if (index === -1) {
      this.changes.keys.push(config);
    } else {
      this.changes.keys[index].confValue = config.confValue;
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

  /*onFileUpload($event: any[], group: UtmModuleGroupType, integrationConfig: UtmModuleGroupConfType) {
    if (integrationConfig.confKey === 'jsonKey') {
      this.encryptService.encrypt(JSON.stringify($event[0])).subscribe(response => {
        const indexGroup = this.groups.findIndex(value => value.id === group.id);
        const indexConf = this.groups[indexGroup].moduleGroupConfigurations.findIndex(value => value.id === integrationConfig.id);
        this.groups[indexGroup].moduleGroupConfigurations[indexConf].confValue = response.body;
        this.addChange(integrationConfig);
      });
    }
  }*/

  onFileUpload($event: any[], group: UtmModuleGroupType, integrationConfig: UtmModuleGroupConfType) {
    if (integrationConfig.confKey === 'jsonKey') {
      const json = JSON.stringify($event[0]);

      const indexGroup = this.groups.findIndex(value => value.id === group.id);
      const indexConf = this.groups[indexGroup].moduleGroupConfigurations.findIndex(value => value.id === integrationConfig.id);

      this.groups[indexGroup].moduleGroupConfigurations[indexConf].confValue = json;

      this.addChange(integrationConfig);
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
    this.savingConfig = true;
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
        (this.config as CollectorConfiguration).deleteAllConfigs(collector.id)
          .subscribe({
            next: () => {
              if (this.groups.length === 1) {
                this.moduleChangeStatusBehavior.setStatus(false, false);
              }

              this.toast.showSuccessBottom('Collector configuration deleted successfully');
              this.getGroups().subscribe();
            },
            error: () => {
              this.toast.showError('Error deleting collector configuration',
                'An error occurred while trying to delete the collector configuration. Please try again.');
            }
          });
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

  isValidDomainOrIp(value: string): boolean {
    const ipRegex = /^(?:\d{1,3}\.){3}\d{1,3}$/;
    const domainRegex = /^(?!:\/\/)([a-zA-Z0-9-_]+\.)+[a-zA-Z]{2,}$/;
    return ipRegex.test(value) || domainRegex.test(value);
  }

  configHasChanges(id: any): boolean {
    return this.changes.keys.some(change => change.groupId === id);
  }

  isVisible(integrationConfig: UtmModuleGroupConfType): boolean {

    if (!integrationConfig.confVisibility) {
      return true;
    }

    const group = this.groups.find(g => g.id === integrationConfig.groupId);

    return group.moduleGroupConfigurations.some(c => c.confKey === integrationConfig.confVisibility.dependsOn
      && integrationConfig.confVisibility.values.includes(c.confValue));
  }

  onChange($event: { value: string; }, integrationConfig: UtmModuleGroupConfType) {

    const group = this.groups.find(g => g.id === integrationConfig.groupId);

    group.moduleGroupConfigurations.map(g => {
        if (g.confVisibility && g.confVisibility.dependsOn === integrationConfig.confKey
          && $event.value !== g.confValue) {
          g.confValue = null;
        }
    });

    this.addChange(integrationConfig);
  }


  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
