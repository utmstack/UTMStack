import { describe, expect, it } from 'vitest'
import { mergeRowsIntoOption } from './echarts'
import { getChartTypeMeta } from '@/features/dashboard/constants'

describe('drawing the answer', () => {
  it('feeds the rows to the chart as its dataset', () => {
    const option = mergeRowsIntoOption(getChartTypeMeta('bar').defaultConfig, [
      { name: 'Port scan detected', count: 153 },
    ])
    expect(option.dataset).toMatchObject({
      source: [{ name: 'Port scan detected', count: 153 }],
    })
  })

  it('draws one line per column when a timeline is split', () => {
    const rows = [
      { at: '2026-07-30', 'aws-cloudtrail': 177, windows: 0 },
      { at: '2026-07-31', 'aws-cloudtrail': 1136, windows: 42 },
    ]
    const option = mergeRowsIntoOption(getChartTypeMeta('line').defaultConfig, rows)

    const series = option.series as { name: string; encode: { x: number; y: number } }[]
    expect(series).toHaveLength(2)
    expect(series.map((s) => s.name)).toEqual(['aws-cloudtrail', 'windows'])
    // Every line reads the same x column and its own value column.
    expect(series.map((s) => s.encode.y)).toEqual([1, 2])
    expect(new Set(series.map((s) => s.encode.x))).toEqual(new Set([0]))
  })

  it('leaves a single-series chart alone', () => {
    const option = mergeRowsIntoOption(getChartTypeMeta('line').defaultConfig, [
      { at: '2026-07-30', count: 906 },
    ])
    expect(option.series).toHaveLength(1)
  })
})
