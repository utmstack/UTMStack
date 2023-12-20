/*
Return CSS class depending of chart type passed as param in const
Example:
UTM_CHART_ICONS[LINE_CHART] -- return 'utm-icon-line'
 */
export const UTM_CHART_ICONS: { [key: string]: string } = {
  LINE_CHART: 'utm-icon-line',
  AREA_CHART: 'utm-icon-area-line',
  PIE_CHART: 'utm-icon-pie',
  VERTICAL_BAR_CHART: 'utm-icon-bar',
  HORIZONTAL_BAR_CHART: 'utm-icon-bar-horizontal',
  TAG_CLOUD_CHART: 'utm-icon-wordcloud',
  TABLE_CHART: 'utm-icon-table',
  LIST_CHART: 'utm-icon-table',
  GAUGE_CHART: 'utm-icon-gauge',
  GOAL_CHART: 'utm-icon-goal',
  METRIC_CHART: 'utm-icon-metric',
  HEATMAP_CHART: 'utm-icon-heatmap',
  COORDINATE_MAP_CHART: 'utm-icon-marker',
  TEXT_CHART: 'utm-icon-text',
};
