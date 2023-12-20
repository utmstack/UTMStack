import {UtmModulesEnum} from '../enum/utm-module.enum';
import {UtmModuleGroupType} from './utm-module-group.type';

export class UtmModuleType {
  configGroups: UtmModuleGroupType;
  id: number;
  integrationPath: string;
  moduleActive: boolean;
  moduleCategory: string;
  moduleDescription: string;
  moduleIcon: string;
  moduleName: UtmModulesEnum;
  needsRestart: true;
  prettyName: string;
  serverId: number;
  activatable: boolean;
}
