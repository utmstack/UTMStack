export const TS = '@timestamp'

export const SELECT_CLS =
  'h-8 cursor-pointer rounded-md border border-border bg-background px-2 text-xs transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

export const OP_KEY: Record<string, string> = {
  IS: 'is',
  IS_NOT: 'isNot',
  CONTAIN: 'contains',
  DOES_NOT_CONTAIN: 'doesNotContain',
  START_WITH: 'startsWith',
  DOES_NOT_START_WITH: 'doesNotStartWith',
  ENDS_WITH: 'endsWith',
  DOES_NOT_END_WITH: 'doesNotEndWith',
  EXIST: 'exists',
  DOES_NOT_EXIST: 'doesNotExist',
  IS_BETWEEN: 'between',
  IS_NOT_BETWEEN: 'notBetween',
  IS_GREATER_THAN: 'greaterThan',
  IS_GREATER_THAN_OR_EQUALS: 'greaterThanOrEquals',
  IS_LESS_THAN: 'lessThan',
  IS_LESS_THAN_OR_EQUALS: 'lessThanOrEquals',
  IS_IN_FIELDS: 'search',
  IS_ONE_OF_TERMS: 'isOneOf',
  IS_NOT_ONE_OF_TERMS: 'isNotOneOf',
}

export function chartTimeLabel(c: string) {
  const d = new Date(c)
  return Number.isNaN(d.getTime())
    ? c
    : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
