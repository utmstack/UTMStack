import {AssetGroupType} from '../../../assets-discover/asset-groups/shared/type/asset-group.type';

export class UtmModuleCollectorType {
  id: number;
  status: string;
  ip: string;
  hostname: string;
  version: string;
  collectorKey: string;
  module: string;
  lastSeen: string;
  group?: AssetGroupType;
  active?: boolean;
}
