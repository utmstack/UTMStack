import {ElasticHealthStatsType} from '../types/elasticsearch/elastic-health-stats.type';

export function getElasticClusterHealth(clusterHealth: ElasticHealthStatsType,
                                        connection: 'UP' | 'DOWN'): 'CRITIC' | 'MEDIUM' | 'UP' | 'DOWN' {
  if (clusterHealth) {
    if (clusterHealth.diskUsedPercent > 84) {
      return 'CRITIC';
    }
    if (clusterHealth.heapPercent > 85) {
      return 'CRITIC';
    }
    if (clusterHealth.diskUsedPercent > 75 && this.clusterHealth.diskUsedPercent <= 85) {
      return 'MEDIUM';
    } else if (clusterHealth.diskUsedPercent <= 50) {
      return 'UP';
    }
    if (clusterHealth.heapPercent >= 50 && this.clusterHealth.heapPercent <= 85) {
      return 'MEDIUM';
    } else if (clusterHealth.heapPercent < 50) {
      return 'UP';
    } else {
      return connection;
    }
  } else {
    return connection;
  }
}

