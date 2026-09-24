import { describe, expect, test } from 'vitest'
import {
  isValidMailRecipient,
  mailRecipientError,
  splitMailRecipients,
} from './mail-node-validity'

describe('mail node recipient validation', () => {
  test('accepts standard email addresses and SOAR templates', () => {
    expect(isValidMailRecipient('user@example.com')).toBe(true)
    expect(isValidMailRecipient('user.name+alerts@example.co.uk')).toBe(true)
    expect(isValidMailRecipient('$(alert.email)')).toBe(true)
    expect(isValidMailRecipient('$[variables.recipient]')).toBe(true)
  })

  test('rejects malformed email addresses', () => {
    expect(isValidMailRecipient('user@example')).toBe(false)
    expect(isValidMailRecipient('user@@example.com')).toBe(false)
    expect(isValidMailRecipient('not-an-email')).toBe(false)
  })

  test('splits and trims comma-separated recipients', () => {
    expect(splitMailRecipients(' alice@example.com, bob@example.com, ')).toEqual([
      'alice@example.com',
      'bob@example.com',
    ])
  })

  test('reports the first invalid address across to and cc', () => {
    expect(
      mailRecipientError({
        to: 'alice@example.com, invalid-address',
        cc: 'also-invalid',
      }),
    ).toEqual({ field: 'to', address: 'invalid-address' })
  })

  test('allows an empty optional cc field', () => {
    expect(mailRecipientError({ to: 'alice@example.com', cc: '  ' })).toBeUndefined()
  })
})