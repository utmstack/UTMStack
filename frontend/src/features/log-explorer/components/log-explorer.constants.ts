import { formatDateTime, parseTimestamp } from '@/shared/lib/datetime'
export const TS = '@timestamp'

export const SELECT_CLS =
  'h-8 cursor-pointer rounded-md border border-border bg-background px-2 text-xs transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

export const OP_KEY: Record<string, string> = {
  IS: 'is',
  IS_NOT: 'isNot',
  CONTAIN: 'contains',
  EXIST: 'exists',
  IS_BETWEEN: 'between',
  IS_IN_FIELDS: 'search',
  IS_ONE_OF_TERMS: 'isOneOf',
}

export function chartTimeLabel(c: string) {
  return Number.isNaN(parseTimestamp(c).getTime()) ? c : formatDateTime(c)
}
