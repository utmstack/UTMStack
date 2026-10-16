import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { LogDocument } from '../types/log-explorer.types'
import { LogTable } from './log-table'

const docs: LogDocument[] = [
  { '@timestamp': '2026-09-24T08:09:07Z', dataSource: 'HOST-1', severity: 'high', action: 'login' },
  { '@timestamp': '2026-09-24T08:10:07Z', dataSource: 'HOST-2', severity: 'low', action: 'logout' },
]

const props = { docs, autoColumns: [] as string[], expanded: null as number | null, onToggle: vi.fn() }

describe('LogTable', () => {
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

describe('picked column ordering', () => {
  it('drags a picked column and renders headers and cells in the new order', () => {
    const onReorderColumn = vi.fn()
    const { rerender } = render(<LogTable {...props} columns={['action', 'dataSource']} onReorderColumn={onReorderColumn} />)
    const handles = screen.getAllByRole('button', { name: 'logExplorer.results.reorderColumn' })
    const values = new Map<string, string>()
    const dataTransfer = { types: ['application/x-utmstack-log-column'], setData: (type: string, value: string) => values.set(type, value), getData: (type: string) => values.get(type) ?? '' }
    fireEvent.dragStart(handles[0], { dataTransfer })
    fireEvent.dragOver(handles[1].parentElement!, { dataTransfer })
    fireEvent.drop(handles[1].parentElement!, { dataTransfer })
    expect(onReorderColumn).toHaveBeenCalledWith('action', 'dataSource')
    rerender(<LogTable {...props} columns={['dataSource', 'action']} onReorderColumn={onReorderColumn} />)
    const headers = screen.getAllByRole('columnheader')
    expect(headers[2]).toHaveTextContent('logExplorer.results.time')
    expect(headers[3]).toHaveTextContent('dataSource')
    expect(headers[4]).toHaveTextContent('action')
    const cells = within(screen.getAllByRole('row')[1]).getAllByRole('cell')
    expect(cells[3]).toHaveTextContent('HOST-1')
    expect(cells[4]).toHaveTextContent('login')
  })

  it('supports keyboard moves, guards edges, and keeps removal independent', () => {
    const onReorderColumn = vi.fn()
    const onRemoveColumn = vi.fn()
    render(<LogTable {...props} columns={['action', 'dataSource']} onReorderColumn={onReorderColumn} onRemoveColumn={onRemoveColumn} />)
    const handles = screen.getAllByRole('button', { name: 'logExplorer.results.reorderColumn' })
    fireEvent.keyDown(handles[0], { key: 'ArrowLeft', altKey: true })
    fireEvent.keyDown(handles[1], { key: 'ArrowRight', altKey: true })
    fireEvent.keyDown(handles[0], { key: 'ArrowRight' })
    expect(onReorderColumn).not.toHaveBeenCalled()
    fireEvent.keyDown(handles[0], { key: 'ArrowRight', altKey: true })
    expect(onReorderColumn).toHaveBeenCalledWith('action', 'dataSource')
    fireEvent.click(screen.getAllByTitle('logExplorer.results.removeColumn')[0])
    expect(onRemoveColumn).toHaveBeenCalledWith('action')
    expect(onReorderColumn).toHaveBeenCalledTimes(1)
  })

  it('does not offer ordering for automatic columns or accept external drags', () => {
    const onReorderColumn = vi.fn()
    const { rerender } = render(<LogTable {...props} columns={[]} autoColumns={['action']} onReorderColumn={onReorderColumn} />)
    expect(screen.queryByRole('button', { name: 'logExplorer.results.reorderColumn' })).not.toBeInTheDocument()
    rerender(<LogTable {...props} columns={['action']} onReorderColumn={onReorderColumn} />)
    fireEvent.drop(screen.getByRole('button').parentElement!, { dataTransfer: { getData: () => 'unknown' } })
    expect(onReorderColumn).not.toHaveBeenCalled()
  })
})
