import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DEMO_RECORDS } from '../data/demo-records'
import { DEMO_ORGANIZATION_BREACH_STATS } from '../data/demo-organization-stats'
import { DarkWebMonitoringPage } from './DarkWebMonitoringPage'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

describe('DarkWebMonitoringPage demo', () => {
  it('keeps the demo statistics in the OrganizationBreachStats response shape', () => {
    expect(DEMO_ORGANIZATION_BREACH_STATS).toEqual({
      id: 10901,
      orgName: 'test01',
      totalEmailsBreaches: 0,
      totalPhonesBreaches: 0,
      totalDomainBreaches: 73,
      totalIpsBreaches: 0,
      totalUsernamesBreaches: 1857,
      totalBreaches: 1930,
    })
  })

  it('renders tab counts from organization breach statistics', () => {
    render(<DarkWebMonitoringPage />)

    expect(screen.getAllByRole('button', { name: /darkWebMonitoring\.tabs\.domains.*73/ }).length).toBeGreaterThan(0)
    expect(screen.getByRole('button', { name: /darkWebMonitoring\.tabs\.usernames.*1,857/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /darkWebMonitoring\.tabs\.ips.*0/ })).toBeInTheDocument()
    expect(screen.getByText('1,930')).toBeInTheDocument()
  })

  it('uses the API field names and status values in its demo records', () => {
    const record = DEMO_RECORDS[0]

    expect(record).toMatchObject({
      id: expect.any(String),
      timestamp: expect.any(String),
      createdAt: expect.any(String),
      organizationElementType: 'DOMAIN',
      breachStatus: 'IN_MITIGATION',
      employee: true,
      echoesCount: 0,
    })
    expect(record).toHaveProperty('ipAddress', null)
    expect(record).toHaveProperty('hashedPassword', null)
    expect(record).toHaveProperty('databaseName')
    expect(record).not.toHaveProperty('status')
    expect(record).not.toHaveProperty('ingestedAt')
  })

  it('renders employee credential summary and monitored surfaces', () => {
    render(<DarkWebMonitoringPage />)

    expect(screen.getByText('darkWebMonitoring.summary.employeeCredentials')).toBeInTheDocument()
    expect(screen.getByText('darkWebMonitoring.surfaces.title')).toBeInTheDocument()
    expect(screen.getByText('darkWebMonitoring.surfaces.monitoredCount')).toBeInTheDocument()
    expect(screen.getAllByText('test01.example').length).toBeGreaterThan(0)
  })

  it('opens the report menu with its export options', () => {
    render(<DarkWebMonitoringPage />)

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.report/ }))

    expect(screen.getByRole('menu')).toBeInTheDocument()
    expect(screen.getByText('darkWebMonitoring.reportMenu.title')).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.exportCsv' })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.customRange' })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.advancedOptions' })).toBeInTheDocument()
  })

  it('opens a breach details side panel when a table row is selected', () => {
    render(<DarkWebMonitoringPage />)
    const table = screen.getByRole('table')
    const row = within(table).getAllByText('taylor@test01.example')[0].closest('tr')
    if (!row) throw new Error('Expected a breach table row')

    fireEvent.click(row)

    const details = screen.getByRole('dialog', { name: 'darkWebMonitoring.details.title' })
    expect(details).toBeInTheDocument()
    expect(within(details).getByText('darkWebMonitoring.details.information')).toBeInTheDocument()
    expect(within(details).getByText('Synthetic demonstration breach')).toBeInTheDocument()
    expect(within(details).getByText('••••••••')).toBeInTheDocument()

    fireEvent.click(within(details).getByRole('button', { name: 'darkWebMonitoring.details.revealSecrets' }))
    expect(within(details).getByText('Demo!4821')).toBeInTheDocument()

    fireEvent.click(within(details).getByRole('button', { name: 'darkWebMonitoring.details.close' }))
    expect(screen.queryByRole('dialog', { name: 'darkWebMonitoring.details.title' })).not.toBeInTheDocument()
  })

  it('opens the custom range dialog and validates date selection', () => {
    render(<DarkWebMonitoringPage />)

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.report/ }))
    fireEvent.click(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.customRange' }))

    expect(screen.getByRole('dialog', { name: 'darkWebMonitoring.exportDialogs.custom.title' })).toBeInTheDocument()
    const selectButton = screen.getByRole('button', { name: /darkWebMonitoring.exportDialogs.custom.select/ })
    expect(selectButton).toBeDisabled()

    const fromDate = screen.getByLabelText('darkWebMonitoring.exportDialogs.custom.from').parentElement?.querySelector('input')
    const toDate = screen.getByLabelText('darkWebMonitoring.exportDialogs.custom.to').parentElement?.querySelector('input')
    if (!fromDate || !toDate) throw new Error('Expected custom range date inputs')
    fireEvent.change(fromDate, { target: { value: '2026-01-01' } })
    fireEvent.change(toDate, { target: { value: '2026-01-31' } })

    expect(selectButton).toBeEnabled()
  })

  it('opens advanced export options and lets the user add a filter', () => {
    render(<DarkWebMonitoringPage />)

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.report/ }))
    fireEvent.click(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.advancedOptions' }))

    expect(screen.getByRole('dialog', { name: 'darkWebMonitoring.exportDialogs.advanced.title' })).toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('darkWebMonitoring.exportDialogs.advanced.value'), {
      target: { value: 'DOMAIN' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'darkWebMonitoring.exportDialogs.advanced.addFilter' }))

    const addedFilter = within(screen.getByRole('list')).getByRole('listitem')
    expect(addedFilter).toHaveTextContent('darkWebMonitoring.exportDialogs.fields.organizationElementType')
    expect(addedFilter).toHaveTextContent('darkWebMonitoring.exportDialogs.operators.contains')
    expect(addedFilter).toHaveTextContent('DOMAIN')
  })

  it('keeps table columns aligned with their group headers', () => {
    render(<DarkWebMonitoringPage />)
    const table = screen.getByRole('table')
    const headers = within(table).getAllByRole('columnheader')

    expect(headers.slice(0, 4).map((header) => header.getAttribute('colspan'))).toEqual(['2', '7', '4', '2'])
    expect(headers.slice(4).map((header) => header.textContent)).toEqual([
      'darkWebMonitoring.table.status',
      'darkWebMonitoring.table.who',
      'darkWebMonitoring.table.domain',
      'darkWebMonitoring.table.email',
      'darkWebMonitoring.table.ipAddress',
      'darkWebMonitoring.table.username',
      'darkWebMonitoring.table.password',
      'darkWebMonitoring.table.name',
      'darkWebMonitoring.table.phone',
      'darkWebMonitoring.table.database',
      'darkWebMonitoring.table.leakName',
      'darkWebMonitoring.table.foundIn',
      'darkWebMonitoring.table.source',
      'darkWebMonitoring.table.breached',
      'darkWebMonitoring.table.ingested',
    ])
    expect(within(table).getAllByRole('row')[2].querySelectorAll('td')).toHaveLength(15)
  })

  it('loads more filtered records while scrolling the table', () => {
    vi.useFakeTimers()
    try {
      render(<DarkWebMonitoringPage />)
      const table = screen.getByRole('table')
      const scrollContainer = screen.getByLabelText('darkWebMonitoring.table.results')

      expect(within(table).getAllByRole('row')).toHaveLength(7)
      expect(screen.queryByRole('button', { name: 'darkWebMonitoring.nextPage' })).not.toBeInTheDocument()

      Object.defineProperties(scrollContainer, {
        scrollHeight: { configurable: true, value: 1000 },
        scrollTop: { configurable: true, value: 900 },
        clientHeight: { configurable: true, value: 100 },
      })
      fireEvent.scroll(scrollContainer)

      expect(screen.getByRole('status')).toHaveTextContent('darkWebMonitoring.loadingMore')
      act(() => {
        vi.advanceTimersByTime(300)
      })
      fireEvent.scroll(scrollContainer)
      act(() => {
        vi.advanceTimersByTime(300)
      })

      expect(within(table).getAllByRole('row')).toHaveLength(9)
      expect(screen.getByText('darkWebMonitoring.endOfResults', { exact: false })).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('masks secrets by default and reveals only after opting in', () => {
    render(<DarkWebMonitoringPage />)

    expect(screen.queryByText('Demo!4821')).not.toBeInTheDocument()
    expect(screen.getAllByText('••••••••').length).toBeGreaterThan(0)

    fireEvent.click(screen.getByLabelText('darkWebMonitoring.reveal'))

    expect(screen.getByText('Demo!4821')).toBeInTheDocument()
  })

  it('filters the table by status and surface tab', () => {
    render(<DarkWebMonitoringPage />)
    const table = screen.getByRole('table')

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.statusFilter.mitigation/ }))
    expect(within(table).getAllByText('taylor@test01.example')).toHaveLength(2)
    expect(within(table).queryByText('alex@test02.example')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.statusFilter.all/ }))
    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.tabs.emails/ }))
    expect(within(table).getByText('jordan@test01.example')).toBeInTheDocument()
    expect(within(table).queryByText('taylor@test01.example')).not.toBeInTheDocument()
  })

  it('allows changing a row status from the colored status select', () => {
    render(<DarkWebMonitoringPage />)

    const rowStatusButton = screen
      .getAllByRole('button', { name: /darkWebMonitoring.status.mitigation/ })
      .find((button) => button.getAttribute('aria-haspopup') === 'listbox')
    if (!rowStatusButton) throw new Error('Expected a row status selector')
    fireEvent.click(rowStatusButton)
    fireEvent.click(screen.getByRole('option', { name: 'darkWebMonitoring.status.open' }))

    expect(screen.getAllByRole('button', { name: /darkWebMonitoring.status.open/ }).length).toBeGreaterThan(0)
    fireEvent.click(screen.getByRole('button', { name: 'darkWebMonitoring.statusFilter.open' }))
    expect(within(screen.getByRole('table')).getByText('taylor@test01.example')).toBeInTheDocument()
  })

  it('searches the selected surface records', () => {
    render(<DarkWebMonitoringPage />)
    fireEvent.change(screen.getByPlaceholderText('darkWebMonitoring.search'), {
      target: { value: 'alex@test02.example' },
    })

    const table = within(screen.getByRole('table'))
    expect(table.getByText('alex@test02.example')).toBeInTheDocument()
    expect(table.queryByText('taylor@test01.example')).not.toBeInTheDocument()
  })
})
