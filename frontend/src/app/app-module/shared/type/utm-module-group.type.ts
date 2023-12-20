import {UtmModuleGroupConfType} from './utm-module-group-conf.type';
import {UtmModuleType} from './utm-module.type';

export class UtmModuleGroupType {
  groupDescription: string;
  groupName: string;
  id: number;
  module: UtmModuleType;
  moduleId: number;
  moduleGroupConfigurations: UtmModuleGroupConfType[];
}
