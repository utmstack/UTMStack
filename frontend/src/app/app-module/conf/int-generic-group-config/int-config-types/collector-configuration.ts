import {HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, finalize, map, switchMap, tap} from 'rxjs/operators';
import {ModalService} from '../../../../core/modal/modal.service';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {EncryptService} from '../../../../shared/services/util/encrypt.service';
import {ModuleChangeStatusBehavior} from '../../../shared/behavior/module-change-status.behavior';
import {UtmModulesEnum} from '../../../shared/enum/utm-module.enum';
import {UtmModuleCollectorService} from '../../../shared/services/utm-module-collector.service';
import {UtmModuleGroupConfService} from '../../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupService} from '../../../shared/services/utm-module-group.service';
import {UtmListCollectorType} from '../../../shared/type/utm-list-collector-type';
import {UtmModuleCollectorType} from '../../../shared/type/utm-module-collector.type';
import {UtmModuleGroupType} from '../../../shared/type/utm-module-group.type';
import {IntegrationConfig} from './integration-config';

@Injectable()
export class CollectorConfiguration extends IntegrationConfig {

    private _collectors: any[];
    private _collectorList: UtmModuleCollectorType[] = [];

    constructor(private utmModuleGroupService: UtmModuleGroupService,
                private toast: UtmToastService,
                private encryptService: EncryptService,
                private utmModuleGroupConfService: UtmModuleGroupConfService,
                private modalService: ModalService,
                private moduleChangeStatusBehavior: ModuleChangeStatusBehavior,
                private collectorService: UtmModuleCollectorService) {
        super();
    }

    getIntegrationConfigs(moduleId: number) {
        return this.utmModuleGroupService.query({moduleId}).pipe(
            tap(response => {
                this.groups = response.body;
                this.moduleId = moduleId;
            }),
            switchMap(response => {
                return this.collectorService.query({module: UtmModulesEnum.AS_400})
                    .pipe(
                        map((res: HttpResponse<UtmListCollectorType>) => {
                            res.body.collectors = res.body.collectors.filter(c => c.status === 'ONLINE');
                            return res.body;
                        }),
                        tap((data: UtmListCollectorType) => {
                            this.collectors = this.collectorService.getCollectorGroupConfig(this.groups, data.collectors);
                            this.collectorList = data.collectors;
                            this.moduleChangeStatusBehavior.setStatus(null, true);
                        }),
                        catchError(error => {
                            this.toast.showError('Error listing collectors',
                                'An error occurred while trying to list collectors. Please try again.');
                            return of(null);
                        })
                    );
            }),
            catchError(error => {
                this.toast.showError('Error listing collector configurations',
                    'An error occurred while trying to list collector configurations. Please try again.');
                return of(null);
            })
        );
    }

    deleteIntegrationConfigs(group: UtmModuleGroupType): Observable<any> {
        const collector = this.collectors.find(c => c.id === parseInt(group.collector, 10));
        if (collector && collector.collector !== '') {
            group = {
                ...collector,
                groups: collector.groups.filter(g => g.id !== group.id)
            };
            return this.saveCollector(group);
        } else {
            group = {
                ...group,
                collector: null
            };
            return this.utmModuleGroupService.delete(group.id);
        }
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

    saveCollector(group: any) {
        return this.collectorService.create(this.getBody(group));
    }

    getBody(collector: any, action = 'CREATE') {
        const collectorDto = this.collectorList.find(c => c.hostname === collector.collector);
        const configs = [];
        collector.groups.forEach((item: { moduleGroupConfigurations: any; }) => {
            const configurations = item.moduleGroupConfigurations;
            configs.push(...configurations);
        });
        return {
            collectorConfig: {
                moduleId: this.moduleId,
                keys: configs,
                collector: {
                    ...collectorDto,
                    group: null,
                }
            },
            action
        };
    }

    get collectors() {
        return this._collectors;
    }

    set collectors(collectors: any[]) {
        this._collectors = collectors;
    }

    get collectorList() {
        return this._collectorList;
    }

    set collectorList(collectorList: UtmModuleCollectorType[]) {
        this._collectorList = collectorList;
    }
}
