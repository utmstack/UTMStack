import {HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, finalize, map, switchMap, tap} from 'rxjs/operators';
import { UtmModuleGroupType } from 'src/app/app-module/shared/type/utm-module-group.type';
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
import {IntegrationConfig} from './integration-config';

@Injectable()
export class GenericConfiguration extends IntegrationConfig {

    constructor(private utmModuleGroupService: UtmModuleGroupService,
                private toast: UtmToastService,
                private encryptService: EncryptService,
                private utmModuleGroupConfService: UtmModuleGroupConfService,
                private modalService: ModalService,
                private moduleChangeStatusBehavior: ModuleChangeStatusBehavior) {
        super();
    }

    getIntegrationConfigs(moduleId: number) {
        return this.utmModuleGroupService.query({moduleId}).pipe(
            tap(response => {
                this.groups = response.body;
                this.moduleId = moduleId;
            }),
            catchError(error => {
                this.toast.showError('Error listing configurations',
                    'An error occurred while trying to list configurations. Please try again.');
                return of(null);
            })
        );
    }

    deleteIntegrationConfigs(group: UtmModuleGroupType): Observable<any> {
        return this.utmModuleGroupService.delete(group.id);
    }
}
