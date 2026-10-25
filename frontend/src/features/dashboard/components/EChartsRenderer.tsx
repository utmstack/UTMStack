import ReactECharts from 'echarts-for-react'
import { useThemeContext } from '@/app/providers/ThemeProvider'
import {
  UTM_THEME_DARK,
  UTM_THEME_LIGHT,
} from '@/features/dashboard/utils/echarts-theme'

export interface ChartClickEvent {
  dataIndex?: number
  seriesName?: string
}

export function EChartsRenderer({
  option,
  onMarkClick,
}: {
  option: Record<string, unknown>
  onMarkClick?: (event: ChartClickEvent) => void
}) {
  const { theme } = useThemeContext()
  return (
    <ReactECharts
      option={option}
      theme={theme === 'dark' ? UTM_THEME_DARK : UTM_THEME_LIGHT}
      notMerge
      lazyUpdate
      onEvents={
        onMarkClick
          ? {
              click: (event: ChartClickEvent) => onMarkClick(event),
            }
          : undefined
      }
      style={{ height: '100%', width: '100%' }}
    />
  )
}
