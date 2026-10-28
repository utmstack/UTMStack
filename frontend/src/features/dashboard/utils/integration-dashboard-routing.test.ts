import { describe, expect, it } from 'vitest'
import { getIntegrationDashboardSearch } from './integration-dashboard-routing'

describe('integration dashboard routing', () => {
  it('opens the dashboard gallery filtered to Windows dashboards', () => {
    expect(getIntegrationDashboardSearch({
      moduleName: 'WINDOWS_AGENT',
      dataType: 'wineventlog',
      name: 'Windows Agent',
    })).toBe('Windows')
  })

  it('uses a broad Linux filter so multiple related dashboards remain visible', () => {
    expect(getIntegrationDashboardSearch({
      moduleName: 'LINUX_AGENT',
      dataType: 'linux',
      name: 'Linux Agent',
    })).toBe('Linux')
  })

  it('filters other integrations by their display name', () => {
    expect(getIntegrationDashboardSearch({
      moduleName: 'NETFLOW',
      dataType: 'netflow',
      name: 'NetFlow Exporter',
    })).toBe('NetFlow Exporter')
  })
})
