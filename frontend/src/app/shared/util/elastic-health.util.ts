import {ElasticHealthStatsType} from '../types/elasticsearch/elastic-health-stats.type';
export function getElasticClusterHealth(clusterHealth: ElasticHealthStatsType,
                                        connection: 'UP' | 'DOWN'): 'CRITIC' | 'MEDIUM' | 'UP' | 'DOWN' {
  if (clusterHealth) {
    const { diskUsedPercent, heapPercent } = clusterHealth;

    if (diskUsedPercent > 84 || heapPercent > 85) {
      return 'CRITIC';
    }

    if ((diskUsedPercent > 75 && diskUsedPercent <= 85) || (heapPercent >= 50 && heapPercent <= 85)) {
      return 'MEDIUM';
    }

    if (diskUsedPercent <= 50 && heapPercent < 50) {
      return 'UP';
    }

    return connection;

  } else {
    return connection;
  }
}

