import {NetScanMetricsType} from '../../../shared/types/net-scan-metrics.type';

export interface AssetGroupType {
  id?: number;
  groupName: string;
  groupDescription?: string;
  metrics?: NetScanMetricsType;
  createdDate?: Date;
  type?: string;
}
