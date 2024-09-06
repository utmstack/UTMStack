import {Observable} from 'rxjs';
import {UtmModuleGroupType} from '../../../shared/type/utm-module-group.type';

export abstract  class IntegrationConfig {
    private _moduleId;
    private _groups: UtmModuleGroupType[];
    abstract getIntegrationConfigs(moduleId: number): Observable<any>;
    abstract deleteIntegrationConfigs(group: UtmModuleGroupType): Observable<any>;

    get moduleId() {
        return this._moduleId;
    }

    set moduleId(moduleId: number) {
        this._moduleId = moduleId;
    }
    get groups() {
        return this._groups;
    }

    set groups(groups: UtmModuleGroupType[]) {
        this._groups = groups;
    }
}
