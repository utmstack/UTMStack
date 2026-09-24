import { useEffect, useMemo, useState } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { filterAggregatableFields, groupableFields } from '@/features/dashboard/utils/aggregatable-fields'
import { isDateType } from '@/features/dashboard/utils/field-types'
import { logExplorerHttpService as svc } from '../services/log-explorer-http.service'
import type { ChartView, FilterType, IndexField } from '../types/log-explorer.types'
import { SELECT_CLS, TS } from './log-explorer.constants'
import { TermsChart } from './TermsChart'
import { TimeChart } from './TimeChart'

// The bucket sizes the histogram endpoint accepts; anything else is refused.
const CALENDAR_INTERVALS = [
  { id: 'minute', label: 'Minute' },
  { id: 'hour', label: 'Hour' },
  { id: 'day', label: 'Day' },
  { id: 'week', label: 'Week' },
]

export function ChartPanel({
  dataset,

  pattern,
  fields,
  filters,
}: {
  dataset: string
  pattern: string | null
  fields: IndexField[]
  filters: FilterType[]
}) {
  const { t } = useTranslation()
  // What a chart can break down by: a value per record — not a sub-document
  // (`origin`, `log`), not a list, not the raw text.
  const selectable = useMemo(
    () => groupableFields(filterAggregatableFields(fields)).sort((a, b) => a.name.localeCompare(b.name)),
    [fields]
  )
  const [fieldName, setFieldName] = useState('')
  const [interval, setInterval] = useState('day')
  const [data, setData] = useState<ChartView | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  // Default to @timestamp (time histogram) when present, else the first field —
  // and again when the dataset changes to one that lacks the chosen field.
  useEffect(() => {
    if (selectable.length === 0 || selectable.some((f) => f.name === fieldName)) return
    setFieldName(selectable.find((f) => f.name === TS)?.name ?? selectable[0].name)
  }, [selectable, fieldName])

  const field = selectable.find((f) => f.name === fieldName) ?? null
  const isDate = isDateType(field?.type)
  const aggField = field?.name ?? ''

  // A data type narrows the chart; without one it charts the whole dataset.
  useEffect(() => {
    if (!field) return
    // Answers can arrive out of order when the field or filters change quickly;
    // only the latest question's answer may reach the screen.
    let stale = false
    setLoading(true)
    setError(false)
    svc
      .chartView({
        dataset,
        dataType: pattern,
        field: aggField,
        fieldDataType: field.type,
        filters,
        interval: isDate ? interval : '',
        top: 20,
      })
      .then((view) => !stale && setData(view))
      .catch(() => {
        if (stale) return
        setData(null)
        setError(true)
      })
      .finally(() => !stale && setLoading(false))
    return () => {
      stale = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dataset, pattern, fieldName, interval, filters])

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b border-border/60 px-4 py-2.5 text-xs">
        <span className="text-muted-foreground">{t('logExplorer.chart.aggregateOn')}</span>
        <select value={fieldName} onChange={(e) => setFieldName(e.target.value)} className={cn(SELECT_CLS, 'min-w-[200px] font-mono')}>
          {selectable.map((f) => (
            <option key={f.name} value={f.name}>
              {f.name}
            </option>
          ))}
        </select>
        {isDate ? (
          <>
            <span className="text-muted-foreground">{t('logExplorer.chart.per')}</span>
            <select value={interval} onChange={(e) => setInterval(e.target.value)} className={SELECT_CLS}>
              {CALENDAR_INTERVALS.map((i) => (
                <option key={i.id} value={i.id}>
                  {t(`logExplorer.intervals.${i.id}`)}
                </option>
              ))}
            </select>
          </>
        ) : (
          <span className="text-muted-foreground">{t('logExplorer.chart.topValues')}</span>
        )}
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-5">
        {loading ? (
          <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" /> {t('logExplorer.chart.building')}
          </div>
        ) : error ? (
          <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" /> {t('logExplorer.chart.failed')}
          </div>
        ) : !data || data.values.length === 0 ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            {t('logExplorer.chart.noData')}
          </div>
        ) : isDate ? (
          <TimeChart data={data} />
        ) : (
          <TermsChart data={data} />
        )}
      </div>
    </div>
  )
}
