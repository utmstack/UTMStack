import {AssetGroupType} from '../../asset-groups/shared/type/asset-group.type';
import {AssetDiscoveryTypeEnum} from '../enums/asset-discovery-type.enum';
import {AssetSeverityEnum} from '../enums/asset-severity.enum';
import {AssetStatusEnum} from '../enums/asset-status.enum';
import {AssetType} from './asset-type';
import {UtmDataInputStatus} from './data-source-input.type';
import {NetScanMetricsType} from './net-scan-metrics.type';
import {NetScanPortsType} from './net-scan-ports.type';
import {NetScanSoftwares} from './net-scan-softwares';

export class NetScanType {
  assetAddresses: string;
  assetAliases: string;
  assetAlias: string;
  assetAlive: boolean;
  assetIp: string;
  assetMac: string;
  assetName: string;
  assetOs: string;
  assetSeverity: AssetSeverityEnum;
  assetSeverityMetric: number;
  assetStatus: AssetStatusEnum;
  assetType: AssetType;
  group?: AssetGroupType;
  discoveredAt: Date;
  id: number;
  modifiedAt: Date;
  ports: NetScanPortsType[];
  assetNotes: string;
  metrics: NetScanMetricsType;
  softwares: NetScanSoftwares[];
  registeredMode: AssetDiscoveryTypeEnum;
  agent: boolean;
  dataInputList: UtmDataInputStatus[];
  assetOsPlatform?: string;
  assetOsMinorVersion?: string;
  assetOsMajorVersion?: string;
}
