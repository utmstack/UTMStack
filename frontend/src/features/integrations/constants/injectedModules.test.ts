import { describe, expect, it } from 'vitest'
import { INJECTED_MODULES } from './injectedModules'

describe('frontend-injected integration guides', () => {
  it('keeps guide-only catalog entries available with unique stable ids', () => {
    const ids = INJECTED_MODULES.map((integration) => integration.id)
    expect(new Set(ids).size).toBe(ids.length)
    expect(INJECTED_MODULES.map((integration) => integration.moduleName)).toContain('SYSLOG')
  })

  it('provides a demonstrative integration for every supported guide channel', () => {
    const channels = new Set(INJECTED_MODULES.map((integration) => integration.guideChannel))
    expect(channels).toEqual(new Set(['syslog', 'netflow', 'windows-agent', 'linux-agent']))
  })

  it('uses the supported Syslog and NetFlow endpoints', () => {
    const syslog = INJECTED_MODULES.find((integration) => integration.moduleName === 'SYSLOG')
    const netflow = INJECTED_MODULES.find((integration) => integration.moduleName === 'NETFLOW')
    expect(syslog).toMatchObject({ dataType: 'syslog', defaultPort: '7014/tcp', logFilterField: 'dataSource' })
    expect(netflow).toMatchObject({ dataType: 'netflow', defaultPort: '2055/udp', logFilterField: 'log.exporter' })
  })

  it('filters nested Windows and Linux product fields with is-one-of', () => {
    const sqlServer = INJECTED_MODULES.find((integration) => integration.moduleName === 'WINDOWS_SQL_SERVER')
    const nginx = INJECTED_MODULES.find((integration) => integration.moduleName === 'NGINX')
    expect(sqlServer?.logFilters).toContainEqual({
      field: 'log.providerName',
      operator: 'IS_ONE_OF_TERMS',
      value: ['MSSQLSERVER'],
    })
    expect(nginx?.logFilters).toContainEqual({
      field: 'log.syslogIdentifier',
      operator: 'IS_ONE_OF_TERMS',
      value: ['nginx'],
    })
  })
})
