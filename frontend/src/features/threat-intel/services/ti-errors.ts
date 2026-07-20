import { ApiError } from '@/shared/lib/api-client'

export function isNotConfigured(e: unknown): boolean {
  return e instanceof ApiError && e.status === 503 && /not configured/i.test(e.message ?? '')
}

export function isNotFound(e: unknown): boolean {
  return e instanceof ApiError && e.status === 404
}

export function describeError(e: unknown): string {
  if (e instanceof ApiError) {
    if (e.status === 502) return 'Threat intel upstream unavailable — try again.'
    return e.message || `Request failed (${e.status}).`
  }
  if (e instanceof Error) return e.message
  return 'Unknown error'
}
