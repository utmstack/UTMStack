import { afterEach, describe, expect, it, vi } from 'vitest'
import { buildDatasourceReportHtml, filterDatasourceReport } from './report'
import type { Datasource } from '../types/datasource.types'

const now = Date.now()
const sources: Datasource[] = [
  { id: 'live', name: 'Live source', lastPingAt: new Date(now - 60_000).toISOString() },
  { id: 'idle', name: 'Idle source', lastPingAt: new Date(now - 60 * 60_000).toISOString() },
  { id: 'offline', name: 'Offline source', lastPingAt: new Date(now - 48 * 60 * 60_000).toISOString() },
  { id: 'never-seen', name: 'Never seen' },
]

describe('datasource reports', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('filters a report to the selected liveness state and respects the Offline tab', () => {
    expect(filterDatasourceReport(sources, 'all').map(({ id }) => id)).toEqual([
      'live',
      'idle',
      'offline',
      'never-seen',
    ])
    expect(filterDatasourceReport(sources, 'live').map(({ id }) => id)).toEqual(['live'])
    expect(filterDatasourceReport(sources, 'idle').map(({ id }) => id)).toEqual(['idle'])
    expect(filterDatasourceReport(sources, 'offline').map(({ id }) => id)).toEqual(['offline', 'never-seen'])
    expect(filterDatasourceReport(sources, 'all', true).map(({ id }) => id)).toEqual(['offline', 'never-seen'])
  })

  it('includes all report columns and escapes source-controlled content', () => {
    const labels = {
      title: 'Data Sources Report',
      generatedAt: 'Generated',
      sourceCount: 'Sources: {{count}}',
      noSources: 'No matching data sources.',
      columns: {
        name: 'Name',
        type: 'Type',
        kind: 'Kind',
        status: 'Status',
        ip: 'IP',
        lastSeen: 'Last seen',
        discovered: 'Discovered',
        labels: 'Labels',
      },
      statuses: { live: 'Live', idle: 'Idle', offline: 'Offline' },
      kinds: { agent: 'Agent', puller: 'Integration', direct: 'Direct' },
    }

    const html = buildDatasourceReportHtml(
      [{ id: '1', name: '<img src=x onerror=alert(1)>', dataType: 'wineventlog', sourceKind: 'agent', labels: 'classified' }],
      labels,
      '2026-01-01 10:00',
    )

    expect(html).toContain('Sources: 1')
    expect(html).toContain('<th>Name</th>')
    expect(html).toContain('&lt;img src=x onerror=alert(1)&gt;')
    expect(html).not.toContain('<img src=x')
  })
})
