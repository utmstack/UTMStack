export interface ElasticHealthStatsType {
  cpuPercent: number;
  diskAvailable: number;
  diskTotal: number;
  diskUsed: number;
  diskUsedPercent: number;
  heapCurrent: number;
  heapMax: number;
  heapPercent: number;
  ramCurrent: number;
  ramMax: number;
  ramPercent: number;
  name?: string;
}
