import type { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Visualization } from '@/features/dashboard/types'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'
import { useVisualizationData } from './useVisualizationData'

// A plain counter, not vi.fn(): vitest reports the rejection its spies record
// for an async function that throws as an unhandled error.
const backend = vi.hoisted(() => ({
  calls: 0,
  answer: (async () => ({})) as () => Promise<unknown>,
}))

vi.mock('@/features/dashboard/service/query.service', () => ({
  createQueryService: () => ({
    run: () => {
      backend.calls++
      return backend.answer()
    },
  }),
}))

const visualization: Visualization = {
  id: 'viz-1',
  dashboardId: 'dash-1',
  spec: JSON.stringify({ dataset: 'alerts', chart: 'metric', metric: { agg: 'count' } }),
  config: '{}',
  layout: '{}',
}

// Absolute, so two mounts share a query key ("now" is resolved per mount).
const time: TimeRange = { from: '2026-01-01T00:00:00.000Z', to: '2026-01-02T00:00:00.000Z', interval: 'hour' }

const failWith500 = () => {
  backend.answer = async () => {
    throw new Error('Request failed with status code 500')
  }
}

// The app's own client is a plain `new QueryClient()`, so the defaults (3
// retries with backoff) are what a widget inherits unless the hook overrides them.
function setup() {
  const client = new QueryClient()
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  )
  return { client, wrapper }
}

describe('useVisualizationData', () => {
  beforeEach(() => {
    backend.calls = 0
  })

  it('asks the backend once when the query fails, instead of retrying', async () => {
    failWith500()
    const { wrapper } = setup()

    const { result } = renderHook(() => useVisualizationData(visualization, time), { wrapper })

    await waitFor(() => expect(result.current.isError).toBe(true))
    // A retry would have been scheduled 1s after the first failure.
    await new Promise((r) => setTimeout(r, 1500))
    expect(backend.calls).toBe(1)
  })

  it('does not ask again when a failed widget mounts again', async () => {
    failWith500()
    const { wrapper } = setup()

    const first = renderHook(() => useVisualizationData(visualization, time), { wrapper })
    await waitFor(() => expect(first.result.current.isError).toBe(true))
    first.unmount()

    const second = renderHook(() => useVisualizationData(visualization, time), { wrapper })
    await waitFor(() => expect(second.result.current.isError).toBe(true))

    expect(backend.calls).toBe(1)
  })

  it('returns the rows when the query works', async () => {
    backend.answer = async () => ({ total: 7 })
    const { wrapper } = setup()

    const { result } = renderHook(() => useVisualizationData(visualization, time), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ rows: [{ value: 7 }], total: 7 })
  })
})
