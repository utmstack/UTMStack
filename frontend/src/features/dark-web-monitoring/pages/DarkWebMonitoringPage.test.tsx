import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DEMO_RECORDS } from '../data/demo-records'
import { DEMO_ORGANIZATION_BREACH_STATS } from '../data/demo-organization-stats'
import { buildExportRows, downloadCsv, MASKED, printReportPdf } from '../lib/export'
import { formatDateMedium, neutralizeCsvCell, safeHtml, safeImageUrl, valueOrDash } from '../lib/format'
import { resolveTimeBounds } from '../components/TimeRangeFilter'
import { TimeRangeFilter } from '../components/TimeRangeFilter'
import { DarkWebMonitoringPage } from './DarkWebMonitoringPage'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}))

/**
 * Renders the page and flushes the mock service, which resolves on a
 * microtask, so the component has left its loading state.
 */
async function renderPage(): Promise<void> {
  render(<DarkWebMonitoringPage />)
  await act(async () => {
    await Promise.resolve()
  })
}

describe('DarkWebMonitoringPage demo', () => {
  it('keeps the demo statistics in the OrganizationBreachStats response shape', () => {
    expect(DEMO_ORGANIZATION_BREACH_STATS).toEqual({
      id: 10901,
      orgName: 'test01',
      totalEmailsBreaches: 2,
      totalPhonesBreaches: 1,
      totalDomainBreaches: 7,
      totalIpsBreaches: 1,
      totalUsernamesBreaches: 1,
      totalBreaches: 12,
    })
  })

  it('renders tab counts from organization breach statistics', async () => {
    await renderPage()

    expect(screen.getAllByRole('button', { name: /darkWebMonitoring\.tabs\.domains.*7/ }).length).toBeGreaterThan(0)
    expect(screen.getByRole('button', { name: /darkWebMonitoring\.tabs\.usernames.*1/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /darkWebMonitoring\.tabs\.ips.*1/ })).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
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
    expect(record).toHaveProperty('password', null)
    expect(record).toHaveProperty('databaseName')
    expect(record).not.toHaveProperty('status')
    expect(record).not.toHaveProperty('ingestedAt')
  })

  it('renders employee credential summary and monitored surfaces', async () => {
    await renderPage()

    expect(screen.getByText('darkWebMonitoring.summary.employeeCredentials')).toBeInTheDocument()
    expect(screen.getByText('darkWebMonitoring.surfaces.title')).toBeInTheDocument()
    expect(screen.getByText('darkWebMonitoring.surfaces.monitoredCount')).toBeInTheDocument()
    expect(screen.getAllByText('test01.example').length).toBeGreaterThan(0)
  })

  it('opens the report menu with its export options', async () => {
    await renderPage()

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.report/ }))

    expect(screen.getByRole('menu')).toBeInTheDocument()
    expect(screen.getByText('darkWebMonitoring.reportMenu.title')).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.exportCsv' })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.customRange' })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: 'darkWebMonitoring.reportMenu.advancedOptions' })).toBeInTheDocument()
  })

  it('opens a breach details side panel when a table row is selected', async () => {
    await renderPage()
    const table = screen.getByRole('table')
    const row = within(table).getAllByText('user01@test01.example')[0].closest('tr')
    if (!row) throw new Error('Expected a breach table row')

    fireEvent.click(row)

    const details = screen.getByRole('dialog', { name: 'darkWebMonitoring.details.title' })
    expect(details).toBeInTheDocument()
    expect(within(details).getByText('darkWebMonitoring.details.information')).toBeInTheDocument()
    expect(within(details).getByText('Synthetic demonstration breach')).toBeInTheDocument()
    expect(within(details).getByText('user01@test01.example')).toBeInTheDocument()

    fireEvent.click(within(details).getByRole('button', { name: 'darkWebMonitoring.details.revealSecrets' }))
    expect(within(details).queryByText('Demo!4821')).not.toBeInTheDocument()

    fireEvent.click(within(details).getByRole('button', { name: 'darkWebMonitoring.details.close' }))
    expect(screen.queryByRole('dialog', { name: 'darkWebMonitoring.details.title' })).not.toBeInTheDocument()
  })

  it('opens the custom range dialog and validates date selection', async () => {
    await renderPage()

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

  it('opens advanced export options and lets the user add a filter', async () => {
    await renderPage()

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

  it('keeps table columns aligned with their group headers', async () => {
    await renderPage()
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

  it('loads more filtered records while scrolling the table', async () => {
    vi.useFakeTimers()
    try {
      render(<DarkWebMonitoringPage />)
      await act(async () => {
        await Promise.resolve()
      })
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

  it('auto-loads the remaining rows when the visible ones fit without scrolling', async () => {
    vi.useFakeTimers()
    try {
      render(<DarkWebMonitoringPage />)
      await act(async () => {
        await Promise.resolve()
      })
      const table = screen.getByRole('table')
      const scrollContainer = screen.getByLabelText('darkWebMonitoring.table.results')

      // 7 domain records, 5 visible. jsdom reports no box geometry, so make the
      // container short enough that its rows fit without a scrollbar.
      expect(within(table).getAllByRole('row')).toHaveLength(7)
      Object.defineProperties(scrollContainer, {
        scrollHeight: { configurable: true, value: 100 },
        clientHeight: { configurable: true, value: 500 },
      })

      // A filter change re-runs the "does it fit?" check and loads the rest.
      const emailsTab = screen.getByRole('button', { name: 'darkWebMonitoring.tabs.emails2' })
      fireEvent.click(emailsTab)
      await act(async () => {
        await Promise.resolve()
      })
      const domainsTab = screen.getByRole('button', { name: 'darkWebMonitoring.tabs.domains7' })
      fireEvent.click(domainsTab)
      await act(async () => {
        await Promise.resolve()
      })
      act(() => {
        vi.advanceTimersByTime(300)
      })

      expect(within(table).getAllByRole('row')).toHaveLength(9)
      expect(screen.getByText('darkWebMonitoring.endOfResults', { exact: false })).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('masks secrets by default and reveals only after opting in', async () => {
    await renderPage()

    expect(screen.queryByText('Demo!4821')).not.toBeInTheDocument()
    expect(screen.queryByText('••••••••')).not.toBeInTheDocument()

    fireEvent.click(screen.getByLabelText('darkWebMonitoring.reveal'))

    expect(screen.queryByText('Demo!4821')).not.toBeInTheDocument()
  })

  it('filters the table by status and surface tab', async () => {
    await renderPage()
    const table = screen.getByRole('table')

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.statusFilter.mitigation/ }))
    expect(within(table).getAllByText('user01@test01.example')).toHaveLength(1)
    expect(within(table).queryByText('user03@test02.example')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.statusFilter.all/ }))
    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring.tabs.emails/ }))
    expect(within(table).getByText('user06@test01.example')).toBeInTheDocument()
    expect(within(table).queryByText('user01@test01.example')).not.toBeInTheDocument()
  })

  it('allows changing a row status from the colored status select', async () => {
    await renderPage()

    const rowStatusButton = screen
      .getAllByRole('button', { name: /darkWebMonitoring.status.mitigation/ })
      .find((button) => button.getAttribute('aria-haspopup') === 'listbox')
    if (!rowStatusButton) throw new Error('Expected a row status selector')
    fireEvent.click(rowStatusButton)
    fireEvent.click(screen.getByRole('option', { name: 'darkWebMonitoring.status.open' }))

    expect(screen.getAllByRole('button', { name: /darkWebMonitoring.status.open/ }).length).toBeGreaterThan(0)
    fireEvent.click(screen.getByRole('button', { name: 'darkWebMonitoring.statusFilter.open' }))
    expect(within(screen.getByRole('table')).getByText('user01@test01.example')).toBeInTheDocument()
  })

  it('searches the selected surface records', async () => {
    await renderPage()
    fireEvent.change(screen.getByPlaceholderText('darkWebMonitoring.search'), {
      target: { value: 'user03@test02.example' },
    })

    const table = within(screen.getByRole('table'))
    expect(table.getByText('user03@test02.example')).toBeInTheDocument()
    expect(table.queryByText('user01@test01.example')).not.toBeInTheDocument()
  })
})

describe('dark web monitoring format helpers', () => {
  it('renders empty values as a dash and formats dates', () => {
    expect(valueOrDash(null)).toBe('—')
    expect(valueOrDash('')).toBe('—')
    expect(valueOrDash('value')).toBe('value')
    expect(formatDateMedium('2026-08-28T12:20:00Z')).toMatch(/2026/)
  })

  it('only allows http(s) and data:image image urls through safeImageUrl', () => {
    expect(safeImageUrl('https://example.com/a.png')).toBe('https://example.com/a.png')
    expect(safeImageUrl('http://example.com/a.png')).toBe('http://example.com/a.png')
    expect(safeImageUrl('data:image/svg+xml,%3Csvg%3E')).toBe('data:image/svg+xml,%3Csvg%3E')
    expect(safeImageUrl('data:text/html,<script>')).toBeNull()
    expect(safeImageUrl('javascript:alert(1)')).toBeNull()
    expect(safeImageUrl(null)).toBeNull()
  })

  it('neutralizes formula-triggering CSV cells', () => {
    expect(neutralizeCsvCell('=SUM(A1)')).toBe("'=SUM(A1)")
    expect(neutralizeCsvCell('+cmd')).toBe("'+cmd")
    expect(neutralizeCsvCell('@import')).toBe("'@import")
    expect(neutralizeCsvCell('-1')).toBe("'-1")
    expect(neutralizeCsvCell('  =SUM(A1)')).toBe("'  =SUM(A1)")
    expect(neutralizeCsvCell('safe text')).toBe('safe text')
  })

  it('sanitizes i18n markup to an inline-tag allowlist', () => {
    expect(safeHtml('a <strong>b</strong> c')).toBe('a <strong>b</strong> c')
    expect(safeHtml('<img src=x onerror=alert(1)>')).toBe('&lt;img src=x onerror=alert(1)&gt;')
    expect(safeHtml('<script>evil()</script>')).toBe('&lt;script&gt;evil()&lt;/script&gt;')
  })

  it('caps time presets at the current moment and collapses unparsed custom bounds', () => {
    const before = Date.now()
    const all = resolveTimeBounds({ kind: 'preset', preset: 'all' })
    const day = resolveTimeBounds({ kind: 'preset', preset: '24h' })
    const cleared = resolveTimeBounds({ kind: 'custom', from: '2026-01-01T00:00:00', to: '' })
    const after = Date.now()
    // Presets never match records ingested in the future: `to` is "now" at call time.
    for (const bound of [all.to, day.to, cleared.to]) {
      expect(bound).toBeGreaterThanOrEqual(before)
      expect(bound).toBeLessThanOrEqual(after)
    }
    expect(all.from).toBe(0)
    expect(day.from).toBe(day.to - 1440 * 60_000)
  })

  it('builds export rows with masked or revealed secrets', () => {
    const withSecret = { ...DEMO_RECORDS[0], password: 'Demo!4821' }
    const maskedRows = buildExportRows([withSecret], (key) => key, true, formatDateMedium)
    expect(maskedRows[0][0]).toBe('darkWebMonitoring.table.status')
    expect(maskedRows[1][1]).toBe('darkWebMonitoring.employee')
    expect(maskedRows[1].some((cell) => cell === MASKED)).toBe(true)

    const revealedRows = buildExportRows([withSecret], (key) => key, false, formatDateMedium)
    expect(revealedRows[1].some((cell) => cell === 'Demo!4821')).toBe(true)
  })
})

describe('dark web monitoring export', () => {
  it('writes a CSV file with the given rows', () => {
    const click = vi.fn()
    const createObjectURL = vi.fn(() => 'blob:mock')
    const revokeObjectURL = vi.fn()
    Object.defineProperty(URL, 'createObjectURL', { value: createObjectURL, configurable: true })
    Object.defineProperty(URL, 'revokeObjectURL', { value: revokeObjectURL, configurable: true })
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
      click(this.download)
    })

    downloadCsv('demo.csv', [['a', 'b'], ['1', '2']])

    expect(click).toHaveBeenCalledWith('demo.csv')
    expect(revokeObjectURL).toHaveBeenCalled()
  })

  it('opens a print window for the PDF report', () => {
    const openWindow = { document: { write: vi.fn(), close: vi.fn() } }
    const openSpy = vi.spyOn(window, 'open').mockReturnValue(openWindow as unknown as Window)

    expect(printReportPdf('Title', [['h'], ['v']])).toBe(true)
    expect(openWindow.document.write).toHaveBeenCalled()

    openSpy.mockRestore()
  })
})

describe('TimeRangeFilter hour spinner', () => {
  it('wraps hours at 24, not 60', () => {
    render(<TimeRangeFilter value={{ kind: 'preset', preset: 'all' }} onChange={() => {}} />)
    fireEvent.click(screen.getByRole('button', { name: /darkWebMonitoring\.time\.all/ }))

    // Open the custom view and find the hour spinner (first one, max 24).
    fireEvent.click(screen.getByRole('button', { name: 'darkWebMonitoring.time.custom' }))
    const upButtons = screen.getAllByRole('button', { name: '↑' })
    const hourUp = upButtons[0]
    expect(hourUp).toBeInTheDocument()

    // The displayed hour starts at 00; stepping up 25 times wraps to 01, not 25.
    for (let i = 0; i < 25; i += 1) {
      fireEvent.click(hourUp)
    }
    expect(screen.getByText('01')).toBeInTheDocument()
  })
})
