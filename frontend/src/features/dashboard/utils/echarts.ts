import type { Row } from '@/features/dashboard/service/opensearch.service'

export interface ParsedChartConfig {
  option: Record<string, unknown> | null
  error: string | null
}

export function parseChartConfig(raw: string): ParsedChartConfig {
  if (!raw || !raw.trim()) {
    return { option: null, error: 'empty config' }
  }
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return { option: null, error: 'config is not an object' }
    }
    return { option: parsed as Record<string, unknown>, error: null }
  } catch (err) {
    return { option: null, error: err instanceof Error ? err.message : 'invalid JSON' }
  }
}

export function mergeRowsIntoOption(
  option: Record<string, unknown>,
  rows: Row[]
): Record<string, unknown> {
  const existingDataset = isRecord(option.dataset) ? option.dataset : {}
  return {
    ...option,
    dataset: {
      ...existingDataset,
      source: rows,
    },
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
