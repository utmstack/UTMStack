import {VisualizationType} from '../chart/types/visualization.type';
import {ALERT_ROUTE, LOG_ROUTE} from '../constants/app-routes.constant';
import {IndexPatternSystemEnumName} from '../enums/index-pattern-system.enum';

/**
 * Resolve link to go after click chart, if eventType is EVENT or VULNERABILITY got to
 * log analyzer section; else go to alert management
 * @param visualization VisualizationType, visualization clicked
 */
export function resolveChartNavigationUrl(visualization: VisualizationType): string {
  if (visualization.pattern.pattern === IndexPatternSystemEnumName.ALERT) {
    return ALERT_ROUTE;
  } else {
    return LOG_ROUTE;
  }
}
