const FROM_PLACEHOLDER = /\{\{\s*from\s*\}\}/g
const TO_PLACEHOLDER = /\{\{\s*to\s*\}\}/g
const TIME_FILTER_PLACEHOLDER = /\{\{\s*timeFilter\s*\}\}/g

const DEFAULT_TIMESTAMP_FIELD = '@timestamp'

export interface SqlTemplateInput {
  sql: string
  fromISO: string
  toISO: string
  timestampField?: string
}

export function applySqlTemplate({
  sql,
  fromISO,
  toISO,
  timestampField = DEFAULT_TIMESTAMP_FIELD,
}: SqlTemplateInput): string {
  const tf = `${timestampField} >= '${fromISO}' AND ${timestampField} <= '${toISO}'`
  return sql
    .replace(TIME_FILTER_PLACEHOLDER, tf)
    .replace(FROM_PLACEHOLDER, fromISO)
    .replace(TO_PLACEHOLDER, toISO)
}

export function hasTimePlaceholder(sql: string): boolean {
  return (
    FROM_PLACEHOLDER.test(sql) ||
    TO_PLACEHOLDER.test(sql) ||
    TIME_FILTER_PLACEHOLDER.test(sql)
  )
}
