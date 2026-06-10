import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { Activity, AlertTriangle, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { ingestionHttpService } from '../services/data-processing-http.service'
import type { GroupBy, IngestionStats, IngestionStatus, IngestionTimeline } from '../types/data-processing.types'

type DropReason = 'parsing_dropped' | 'analysis_dropped' | 'correlation_dropped'
const DROP_REASONS: DropReason[] = ['parsing_dropped', 'analysis_dropped', 'correlation_dropped']
const EXPLORER_STATUSES: IngestionStatus[] = ['received', 'parsing_dropped', 'analysis_dropped', 'correlation_dropped', 'all']

function compact(n: number): string {
  if (n >= 1e9) return (n / 1e9).toFixed(1).replace(/\.0$/, '') + 'B'
  if (n >= 1e6) return (n / 1e6).toFixed(1).replace(/\.0$/, '') + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1).replace(/\.0$/, '') + 'k'
  return String(n)
}

const RANGES = [
  { id: '1h', hours: 1 },
  { id: '24h', hours: 24 },
  { id: '7d', hours: 24 * 7 },
  { id: '30d', hours: 24 * 30 },
] as const
type RangeId = (typeof RANGES)[number]['id']

/** Compute the from/to window (RFC3339) for the selected range, at call time. */
function rangeWindow(id: RangeId): { from: string; to: string } {
  const hours = RANGES.find((r) => r.id === id)?.hours ?? 24
  const now = Date.now()
  return { from: new Date(now - hours * 3_600_000).toISOString(), to: new Date(now).toISOString() }
}

function RangeSelector({ value, onChange }: { value: RangeId; onChange: (id: RangeId) => void }) {
  return (
    <div className="inline-flex rounded-md border border-border p-0.5">
      {RANGES.map((r) => (
        <button
          key={r.id}
          onClick={() => onChange(r.id)}
          className={cn(
            'rounded px-2.5 py-1 text-xs transition-colors',
            value === r.id ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
          )}
        >
          {r.id}
        </button>
      ))}
    </div>
  )
}

export function DataProcessingPage() {
  const { t } = useTranslation()
  const [range, setRange] = useState<RangeId>('24h')
  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-semibold">
            <Activity size={18} strokeWidth={1.75} />
            {t('dataProcessing.title')}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">{t('dataProcessing.subtitle')}</p>
        </div>
        <RangeSelector value={range} onChange={setRange} />
      </header>

      <Headline t={t} range={range} />
      <Explorer t={t} range={range} />
    </div>
  )
}

/* ─── Headline KPIs + drop reasons (fixed 24h totals) ──────────────────── */

interface Totals {
  received: number
  parsing_dropped: number
  analysis_dropped: number
  correlation_dropped: number
}

function Headline({ t, range }: { t: TFunction; range: RangeId }) {
  const [totals, setTotals] = useState<Totals | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const w = rangeWindow(range)
      const [rec, par, ana, cor] = await Promise.all([
        ingestionHttpService.stats({ groupBy: 'dataType', status: 'received', top: 1, ...w }),
        ingestionHttpService.stats({ groupBy: 'dataType', status: 'parsing_dropped', top: 1, ...w }),
        ingestionHttpService.stats({ groupBy: 'dataType', status: 'analysis_dropped', top: 1, ...w }),
        ingestionHttpService.stats({ groupBy: 'dataType', status: 'correlation_dropped', top: 1, ...w }),
      ])
      setTotals({
        received: rec.total,
        parsing_dropped: par.total,
        analysis_dropped: ana.total,
        correlation_dropped: cor.total,
      })
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }, [range])

  useEffect(() => {
    void load()
  }, [load])

  const droppedTotal = totals
    ? totals.parsing_dropped + totals.analysis_dropped + totals.correlation_dropped
    : 0
  const denom = (totals?.received ?? 0) + droppedTotal
  const dropRate = denom > 0 ? (droppedTotal / denom) * 100 : 0
  const maxReason = Math.max(
    totals?.parsing_dropped ?? 0,
    totals?.analysis_dropped ?? 0,
    totals?.correlation_dropped ?? 0,
    1,
  )

  if (error) {
    return (
      <div className="mt-5 flex items-center gap-2 rounded-xl border border-border bg-card px-6 py-5 text-sm text-muted-foreground">
        <AlertTriangle size={15} className="text-amber-500" />
        {t('dataProcessing.ingestion.loadError')}
        <button onClick={() => void load()} className="ml-2 text-primary hover:underline">
          {t('dataProcessing.filters.retry')}
        </button>
      </div>
    )
  }

  return (
    <div className="mt-5 grid grid-cols-1 gap-3 md:grid-cols-[1fr_1fr_1.3fr]">
      <Kpi label={t('dataProcessing.kpi.received')} value={compact(totals?.received ?? 0)} loading={loading} tone="text-emerald-600 dark:text-emerald-400" />
      <Kpi
        label={t('dataProcessing.kpi.dropped')}
        value={compact(droppedTotal)}
        loading={loading}
        tone="text-amber-600 dark:text-amber-400"
        sub={t('dataProcessing.kpi.dropRate', { rate: dropRate.toFixed(dropRate < 10 ? 1 : 0) })}
      />
      <div className="rounded-xl border border-border bg-card p-5">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('dataProcessing.section.dropReasons')}</div>
        <ul className="mt-3 space-y-2">
          {DROP_REASONS.map((r) => (
            <li key={r} className="flex items-center gap-3 text-xs">
              <span className="w-32 shrink-0 truncate text-muted-foreground">{t(`dataProcessing.ingestion.status.${r}`)}</span>
              <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div className="h-full rounded-full bg-amber-500/70" style={{ width: `${((totals?.[r] ?? 0) / maxReason) * 100}%` }} />
              </div>
              <span className="w-12 text-right font-mono tabular-nums text-muted-foreground">{compact(totals?.[r] ?? 0)}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

function Kpi({
  label,
  value,
  sub,
  tone,
  loading,
}: {
  label: string
  value: string
  sub?: string
  tone?: string
  loading: boolean
}) {
  return (
    <div className="rounded-xl border border-border bg-card p-5">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-1 text-3xl font-semibold tabular-nums', tone)}>
        {loading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : value}
      </div>
      {sub && <div className="mt-0.5 text-xs text-muted-foreground">{sub}</div>}
    </div>
  )
}

/* ─── Explorer: timeline + breakdown driven by status + groupBy ────────── */

function Explorer({ t, range }: { t: TFunction; range: RangeId }) {
  const [groupBy, setGroupBy] = useState<GroupBy>('dataSource')
  const [status, setStatus] = useState<IngestionStatus>('received')
  const [stats, setStats] = useState<IngestionStats | null>(null)
  const [timeline, setTimeline] = useState<IngestionTimeline | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const w = rangeWindow(range)
      const [s, tl] = await Promise.all([
        ingestionHttpService.stats({ groupBy, status, top: 10, ...w }),
        ingestionHttpService.timeline({ status, ...w }),
      ])
      setStats(s)
      setTimeline(tl)
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }, [groupBy, status, range])

  useEffect(() => {
    void load()
  }, [load])

  const points = timeline?.points ?? []
  const buckets = stats?.buckets ?? []
  const maxBucket = Math.max(...buckets.map((b) => b.count), 1)

  return (
    <div className="mt-6 rounded-xl border border-border bg-card p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="text-sm font-semibold">{t('dataProcessing.section.overTime')}</div>
        <div className="flex items-center gap-2">
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value as IngestionStatus)}
            className="h-8 rounded-md border border-border bg-background px-2 text-xs"
          >
            {EXPLORER_STATUSES.map((s) => (
              <option key={s} value={s}>
                {t(`dataProcessing.ingestion.status.${s}`)}
              </option>
            ))}
          </select>
          <button
            onClick={() => void load()}
            title={t('dataProcessing.filters.refresh')}
            className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
          </button>
        </div>
      </div>

      {error ? (
        <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
          <AlertTriangle size={15} className="text-amber-500" />
          {t('dataProcessing.ingestion.loadError')}
        </div>
      ) : (
        <>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-2xl font-semibold tabular-nums">{compact(stats?.total ?? 0)}</span>
            <span className="font-mono text-xs text-muted-foreground">· {range}</span>
          </div>

          <TimelineChart points={points} />

          <div className="mt-5 flex items-center justify-between">
            <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('dataProcessing.section.top')}</div>
            <div className="inline-flex rounded-md border border-border p-0.5">
              {(['dataSource', 'dataType'] as GroupBy[]).map((g) => (
                <button
                  key={g}
                  onClick={() => setGroupBy(g)}
                  className={cn(
                    'rounded px-2.5 py-1 text-xs transition-colors',
                    groupBy === g ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  {t(`dataProcessing.ingestion.groupBy.${g}`)}
                </button>
              ))}
            </div>
          </div>

          <ul className="mt-3 space-y-1.5">
            {buckets.map((b) => (
              <li key={b.key} className="flex items-center gap-3 text-xs">
                <span className="w-44 shrink-0 truncate font-mono text-muted-foreground" title={b.key}>
                  {b.key || '—'}
                </span>
                <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                  <div className="h-full rounded-full bg-sky-500/70" style={{ width: `${(b.count / maxBucket) * 100}%` }} />
                </div>
                <span className="w-14 text-right font-mono tabular-nums text-muted-foreground">{compact(b.count)}</span>
              </li>
            ))}
            {!loading && buckets.length === 0 && (
              <li className="text-xs text-muted-foreground">{t('dataProcessing.ingestion.noData')}</li>
            )}
          </ul>
        </>
      )}
    </div>
  )
}

function TimelineChart({ points }: { points: { timestamp: string; count: number }[] }) {
  if (points.length < 2) return <div className="mt-4 h-24" />
  const w = 1000
  const h = 100
  const max = Math.max(...points.map((p) => p.count), 1) * 1.15
  const xs = points.map((_, i) => (i * w) / (points.length - 1))
  const ys = points.map((p) => h - (p.count / max) * h)
  const line = points.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const cx1 = xs[i - 1] + (xs[i] - xs[i - 1]) / 2
    const cx2 = xs[i] - (xs[i] - xs[i - 1]) / 2
    return `${acc} C ${cx1} ${ys[i - 1]}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const area = `${line} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
      <defs>
        <linearGradient id="dpGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="rgb(14 165 233)" stopOpacity="0.28" />
          <stop offset="100%" stopColor="rgb(14 165 233)" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={area} fill="url(#dpGrad)" />
      <path d={line} fill="none" stroke="rgb(14 165 233)" strokeOpacity="0.85" strokeWidth="1.75" strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  )
}
