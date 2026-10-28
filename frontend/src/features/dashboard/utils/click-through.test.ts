import { describe, expect, it } from 'vitest'
import type { VizSpec } from '@/features/dashboard/types'
import { specFiltersToLogFilters, termFilter, timeBucketRange, widgetBaseFilters } from './click-through'

describe('dashboard click-through filters', () => {
  it('uses term matching for single nested numeric and text values', () => {
    const spec: VizSpec = {
      dataset: 'logs',
      dataType: 'wineventlog',
      chart: 'category',
      dimension: 'log.eventCode',
      metric: { agg: 'count' },
      filters: [{ field: 'origin.user', op: 'eq', value: 'alice' }],
    }

    expect(widgetBaseFilters(spec)).toEqual([
      { field: 'dataType', operator: 'IS_ONE_OF_TERMS', value: ['wineventlog'] },
      { field: 'origin.user', operator: 'IS_ONE_OF_TERMS', value: ['alice'] },
    ])
  })

  it('preserves filters that cannot be represented as URL parameters', () => {
    expect(
      specFiltersToLogFilters([
        { field: 'target.user', op: 'not_in', value: ['service', 'guest'] },
        { field: 'origin.ip', op: 'exists' },
      ]),
    ).toEqual([
      { field: 'target.user', operator: 'IS_NOT_ONE_OF_TERMS', value: ['service', 'guest'] },
      { field: 'origin.ip', operator: 'EXIST' },
    ])
  })

  it('uses terms matching for a nested numeric chart value', () => {
    expect(termFilter('log.eventCode', 4624)).toEqual({
      field: 'log.eventCode',
      operator: 'IS_ONE_OF_TERMS',
      value: [4624],
    })
  })

  it('turns a rendered hourly label into an absolute one-hour range', () => {
    const range = timeBucketRange('2026-10-25 06:00', '1h')
    expect(range).not.toBeNull()
    expect(new Date(range!.to).getTime() - new Date(range!.from!).getTime()).toBe(60 * 60 * 1000 - 1)
  })
})
