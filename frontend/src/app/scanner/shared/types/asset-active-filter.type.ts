export class AssetActiveFilterType {
  'created.greaterThan'?: string;
  'created.lessThan'?: string;
  'hostIp.contains'?: string;
  'hostOs.contains'?: string;
  'hostSeverity.equals'?: number | string;
  'hostSeverity.greaterThan'?: number;
  'hostSeverity.lessThan'?: number;
  'name.contains'?: string;
  'minQod.greaterThan': number;
  task: string;
  host?: string;
  limit?: number;
  reportId: string;
}
