import { fireEvent, render, screen, within } from '@testing-library/react'
import type { TFunction } from 'i18next'
import { describe, expect, it, vi } from 'vitest'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import type { ApiKey } from '../types/api-key.types'
import { buildKeyColumns } from './key-columns'

const t = ((key: string) => key) as unknown as TFunction

const keys: ApiKey[] = [
  { id: 1, name: 'ci-deploy', allowed_ip: ['10.0.0.1', '10.0.0.2'], created_at: '2026-01-05T00:00:00Z', generated_at: '2026-02-05T00:00:00Z', expires_at: null },
  { id: 2, name: 'grafana', allowed_ip: [], created_at: '2026-01-06T00:00:00Z', generated_at: '2026-01-06T00:00:00Z', expires_at: '2027-01-06T00:00:00Z' },
]

function renderTable(actions = { onEdit: vi.fn(), onRotate: vi.fn(), onDelete: vi.fn() }) {
  render(
    <ResizableDataTable
      columns={buildKeyColumns(t, actions)}
      data={keys}
      flexColumnId="name"
      getRowId={(k) => String(k.id)}
    />,
  )
  return { actions, table: screen.getByRole('table') }
}

describe('API key columns', () => {
  it('shows a row per key with its addresses, or any address when there are none', () => {
    const { table } = renderTable()

    expect(within(table).getAllByRole('row')).toHaveLength(3) // header + 2
    expect(within(table).getByText('ci-deploy')).toBeInTheDocument()
    expect(within(table).getByText('10.0.0.1, 10.0.0.2')).toBeInTheDocument()
    expect(within(table).getByText('apiKeys.anyIp')).toBeInTheDocument()
  })

  it('says a key that never expires never expires', () => {
    const { table } = renderTable()

    expect(within(table).getByText('apiKeys.never')).toBeInTheDocument()
  })

  it('sends each action the key of its own row', () => {
    const { actions, table } = renderTable()
    const grafana = within(table).getAllByRole('row')[2]

    fireEvent.click(within(grafana).getByLabelText('apiKeys.action.rotate'))
    fireEvent.click(within(grafana).getByLabelText('apiKeys.action.edit'))
    fireEvent.click(within(grafana).getByLabelText('apiKeys.action.delete'))

    expect(actions.onRotate).toHaveBeenCalledWith(keys[1])
    expect(actions.onEdit).toHaveBeenCalledWith(keys[1])
    expect(actions.onDelete).toHaveBeenCalledWith(keys[1])
  })

  it('keeps the actions column at a fixed width', () => {
    const { table } = renderTable()

    // The last header is the actions column; nothing can be dragged onto it.
    const headers = within(table).getAllByRole('columnheader')
    expect(headers[headers.length - 1]).toHaveTextContent('apiKeys.col.actions')
    expect(headers[headers.length - 1].querySelector('[class*="cursor-col-resize"]')).toBeNull()
  })
})
