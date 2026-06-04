import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, finalize, map, switchMap, tap} from 'rxjs/operators';
import { UtmModuleGroupType } from 'src/app/app-module/shared/type/utm-module-group.type';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {UtmModuleGroupService} from '../../../shared/services/utm-module-group.service';

import {IntegrationConfig} from './integration-config';

@Injectable()
export class GenericConfiguration extends IntegrationConfig {

    constructor(private utmModuleGroupService: UtmModuleGroupService,
                private toast: UtmToastService) {
        super();
    }

    getIntegrationConfigs(moduleId: number) {
        return this.utmModuleGroupService.query({moduleId}).pipe(
            tap(response => {
              this.groups = response.body.map(group => {
                group.moduleGroupConfigurations.forEach(config => {
                  if (config.confOptions) {
                    config.confOptions = this.tryParseJSON(config.confOptions);
                  }
                  if (config.confVisibility) {
                    config.confVisibility = this.tryParseJSON(config.confVisibility);
                  }
                });
                return group;
              });
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

  tryParseJSON(value: any): any {
    try {
      return JSON.parse(value);
    } catch {
      return value;
    }
  }
}
