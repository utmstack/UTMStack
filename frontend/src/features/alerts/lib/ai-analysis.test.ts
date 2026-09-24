import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Alert } from '../types/alert.types'
import { aiAnalysisPayload, hasAiSummary, waitForAiSummary } from './ai-analysis'

const AI = '[AI SOC Agent] Score: 80/100 - Open - High Risk | Action: isolate'

const alert: Alert = {
  id: 'a-1',
  name: 'Firewall rule deleted',
  severity: 'high',
  status: 'Open',
  dataSource: 'WIN-01',
  notes: 'analyst wrote this',
  history: [{ user: 'x', action: 'UPDATE_STATUS' }] as never,
  events: [{ id: 'e-1' }] as never,
}

describe('aiAnalysisPayload', () => {
  it('sends the alert, without its events, history or notes', () => {
    const payload = aiAnalysisPayload(alert)

    expect(payload).toMatchObject({ id: 'a-1', name: 'Firewall rule deleted', severity: 'high', dataSource: 'WIN-01' })
    expect(payload).not.toHaveProperty('events')
    expect(payload).not.toHaveProperty('history')
    expect(payload).not.toHaveProperty('notes')
  })

  it('leaves out what the alert does not have', () => {
    expect(Object.keys(aiAnalysisPayload({ id: 'x' }))).toEqual(['id'])
  })
})

describe('hasAiSummary', () => {
  it('finds the assessment in the notes or in the status observation', () => {
    expect(hasAiSummary({ id: 'a', notes: `analyst\n\n${AI}` })).toBe(true)
    expect(hasAiSummary({ id: 'a', statusObservation: AI })).toBe(true)
    expect(hasAiSummary({ id: 'a', notes: 'analyst only' })).toBe(false)
    expect(hasAiSummary(null)).toBe(false)
  })
})

describe('waitForAiSummary', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('returns the alert as soon as it carries a summary', async () => {
    const fetchAlert = vi
      .fn<() => Promise<Alert | null>>()
      .mockResolvedValueOnce({ id: 'a' })
      .mockResolvedValueOnce({ id: 'a' })
      .mockResolvedValueOnce({ id: 'a', notes: AI })

    const pending = waitForAiSummary(fetchAlert, { intervalMs: 1000, timeoutMs: 60_000 })
    await vi.advanceTimersByTimeAsync(3000)

    expect((await pending)?.notes).toBe(AI)
    expect(fetchAlert).toHaveBeenCalledTimes(3)
  })

  it('keeps waiting through a failed read', async () => {
    const fetchAlert = vi
      .fn<() => Promise<Alert | null>>()
      .mockRejectedValueOnce(new Error('store busy'))
      .mockResolvedValueOnce({ id: 'a', notes: AI })

    const pending = waitForAiSummary(fetchAlert, { intervalMs: 1000, timeoutMs: 60_000 })
    await vi.advanceTimersByTimeAsync(2000)

    expect((await pending)?.id).toBe('a')
  })

  it('gives up when the time is up', async () => {
    const fetchAlert = vi.fn<() => Promise<Alert | null>>().mockResolvedValue({ id: 'a' })

    const pending = waitForAiSummary(fetchAlert, { intervalMs: 1000, timeoutMs: 5000 })
    await vi.advanceTimersByTimeAsync(6000)

    expect(await pending).toBeNull()
  })

  it('stops as soon as it is cancelled', async () => {
    let cancelled = false
    const fetchAlert = vi.fn<() => Promise<Alert | null>>().mockResolvedValue({ id: 'a' })

    const pending = waitForAiSummary(fetchAlert, { intervalMs: 1000, timeoutMs: 60_000, isCancelled: () => cancelled })
    await vi.advanceTimersByTimeAsync(1500)
    cancelled = true
    await vi.advanceTimersByTimeAsync(5000)

    expect(await pending).toBeNull()
    expect(fetchAlert.mock.calls.length).toBeLessThanOrEqual(2)
  })
})
