import ReactECharts from 'echarts-for-react'

export function EChartsRenderer({ option }: { option: Record<string, unknown> }) {
  return (
    <ReactECharts
      option={option}
      notMerge
      lazyUpdate
      style={{ height: '100%', width: '100%' }}
    />
  )
}
