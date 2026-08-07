import { describe, expect, test, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useAlertsList } from '@/features/alerts/hooks/use-alerts-list'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import type { FilterType } from '@/features/alerts/types/alert.types'

const ALL = Array.from({ length: 25 }, (_, i) => ({ id: `all-${i}` }))
const ONE = [{ id: 'the-one' }]

beforeEach(() => vi.restoreAllMocks())

const listSpy = () =>
  vi.spyOn(alertsHttpService, 'list').mockImplementation(async ({ page, filters }) => {
    const filtered = filters.some((f) => f.field === 'id')
    if (filtered) return { data: page === 0 ? (ONE as never[]) : [], total: 1 }
    return { data: (page === 0 ? ALL : ALL) as never[], total: 2128 }
  })

describe('useAlertsList', () => {
  test('the first page asks the endpoint for page 0', async () => {
    const spy = listSpy()
    renderHook(() => useAlertsList(0, 25, []))
    await waitFor(() => expect(spy).toHaveBeenCalled())
    expect(spy.mock.calls[0][0].page).toBe(0)
  })

  // The sentinel can bump the page in the gap between a filter changing and its
  // first page landing. That page belongs to the old query.
  test('a stale page bump does not append one query onto another', async () => {
    listSpy()
    const noFilter: FilterType[] = []
    const idFilter: FilterType[] = [{ field: 'id', operator: 'IS', value: 'the-one' }]

    const { result, rerender } = renderHook(({ p, f }) => useAlertsList(p, 25, f), {
      initialProps: { p: 0, f: noFilter },
    })
    await waitFor(() => expect(result.current.alerts).toHaveLength(25))

    // filter changes while the page is still the scrolled-to one
    await act(async () => {
      rerender({ p: 1, f: idFilter })
    })
    await waitFor(() => expect(result.current.total).toBe(1))

    expect(result.current.alerts).toHaveLength(1)
    expect(result.current.alerts[0]).toMatchObject({ id: 'the-one' })
  })

  test('hasMore stays false until the current query has landed', async () => {
    listSpy()
    const { result } = renderHook(() => useAlertsList(0, 25, []))
    await waitFor(() => expect(result.current.alerts).toHaveLength(25))
    expect(result.current.hasMore).toBe(true)
  })
})
