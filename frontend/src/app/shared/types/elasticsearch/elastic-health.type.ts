import {ElasticHealthStatsType} from './elastic-health-stats.type';

export interface ElasticHealthType {
  nodes: ElasticHealthStatsType[];
  resume: ElasticHealthStatsType;
}
