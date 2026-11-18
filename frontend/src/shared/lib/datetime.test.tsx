import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { formatDateTime, formatTime, parseTimestamp, setDateFormatPrefs, useDateFormat } from './datetime'
import { absTime, relativeTime } from '@/features/alerts/lib/alert-meta'
import { shortTime } from '@/features/log-explorer/components/log-results'

const value = '2026-10-15 23:12:00.123456789'
afterEach(() => setDateFormatPrefs({ timezone: 'UTC', dateFormat: 'medium' }))

describe('UTC timestamp display preferences', () => {
  it('normalizes fractions and preserves explicit offsets', () => {
    expect(parseTimestamp(value).toISOString()).toBe('2026-10-15T23:12:00.123Z')
    expect(parseTimestamp('2026-10-16T02:12:00.123456789+03:00').toISOString()).toBe('2026-10-15T23:12:00.123Z')
    expect(parseTimestamp('2026-10-15T23:12:00Z').toISOString()).toBe('2026-10-15T23:12:00.000Z')
    expect(parseTimestamp(0).getTime()).toBe(0)
    expect(parseTimestamp(new Date(0)).getTime()).toBe(0)
  })
  it('uses the selected timezone and date style for logs and alerts', () => {
    for (const timezone of ['UTC', 'Europe/Istanbul', 'America/New_York', 'system']) {
      for (const dateFormat of ['short', 'long'] as const) {
        setDateFormatPrefs({ timezone, dateFormat })
        const expected = new Intl.DateTimeFormat(undefined, {
          dateStyle: dateFormat, timeStyle: 'medium',
          timeZone: timezone === 'system' ? undefined : timezone,
        }).format(new Date('2026-10-15T23:12:00.123Z'))
        expect(formatDateTime(value)).toBe(expected)
        expect(shortTime(value)).toBe(expected)
        expect(absTime(value)).toBe(expected)
      }
    }
  })
  it('updates subscribed displays without reload', () => {
    function Display() {
      const { formatTime } = useDateFormat()
      return <span>{formatTime(value)}</span>
    }
    setDateFormatPrefs({ timezone: 'UTC' })
    render(<Display />)
    const before = formatTime(value)
    expect(screen.getByText(before)).toBeInTheDocument()
    act(() => setDateFormatPrefs({ timezone: 'Europe/Istanbul' }))
    expect(formatTime(value)).not.toBe(before)
    expect(screen.getByText(formatTime(value))).toBeInTheDocument()
  })
  it('handles invalid input and normalizes relative timestamps', () => {
    expect(formatDateTime('invalid')).toBe('—')
    expect(shortTime('invalid')).toBe('invalid')
    expect(absTime()).toBe('—')
    expect(relativeTime(value)).toBe(relativeTime('2026-10-15T23:12:00.123Z'))
  })
})
