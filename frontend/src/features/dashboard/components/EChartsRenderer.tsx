import ReactECharts from 'echarts-for-react'
import { useThemeContext } from '@/app/providers/ThemeProvider'
import {
  UTM_THEME_DARK,
  UTM_THEME_LIGHT,
} from '@/features/dashboard/utils/echarts-theme'

export function EChartsRenderer({ option }: { option: Record<string, unknown> }) {
  const { theme } = useThemeContext()
  return (
    <ReactECharts
      option={option}
      theme={theme === 'dark' ? UTM_THEME_DARK : UTM_THEME_LIGHT}
      notMerge
      lazyUpdate
      style={{ height: '100%', width: '100%' }}
    />
  )
}
