import { describe, expect, it } from 'vitest'
import type { FilterType } from '../types/log-explorer.types'
import { resolveDashboardDataType } from './dashboard-search'

const dataTypeFilter: FilterType = {
  field: 'dataType',
  operator: 'IS_ONE_OF_TERMS',
  value: ['wineventlog'],
}
const eventCodeFilter: FilterType = {
  field: 'log.eventCode',
  operator: 'IS_ONE_OF_TERMS',
  value: ['4670', '5145'],
}

describe('dashboard search data type', () => {
  it('moves a single dataType filter into the selector', () => {
    expect(resolveDashboardDataType(
      [dataTypeFilter, eventCodeFilter],
      undefined,
      ['wineventlog', 'linux'],
    )).toEqual({ dataType: 'wineventlog', filters: [eventCodeFilter] })
  })

  it('prefers the widget dataType over the dashboard filter', () => {
    expect(resolveDashboardDataType(
      [dataTypeFilter, eventCodeFilter],
      'linux',
      ['wineventlog', 'linux'],
    )).toEqual({ dataType: 'linux', filters: [eventCodeFilter] })
  })

  it('keeps all filters and selects all data types when no type is sent', () => {
    expect(resolveDashboardDataType(
      [eventCodeFilter],
      undefined,
      ['wineventlog', 'linux'],
    )).toEqual({ dataType: null, filters: [eventCodeFilter] })
  })

  it('preserves a multi-value or unavailable dataType filter instead of widening results', () => {
    const multiTypeFilter = { ...dataTypeFilter, value: ['wineventlog', 'linux'] }
    expect(resolveDashboardDataType(
      [multiTypeFilter, eventCodeFilter],
      undefined,
      ['wineventlog', 'linux'],
    )).toEqual({ dataType: null, filters: [multiTypeFilter, eventCodeFilter] })
  })
})
