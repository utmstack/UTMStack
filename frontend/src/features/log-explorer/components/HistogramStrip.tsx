import { memo, useEffect, useState } from 'react'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'
import { logExplorerHttpService as svc } from '../services/log-explorer-http.service'
import type { ChartView, FilterType } from '../types/log-explorer.types'
import { TS, chartTimeLabel } from './log-explorer.constants'

// Pick a date-histogram interval that yields a readable number of buckets for
// the selected time window.
function histogramInterval(from?: string | null, to?: string | null): string {
  if (!from || !to) return 'hour'
  const span = new Date(to).getTime() - new Date(from).getTime()
  if (Number.isNaN(span) || span <= 0) return 'hour'
  const H = 3_600_000
  const D = 24 * H
  if (span <= 2 * H) return 'minute'
  if (span <= 2 * D) return 'hour'
  if (span <= 60 * D) return 'day'
  if (span <= 365 * D) return 'week'
  return 'month'
}

// Always-on event-volume histogram shown above the results table (Discover-style).
// Aggregates the same filtered/time-scoped result set on @timestamp so analysts
// see spikes and gaps without leaving the table. Uses flex columns (not a stretched
// SVG) so the strip never collapses into a solid block.
// Memoized: props are reference-stable (activeFilterList useMemo, pattern/range
// state) so SQL-mode keystrokes don't re-enter this subtree.
function HistogramStripImpl({
  pattern,
  filters,
  range,
}: {
  pattern: string
  filters: FilterType[]
  range: TimeRange
}) {
  const interval = histogramInterval(range.from, range.to)
  const [data, setData] = useState<ChartView | null>(null)

  useEffect(() => {
    let cancelled = false
    svc
      .chartView({
        indexPattern: pattern,
        field: TS,
        fieldDataType: 'date',
        filters,
        interval,
        top: 50,
      })
      .then((d) => {
        if (!cancelled) setData(d)
      })
      .catch(() => {
        if (!cancelled) setData(null)
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pattern, filters, interval])

  if (!data || data.values.length === 0) return null
  const max = Math.max(1, ...data.values)
  return (
    <div className="shrink-0 border-b border-border/60 px-4 pb-1.5 pt-2.5">
      <div className="flex h-12 items-end justify-center gap-px">
        {data.values.map((v, i) => (
          <div key={i} className="flex h-full max-w-[14px] flex-1 items-end">
            <div
              className="w-full rounded-t-[1px] bg-primary/40 transition-colors hover:bg-primary/70"
              style={{ height: `${v <= 0 ? 0 : Math.max(2, (v / max) * 100)}%` }}
              title={`${chartTimeLabel(data.categories[i])} · ${v.toLocaleString()}`}
            />
          </div>
        ))}
      </div>
      {data.categories.length > 1 && (
        <div className="mt-1 flex justify-between font-mono text-[9px] text-muted-foreground/70">
          <span>{chartTimeLabel(data.categories[0])}</span>
          <span>{chartTimeLabel(data.categories[data.categories.length - 1])}</span>
        </div>
      )}
    </div>
  )
}

export const HistogramStrip = memo(HistogramStripImpl)
