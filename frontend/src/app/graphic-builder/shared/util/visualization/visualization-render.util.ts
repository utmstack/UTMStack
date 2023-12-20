import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {ElasticFilterDefaultTime} from '../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';

/**
 * Resolve default time for visualization
 * @param visualization VisualizationType
 */
export function resolveDefaultVisualizationTime(visualization: VisualizationType): ElasticFilterDefaultTime {
  const indexTime = visualization.filterType.findIndex(value => value.field === '@timestamp');
  if (indexTime !== -1) {
    const from = visualization.filterType[indexTime].value[0];
    const to = visualization.filterType[indexTime].value[1];
    return new ElasticFilterDefaultTime(from, to);
  } else {
    return null;
  }
}
