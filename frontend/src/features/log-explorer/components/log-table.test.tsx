import { formatDateTime, setDateFormatPrefs } from '@/shared/lib/datetime'
import { act, fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { LogDocument } from '../types/log-explorer.types'
import { LogTable } from './log-table'

const docs: LogDocument[] = [
  { '@timestamp': '2026-09-24T08:09:07Z', dataSource: 'HOST-1', severity: 'high', action: 'login' },
  { '@timestamp': '2026-09-24T08:10:07Z', dataSource: 'HOST-2', severity: 'low', action: 'logout' },
]

const props = { docs, autoColumns: [] as string[], expanded: null as number | null, onToggle: vi.fn() }

describe('LogTable', () => {
  it('updates memoized timestamp cells when timezone changes', () => {
    setDateFormatPrefs({ timezone: 'UTC', dateFormat: 'short' })
    render(<LogTable {...props} columns={['action']} />)
    const before = formatDateTime(docs[0]['@timestamp'] as string)
    expect(screen.getByText(before)).toBeInTheDocument()
    act(() => setDateFormatPrefs({ timezone: 'Europe/Istanbul' }))
    expect(screen.queryByText(before)).not.toBeInTheDocument()
    expect(screen.getByText(formatDateTime(docs[0]['@timestamp'] as string))).toBeInTheDocument()
    act(() => setDateFormatPrefs({ timezone: 'UTC', dateFormat: 'medium' }))
  })

  it('shows one row per document with the source and the picked columns', () => {
    render(<LogTable {...props} columns={['action']} />)

    const table = screen.getByRole('table')
    expect(within(table).getAllByRole('row')).toHaveLength(3) // header + 2
    expect(within(table).getByText('login')).toBeInTheDocument()
    expect(within(table).getByText('logout')).toBeInTheDocument()
  })

  it('lets a picked column be @timestamp beside the Time column', () => {
    render(<LogTable {...props} columns={['@timestamp', 'action']} />)

    const headers = within(screen.getByRole('table')).getAllByRole('columnheader')
    // expand, level, time, @timestamp, action
    expect(headers).toHaveLength(5)
    expect(headers[3]).toHaveTextContent('@timestamp')
  })

  it('reports which row was clicked', () => {
    const onToggle = vi.fn()
    render(<LogTable {...props} columns={['action']} onToggle={onToggle} />)

    fireEvent.click(screen.getByText('logout'))

    expect(onToggle).toHaveBeenCalledWith(1)
  })

  it('opens the detail of the expanded row only', () => {
    render(<LogTable {...props} columns={['action']} expanded={0} />)

    const table = screen.getByRole('table')
    expect(within(table).getAllByText('logExplorer.detail.parsedFields')).toHaveLength(1)
  })

  it('offers to remove a picked column', () => {
    const onRemoveColumn = vi.fn()
    render(<LogTable {...props} columns={['action']} onRemoveColumn={onRemoveColumn} />)

    fireEvent.click(screen.getByTitle('logExplorer.results.removeColumn'))

    expect(onRemoveColumn).toHaveBeenCalledWith('action')
  })
})
