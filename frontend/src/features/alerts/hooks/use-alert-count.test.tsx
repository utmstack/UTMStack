import { describe, expect, test, vi, beforeEach, afterEach } from 'vitest'
import { render } from '@testing-library/react'
import { ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { AlertCountProvider, useAlertCount, type AlertCountHandler } from './use-alert-count'
import { alertsHttpService } from '../services/alerts-http.service'

const Wrap = ({ children }: { children: ReactNode }) => (
  <MemoryRouter>
    <AlertCountProvider>{children}</AlertCountProvider>
  </MemoryRouter>
)

beforeEach(() => vi.useFakeTimers())
afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

function countOpenSpy(sequence: number[]) {
  return vi.spyOn(alertsHttpService, 'countOpen').mockImplementation(async () => sequence.shift() ?? sequence[sequence.length - 1])
}

function Register({ handler }: { handler: AlertCountHandler }) {
  const { registerHandler } = useAlertCount()
  registerHandler(handler, 'test-idx')
  return null
}

describe('AlertCountProvider', () => {
  test('polls every 5 seconds and reports the total count', async () => {
    const spy = countOpenSpy([10])
    render(<Wrap><div /></Wrap>)
    expect(spy).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(5000)
    expect(spy).toHaveBeenCalledTimes(2)
  })

  test('calls handlers with the delta (and navigate fn) when count goes up', async () => {
    countOpenSpy([10, 13, 13])
    const handler = vi.fn()
    render(<Wrap><Register handler={handler} /></Wrap>)
    await vi.advanceTimersByTimeAsync(5000)
    expect(handler).toHaveBeenCalledTimes(1)
    expect(handler.mock.calls[0][0]).toBe(3)
    expect(typeof handler.mock.calls[0][1]).toBe('function')
  })

  test('a drop in count resets baseline, handlers not called', async () => {
    countOpenSpy([10, 8, 11])
    const handler = vi.fn()
    render(<Wrap><Register handler={handler} /></Wrap>)
    await vi.advanceTimersByTimeAsync(10000)
    expect(handler).toHaveBeenCalledTimes(1)
    expect(handler).toHaveBeenCalledWith(3, expect.any(Function))
  })
})
