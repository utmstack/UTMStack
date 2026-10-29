import { describe, expect, it } from 'vitest'
import { applyInjectedGuideMetadata, INJECTED_MODULES, mergeInjectedIntegrations } from './injectedModules'

describe('frontend-injected integration guides', () => {
  it('keeps guide-only catalog entries available with unique stable ids', () => {
    const ids = INJECTED_MODULES.map((integration) => integration.id)
    expect(new Set(ids).size).toBe(ids.length)
    expect(INJECTED_MODULES.map((integration) => integration.moduleName)).toContain('SYSLOG')
  })

  it('provides a demonstrative integration for every supported guide channel', () => {
    const channels = new Set(
      INJECTED_MODULES
        .map((integration) => integration.guideChannel)
        .filter((channel) => channel !== undefined),
    )
    expect(channels).toEqual(new Set(['syslog', 'netflow', 'windows-agent', 'linux-agent']))
  })

  it('assigns every injected integration an explicit ingest type', () => {
    expect(INJECTED_MODULES.every((integration) => integration.ingestType)).toBe(true)
    expect(INJECTED_MODULES.find((integration) => integration.moduleName === 'NGINX')?.ingestType).toBe('agent')
    expect(INJECTED_MODULES.find((integration) => integration.moduleName === 'IIS')?.ingestType).toBe('agent')
    expect(INJECTED_MODULES.find((integration) => integration.moduleName === 'SYSLOG')?.ingestType).toBe('forwarder')
    expect(INJECTED_MODULES.find((integration) => integration.moduleName === 'NGINX')?.kind).toBe('agents & syslog')
  })

  it('uses the supported Syslog and NetFlow endpoints', () => {
    const syslog = INJECTED_MODULES.find((integration) => integration.moduleName === 'SYSLOG')
    const netflow = INJECTED_MODULES.find((integration) => integration.moduleName === 'NETFLOW')
    expect(syslog).toMatchObject({
      dataType: 'syslog',
      forwarderDataType: 'syslog',
      defaultPort: '7014/tcp',
      logFilterField: 'dataSource',
    })
    expect(netflow).toMatchObject({
      dataType: 'netflow',
      forwarderDataType: 'netflow',
      defaultPort: '2055/udp',
      logFilterField: 'log.exporter',
    })
  })

  it('adds guide metadata to backend rows without replacing backend-owned fields', () => {
    const syslog = INJECTED_MODULES.find((integration) => integration.moduleName === 'SYSLOG')!
    const backendIntegration = {
      ...syslog,
      id: 'backend-syslog-id',
      dataType: 'backend-syslog-type',
      ingestType: 'collector',
      forwarderDataType: undefined,
      defaultPort: '6514/tcp',
      guideChannel: undefined,
    }

    expect(applyInjectedGuideMetadata(backendIntegration)).toMatchObject({
      id: 'backend-syslog-id',
      dataType: 'backend-syslog-type',
      ingestType: 'collector',
      forwarderDataType: 'syslog',
      defaultPort: '6514/tcp',
      guideChannel: 'syslog',
      guideOnly: true,
    })
  })

  it('leaves integrations without matching guide metadata unchanged', () => {
    const integration = { ...INJECTED_MODULES[0], moduleName: 'UNMAPPED_INTEGRATION' }
    expect(applyInjectedGuideMetadata(integration)).toBe(integration)
  })

  it('deduplicates backend and injected rows case-insensitively while keeping backend identity', () => {
    const injectedSyslog = INJECTED_MODULES.find((integration) => integration.moduleName === 'SYSLOG')!
    const backendSyslog = {
      ...injectedSyslog,
      id: 'backend-syslog-id',
      moduleName: 'syslog',
      dataType: 'backend-syslog-type',
    }

    const merged = mergeInjectedIntegrations([backendSyslog])
    const syslogEntries = merged.filter((integration) => integration.moduleName?.toUpperCase() === 'SYSLOG')

    expect(syslogEntries).toHaveLength(1)
    expect(syslogEntries[0]).toMatchObject({
      id: 'backend-syslog-id',
      dataType: 'backend-syslog-type',
      forwarderDataType: 'syslog',
    })
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
