import * as echarts from 'echarts'

// UTMStack series palette — ported verbatim from the legacy dashboards
// (UTM_COLOR_THEME) so charts keep the same recognizable colors.
export const UTM_PALETTE = [
  '#03A9F4', '#FF7043', '#EC407A', '#8BC34A', '#FF9800', '#795548', '#777777',
  '#607D8B', '#42A5F5', '#EF5350', '#66BB6A', '#009688', '#26C6DA', '#AB47BC',
  '#7E57C2', '#2196F3', '#F44336', '#4CAF50', '#F4511E', '#00BCD4', '#E91E63',
  '#9C27B0', '#673AB7', '#3F51B5', '#5C6BC0', '#29B6F6', '#26A69A', '#9CCC65',
  '#FFA726', '#8D6E63', '#888888', '#78909C',
]

function buildTheme(isDark: boolean) {
  const label = isDark ? '#94a3b8' : '#475569'
  const axisLine = isDark ? '#475569' : '#cbd5e1'
  const splitLine = isDark ? 'rgba(148,163,184,0.16)' : 'rgba(100,116,139,0.15)'

  const axisCommon = {
    axisLine: { show: true, lineStyle: { color: axisLine } },
    axisTick: { show: false },
    axisLabel: { color: label, fontSize: 11 },
    splitLine: { show: true, lineStyle: { color: splitLine, type: 'dashed' as const } },
  }

  return {
    color: UTM_PALETTE,
    backgroundColor: 'transparent',
    textStyle: { fontFamily: 'inherit', color: label },
    title: { textStyle: { color: label }, subtextStyle: { color: label } },
    // Category axis: no vertical grid lines; value axis: dashed horizontal lines.
    categoryAxis: { ...axisCommon, splitLine: { show: false } },
    valueAxis: { ...axisCommon },
    logAxis: { ...axisCommon },
    timeAxis: { ...axisCommon },
    legend: { textStyle: { color: label } },
    tooltip: {
      backgroundColor: isDark ? '#0f172a' : '#ffffff',
      borderColor: isDark ? '#1e293b' : '#e2e8f0',
      borderWidth: 1,
      textStyle: { color: isDark ? '#e2e8f0' : '#0f172a', fontSize: 12 },
    },
  }
}

export const UTM_THEME_LIGHT = 'utm-light'
export const UTM_THEME_DARK = 'utm-dark'

// Register once at module load. echarts-for-react shares this same echarts
// singleton, so the theme names resolve when passed via the `theme` prop.
echarts.registerTheme(UTM_THEME_LIGHT, buildTheme(false))
echarts.registerTheme(UTM_THEME_DARK, buildTheme(true))
