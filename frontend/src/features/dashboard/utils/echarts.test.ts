import { describe, expect, it } from 'vitest'
import { defaultChartSkeleton, mergeRowsIntoOption } from './echarts'
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

  // mergeRowsIntoOption never invents a series — a config with no series/axes
  // of its own merges into a bare legend, nothing plotted. This is what the
  // seeded "Agents" dashboard's config actually was ({ __builder: { chartType:
  // "bar" } }, no series) before this was fixed: real, non-empty rows, chart
  // rendered with nothing in it.
  it('a config with no series draws nothing, even with real rows', () => {
    const option = mergeRowsIntoOption({}, [{ severity: 'critical', count: 12 }])
    expect(option.series).toBeUndefined()
  })
})

describe('defaultChartSkeleton', () => {
  it("uses the builder's chart type when one was saved", () => {
    expect(defaultChartSkeleton('pie', 'category')).toEqual(getChartTypeMeta('pie').defaultConfig)
  })

  it('falls back to a bar skeleton for a category spec with no builder info', () => {
    expect(defaultChartSkeleton(undefined, 'category')).toEqual(getChartTypeMeta('bar').defaultConfig)
  })

  it('falls back to a line skeleton for a time spec with no builder info', () => {
    expect(defaultChartSkeleton(undefined, 'time')).toEqual(getChartTypeMeta('line').defaultConfig)
  })

  it('is empty when neither a builder chart type nor a plottable spec chart is known', () => {
    expect(defaultChartSkeleton(undefined, undefined)).toEqual({})
    expect(defaultChartSkeleton(undefined, 'metric')).toEqual({})
  })

  // The actual bug: a bootstrap-seeded widget whose config only ever
  // recorded { __builder: { chartType: "bar" } } — no series, no axes —
  // now draws a real bar chart once its skeleton is filled in first.
  it('recovers a real chart from a config that only named its chart type', () => {
    const skeleton = defaultChartSkeleton('bar', 'category')
    const option = mergeRowsIntoOption(skeleton, [
      { severity: 'critical', count: 12 },
      { severity: 'high', count: 34 },
    ])
    expect(option.series).toBeDefined()
    expect(option.dataset).toMatchObject({
      source: [
        { severity: 'critical', count: 12 },
        { severity: 'high', count: 34 },
      ],
    })
  })
})
