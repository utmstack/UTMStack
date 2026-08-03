import { useMemo, useState } from 'react'
import { AlertTriangle, Loader2, RefreshCw, ShieldAlert } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { TimeRangePicker, presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { AlertsFilterBar } from '@/features/alerts/components/alerts-filter-bar'
import { FILTER_OPS, TS } from '@/features/alerts/lib/alert-meta'
import type { CustomFilter, FilterType } from '@/features/alerts/types/alert.types'
import { useAdversaries } from '../hooks/use-adversaries'
import { AdversariesSankey } from '../components/adversaries-sankey'

export function AdversariesPage() {
  const { t } = useTranslation()
  const [customFilters, setCustomFilters] = useState<CustomFilter[]>([])
  const [range, setRange] = useState<TimeRange>(() => presetRange('7d'))

  const filters = useMemo<FilterType[]>(() => {
    const f: FilterType[] = []
    if (range.from) f.push({ field: TS, operator: 'IS_BETWEEN', value: [range.from, range.to] })
    for (const cf of customFilters) {
      const needsValue = FILTER_OPS.find((o) => o.id === cf.operator)?.needsValue ?? true
      f.push({ field: cf.field, operator: cf.operator, value: needsValue ? cf.value : undefined })
    }
    return f
  }, [range.from, range.to, customFilters])

  const { data, loading, error, refresh } = useAdversaries(filters)

  return (
    <div className="flex h-[calc(100vh-56px)] w-full  flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <ShieldAlert size={14} strokeWidth={1.75} />
          <span>
            <span className="font-medium text-foreground">{data.length}</span>{' '}
            {t('adversaries.title').toLowerCase()}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <TimeRangePicker value={range} onChange={setRange} allowAllTime align="right" />
          <button
            onClick={refresh}
            title={t('adversaries.refresh')}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
          </button>
        </div>
      </header>

      <div className="mt-3">
        <AlertsFilterBar
          filters={customFilters}
          onAdd={(f) => setCustomFilters((c) => [...c, f])}
          onUpdate={(i, f) => setCustomFilters((c) => c.map((x, idx) => (idx === i ? f : x)))}
          onRemove={(i) => setCustomFilters((c) => c.filter((_, idx) => idx !== i))}
          onClear={() => setCustomFilters([])}
        />
      </div>

      <div className="mt-3 min-h-0 flex-1">
        {error ? (
          <div className="flex h-full flex-col items-center justify-center gap-3 rounded-xl border border-border bg-card text-sm">
            <span className="inline-flex items-center gap-2 text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t('adversaries.loadError')}
            </span>
            <Button variant="outline" size="sm" onClick={refresh}>
              {t('adversaries.retry')}
            </Button>
          </div>
        ) : loading ? (
          <div className="flex h-full items-center justify-center gap-2 rounded-xl border border-border bg-card text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        ) : (
          <AdversariesSankey data={data} />
        )}
      </div>
    </div>
  )
}
