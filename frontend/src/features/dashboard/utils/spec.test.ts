import { describe, expect, it } from 'vitest'
import { builderToSpec, resultToRows, specIsComplete } from './spec'
import { makeInitialBuilder } from './builder-config'
import type { BuilderState, QueryResult, VizSpec } from '@/features/dashboard/types'

// The payloads below are what the backend actually answered against the event
// store, trimmed. A widget draws whatever this turns them into, so the shape is
// the contract worth pinning.

function builder(patch: Partial<BuilderState>): BuilderState {
  return { ...makeInitialBuilder(), dataset: 'logs', ...patch }
}

describe('the question a widget asks', () => {
  it('counts records, because the store does nothing else', () => {
    const spec = builderToSpec(builder({ chartType: 'metric' }))
    expect(spec).toMatchObject({ dataset: 'logs', chart: 'metric', metric: { agg: 'count' } })
  })

  it('breaks a bar chart down by a field', () => {
    const spec = builderToSpec(
      builder({ chartType: 'bar', breakdown: 'field', dimension: 'name', limit: 5 })
    )
    expect(spec).toMatchObject({ chart: 'category', dimension: 'name', limit: 5 })
  })

  it('plots a line over time, and the field only splits it', () => {
    const spec = builderToSpec(
      builder({ chartType: 'line', breakdown: 'time', interval: '1d', dimension: 'dataType' })
    )
    expect(spec).toMatchObject({ chart: 'time', interval: '1d', dimension: 'dataType' })
  })

  it('asks a table for records, with the columns picked', () => {
    const spec = builderToSpec(builder({ chartType: 'table', columns: ['name', 'severity'] }))
    expect(spec).toMatchObject({ chart: 'table', columns: ['name', 'severity'] })
  })

  it('translates the filter vocabulary into the store operators', () => {
    const spec = builderToSpec(
      builder({
        filters: [
          { id: '1', field: 'severity', operator: 'IS_NOT', value: 'low' },
          { id: '2', field: 'name', operator: 'START_WITH', value: 'Port' },
          { id: '3', field: 'host', operator: 'EXIST', value: null },
          { id: '4', field: 'user', operator: 'IS', value: '' },
        ],
      })
    )
    expect(spec.filters).toEqual([
      { field: 'severity', op: 'not_eq', value: 'low' },
      { field: 'name', op: 'starts_with', value: 'Port' },
      { field: 'host', op: 'exists' },
    ])
  })

  it('knows a breakdown chart is unanswerable until it has a field', () => {
    const incomplete = builderToSpec(builder({ chartType: 'bar', breakdown: 'field' }))
    expect(specIsComplete(incomplete)).toBe(false)
    expect(specIsComplete({ ...incomplete, dimension: 'name' })).toBe(true)
  })
})

describe('the answer, as the renderers read it', () => {
  it('turns a count into the one row a metric shows', () => {
    const spec = { dataset: 'logs', chart: 'metric', metric: { agg: 'count' } } as VizSpec
    expect(resultToRows({ total: 48015 }, spec)).toEqual([{ value: 48015 }])
  })

  it('names a bucket column after the field it broke down by', () => {
    const spec = {
      dataset: 'alerts',
      chart: 'category',
      dimension: 'name',
      metric: { agg: 'count' },
    } as VizSpec
    const result: QueryResult = {
      buckets: [
        { key: 'Port scan detected', count: 153 },
        { key: 'Scheduled task created', count: 159 },
      ],
    }
    expect(resultToRows(result, spec)).toEqual([
      { name: 'Port scan detected', count: 153 },
      { name: 'Scheduled task created', count: 159 },
    ])
  })

  it('gives a split timeline one column per line, zero-filling the gaps', () => {
    const spec = {
      dataset: 'logs',
      chart: 'time',
      interval: '1d',
      dimension: 'dataType',
      metric: { agg: 'count' },
    } as VizSpec
    const result: QueryResult = {
      series: [
        {
          key: 'aws-cloudtrail',
          points: [
            { at: '2026-07-30T00:00:00Z', count: 177 },
            { at: '2026-07-31T00:00:00Z', count: 1136 },
          ],
        },
        // This one had no records on the first day.
        { key: 'windows', points: [{ at: '2026-07-31T00:00:00Z', count: 42 }] },
      ],
    }

    // Instants are labelled in the reader's own time zone, so the label is
    // derived rather than written down.
    const day = (iso: string) => {
      const d = new Date(iso)
      const pad = (n: number) => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    }

    const rows = resultToRows(result, spec)
    expect(rows).toHaveLength(2)
    expect(rows[0]).toEqual({
      at: day('2026-07-30T00:00:00Z'),
      'aws-cloudtrail': 177,
      windows: 0,
    })
    expect(rows[1]).toEqual({
      at: day('2026-07-31T00:00:00Z'),
      'aws-cloudtrail': 1136,
      windows: 42,
    })
  })

  it('flattens records and shows a column under its last segment', () => {
    const spec = {
      dataset: 'alerts',
      chart: 'table',
      metric: { agg: 'count' },
      columns: ['name', 'origin.geolocation.country'],
    } as VizSpec
    const result: QueryResult = {
      rows: [{ name: 'Port scan', origin: { geolocation: { country: 'RU' } }, noise: 1 }],
    }
    expect(resultToRows(result, spec)).toEqual([{ name: 'Port scan', country: 'RU' }])
  })

  it('keeps the whole path when two columns would read the same', () => {
    const spec = {
      dataset: 'alerts',
      chart: 'table',
      metric: { agg: 'count' },
      columns: ['origin.ip', 'destination.ip'],
    } as VizSpec
    const result: QueryResult = {
      rows: [{ origin: { ip: '1.1.1.1' }, destination: { ip: '2.2.2.2' } }],
    }
    expect(resultToRows(result, spec)).toEqual([
      { 'origin.ip': '1.1.1.1', 'destination.ip': '2.2.2.2' },
    ])
  })
})
