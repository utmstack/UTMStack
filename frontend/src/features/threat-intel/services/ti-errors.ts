import { ApiError } from '@/shared/lib/api-client'

export function isNotConfigured(e: unknown): boolean {
  return e instanceof ApiError && e.status === 503 && /not configured/i.test(e.message ?? '')
}

export function isNotFound(e: unknown): boolean {
  return e instanceof ApiError && e.status === 404
}

// Backend go-sdk validation errors leak the raw regexp, e.g.
//   value 'evil-domain[.]example' does not match with regexp '^(?i)[a-z0-9]+...$': ...
// Keep the offending value, drop the pattern and duplicated tail.
const VALIDATION_RE = /value '([^']*)' does not match with regexp/

function cleanValidationMessage(raw: string): string {
  const m = raw.match(VALIDATION_RE)
  if (m) return `Invalid value '${m[1]}'`
  return raw
}

export function describeError(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.status === 502) return 'Threat intel upstream unavailable — try again.'
    if (e.status === 400) return cleanValidationMessage(e.message) || `Request failed (${e.status}).`
    return e.message || `Request failed (${e.status}).`
  }
  if (e instanceof Error) return e.message
  return 'Unknown error'
}
