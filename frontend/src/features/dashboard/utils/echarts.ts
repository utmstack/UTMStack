import type { TimeRange } from '@/shared/components/ui/time-range-picker'

export interface ParsedChartConfig {
  option: Record<string, unknown> | null
  error: string | null
}

// TODO: the time range is wired in but does not yet refetch data. The legacy
// backend executed Visualization.sqlQuery server-side with the picker's range
// applied. Here we only stash from/to on the option's dataset metadata so a
// future query runner can pick it up.
export function parseChartConfig(raw: string, time?: TimeRange): ParsedChartConfig {
  if (!raw || !raw.trim()) {
    return { option: null, error: 'empty config' }
  }
  let option: Record<string, unknown>
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      return { option: null, error: 'config is not an object' }
    }
    option = parsed as Record<string, unknown>
  } catch (err) {
    return { option: null, error: err instanceof Error ? err.message : 'invalid JSON' }
  }
  if (time) {
    option = withTimeRange(option, time)
  }
  return { option, error: null }
}

function withTimeRange(option: Record<string, unknown>, time: TimeRange): Record<string, unknown> {
  const dataset = isRecord(option.dataset) ? { ...option.dataset } : {}
  const source = isRecord(dataset.source) ? { ...dataset.source } : dataset.source
  if (isRecord(source) || source === undefined) {
    const meta = isRecord(source) ? source : ({} as Record<string, unknown>)
    return {
      ...option,
      dataset: {
        ...dataset,
        source: { ...meta, __timeRange: { from: time.from, to: time.to, interval: time.interval } },
      },
    }
  }
  return option
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
