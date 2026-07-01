import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { AlertTriangle, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { ingestionHttpService } from '../services/data-processing-http.service'
import type { GroupBy, IngestionBucket, IngestionStatus } from '../types/data-processing.types'
import { AreaChart, Donut, Sparkline, Treemap, compact, relTime, useCountUp, useMounted, type TreemapItem } from '../components/viz'

/* ─── Status palette ───────────────────────────────────────────────────── */

const DROP_STATUSES = ['parsing_dropped', 'analysis_dropped', 'correlation_dropped'] as const
type DropStatus = (typeof DROP_STATUSES)[number]

const COLORS = {
  received: 'rgb(16 185 129)', // emerald
  dropped: 'rgb(245 158 11)', // amber
  parsing_dropped: 'rgb(245 158 11)', // amber
  analysis_dropped: 'rgb(249 115 22)', // orange
  correlation_dropped: 'rgb(244 63 94)', // rose
  stored: 'rgb(14 165 233)', // sky
}
// Cycling palette for the breakdown / treemap.
const PALETTE = [
  'rgb(14 165 233)',
  'rgb(99 102 241)',
  'rgb(168 85 247)',
  'rgb(236 72 153)',
  'rgb(244 63 94)',
  'rgb(249 115 22)',
  'rgb(234 179 8)',
  'rgb(16 185 129)',
  'rgb(20 184 166)',
  'rgb(6 182 212)',
]

const RANGES = [
  { id: '1h', hours: 1 },
  { id: '24h', hours: 24 },
  { id: '7d', hours: 24 * 7 },
  { id: '30d', hours: 24 * 30 },
] as const
type RangeId = (typeof RANGES)[number]['id']

function rangeWindow(id: RangeId): { from: string; to: string; minutes: number } {
  const hours = RANGES.find((r) => r.id === id)?.hours ?? 24
  const now = Date.now()
  return { from: new Date(now - hours * 3_600_000).toISOString(), to: new Date(now).toISOString(), minutes: hours * 60 }
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function DataProcessingPage() {
  const { t } = useTranslation()
  const [range, setRange] = useState<RangeId>('24h')
  const [nonce, setNonce] = useState(0)

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-end gap-3">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setNonce((n) => n + 1)}
            title={t('dataProcessing.filters.refresh')}
            className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <RefreshCw size={14} />
          </button>
          <RangeSelector value={range} onChange={setRange} />
        </div>
      </header>

      <Overview t={t} range={range} nonce={nonce} />
      <Breakdown t={t} range={range} nonce={nonce} />
    </div>
  )
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

/* ─── Overview: pipeline funnel + KPIs + timeline + drop donut ─────────── */

interface OverviewData {
  totals: Record<'received' | DropStatus, number>
  series: { received: { t: string; v: number }[]; dropped: { t: string; v: number }[] }
}

function useOverview(range: RangeId, nonce: number) {
  const [data, setData] = useState<OverviewData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    const w = rangeWindow(range)
    Promise.all([
      ingestionHttpService.timeline({ status: 'received', from: w.from, to: w.to }),
      ingestionHttpService.timeline({ status: 'parsing_dropped', from: w.from, to: w.to }),
      ingestionHttpService.timeline({ status: 'analysis_dropped', from: w.from, to: w.to }),
      ingestionHttpService.timeline({ status: 'correlation_dropped', from: w.from, to: w.to }),
    ])
      .then(([rec, par, ana, cor]) => {
        if (cancelled) return
        const sum = (pts?: { count: number }[]) => (pts ?? []).reduce((a, p) => a + p.count, 0)
        const recPts = rec.points ?? []
        // Align drop series to the received timeline's timestamps.
        const dropMap = new Map<string, number>()
        for (const tl of [par, ana, cor]) for (const p of tl.points ?? []) dropMap.set(p.timestamp, (dropMap.get(p.timestamp) ?? 0) + p.count)
        setData({
          totals: {
            received: sum(recPts),
            parsing_dropped: sum(par.points),
            analysis_dropped: sum(ana.points),
            correlation_dropped: sum(cor.points),
          },
          series: {
            received: recPts.map((p) => ({ t: p.timestamp, v: p.count })),
            dropped: recPts.map((p) => ({ t: p.timestamp, v: dropMap.get(p.timestamp) ?? 0 })),
          },
        })
      })
      .catch(() => !cancelled && setError(true))
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [range, nonce])

  return { data, loading, error }
}

function Overview({ t, range, nonce }: { t: TFunction; range: RangeId; nonce: number }) {
  const { data, loading, error } = useOverview(range, nonce)

  if (error) {
    return (
      <div className="mt-5 flex items-center gap-2 rounded-xl border border-border bg-card px-6 py-5 text-sm text-muted-foreground">
        <AlertTriangle size={15} className="text-amber-500" />
        {t('dataProcessing.ingestion.loadError')}
      </div>
    )
  }

  const totals = data?.totals ?? { received: 0, parsing_dropped: 0, analysis_dropped: 0, correlation_dropped: 0 }
  const dropped = totals.parsing_dropped + totals.analysis_dropped + totals.correlation_dropped
  const stored = Math.max(0, totals.received - dropped)
  const denom = totals.received || 1
  const dropRate = (dropped / (totals.received || 1)) * 100
  const survival = 100 - dropRate
  const minutes = rangeWindow(range).minutes
  const perMin = minutes > 0 ? totals.received / minutes : 0

  return (
    <div className="mt-5 space-y-5">
      {/* Pipeline funnel */}
      <PipelineFunnel t={t} totals={totals} stored={stored} denom={denom} loading={loading} />

      {/* KPIs */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <Kpi label={t('dataProcessing.kpi.received')} value={totals.received} tone={COLORS.received} spark={data?.series.received.map((p) => p.v)} loading={loading} />
        <Kpi label={t('dataProcessing.kpi.stored')} value={stored} tone={COLORS.stored} sub={t('dataProcessing.kpi.survival', { rate: survival.toFixed(survival < 99.95 ? 1 : 0) })} loading={loading} />
        <Kpi label={t('dataProcessing.kpi.dropped')} value={dropped} tone={COLORS.dropped} sub={t('dataProcessing.kpi.dropRate', { rate: dropRate.toFixed(dropRate < 10 ? 1 : 0) })} spark={data?.series.dropped.map((p) => p.v)} loading={loading} />
        <Kpi label={t('dataProcessing.kpi.throughput')} value={perMin} fmt={(n) => t('dataProcessing.kpi.perMin', { value: compact(n) })} tone="rgb(168 85 247)" loading={loading} />
      </div>

      {/* Timeline + donut */}
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-[1.7fr_1fr]">
        <div className="rounded-xl border border-border bg-card p-5">
          <div className="mb-1 flex items-center justify-between">
            <div className="text-sm font-semibold">{t('dataProcessing.chart.title')}</div>
            <Legend items={[{ color: COLORS.received, label: t('dataProcessing.chart.received') }, { color: COLORS.dropped, label: t('dataProcessing.chart.dropped') }]} />
          </div>
          {loading && !data ? (
            <Center><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></Center>
          ) : (data?.series.received.length ?? 0) < 2 ? (
            <Center>{t('dataProcessing.ingestion.noData')}</Center>
          ) : (
            <AreaChart
              series={[
                { key: 'received', label: t('dataProcessing.chart.received'), color: COLORS.received, points: data!.series.received },
                { key: 'dropped', label: t('dataProcessing.chart.dropped'), color: COLORS.dropped, points: data!.series.dropped },
              ]}
            />
          )}
        </div>

        <div className="rounded-xl border border-border bg-card p-5">
          <div className="text-sm font-semibold">{t('dataProcessing.donut.title')}</div>
          <div className="mt-2 flex items-center gap-5">
            <div className="relative shrink-0">
              <Donut segments={DROP_STATUSES.map((s) => ({ value: totals[s], color: COLORS[s], label: s }))} />
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-xl font-semibold tabular-nums">{compact(dropped)}</span>
                <span className="text-[10px] text-muted-foreground">{t('dataProcessing.donut.dropped')}</span>
              </div>
            </div>
            <ul className="flex-1 space-y-2">
              {DROP_STATUSES.map((s) => (
                <li key={s} className="flex items-center gap-2 text-xs">
                  <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ backgroundColor: COLORS[s] }} />
                  <span className="min-w-0 truncate text-muted-foreground">{t(`dataProcessing.ingestion.status.${s}`)}</span>
                  <span className="ml-auto font-mono tabular-nums">{compact(totals[s])}</span>
                  <span className="w-9 text-right font-mono text-[10px] text-muted-foreground">{dropped > 0 ? Math.round((totals[s] / dropped) * 100) : 0}%</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Pipeline funnel ──────────────────────────────────────────────────── */

function PipelineFunnel({ t, totals, stored, denom, loading }: { t: TFunction; totals: Record<'received' | DropStatus, number>; stored: number; denom: number; loading: boolean }) {
  const mounted = useMounted()
  const afterParsing = Math.max(0, totals.received - totals.parsing_dropped)
  const afterAnalysis = Math.max(0, afterParsing - totals.analysis_dropped)
  const stages: { key: string; label: string; value: number; drop?: number; dropLabel?: string; color: string }[] = [
    { key: 'received', label: t('dataProcessing.pipeline.stages.received'), value: totals.received, color: COLORS.received },
    { key: 'parsed', label: t('dataProcessing.pipeline.stages.parsed'), value: afterParsing, drop: totals.parsing_dropped, dropLabel: t('dataProcessing.ingestion.status.parsing_dropped'), color: COLORS.received },
    { key: 'analyzed', label: t('dataProcessing.pipeline.stages.analyzed'), value: afterAnalysis, drop: totals.analysis_dropped, dropLabel: t('dataProcessing.ingestion.status.analysis_dropped'), color: COLORS.stored },
    { key: 'stored', label: t('dataProcessing.pipeline.stages.stored'), value: stored, drop: totals.correlation_dropped, dropLabel: t('dataProcessing.ingestion.status.correlation_dropped'), color: COLORS.stored },
  ]
  return (
    <div className="rounded-xl border border-border bg-card p-5">
      <div className="mb-4 text-sm font-semibold">{t('dataProcessing.pipeline.title')}</div>
      <div className="space-y-1">
        {stages.map((s, i) => {
          const pct = (s.value / denom) * 100
          return (
            <div key={s.key}>
              {s.drop != null && s.drop > 0 && (
                <div className="flex items-center gap-2 py-1 pl-[136px] text-[11px] text-amber-600/90 dark:text-amber-400/90">
                  <span className="font-mono">↓ −{compact(s.drop)}</span>
                  <span className="text-muted-foreground">{s.dropLabel}</span>
                </div>
              )}
              <div className="flex items-center gap-3">
                <span className="w-[124px] shrink-0 text-xs text-muted-foreground">{s.label}</span>
                <div className="relative h-9 flex-1 overflow-hidden rounded-md bg-muted/40">
                  <div
                    className="flex h-full items-center rounded-md px-3 text-xs font-medium text-white"
                    style={{
                      width: mounted ? `${Math.max(pct, 3)}%` : '0%',
                      background: `linear-gradient(90deg, ${s.color}, ${s.color}cc)`,
                      transition: `width 800ms cubic-bezier(0.22,1,0.36,1) ${i * 90}ms`,
                    }}
                  >
                    <span className="tabular-nums drop-shadow">{loading ? '' : compact(s.value)}</span>
                  </div>
                </div>
                <span className="w-12 shrink-0 text-right font-mono text-xs tabular-nums text-muted-foreground">{pct.toFixed(pct < 99.95 ? 1 : 0)}%</span>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

/* ─── KPI card ─────────────────────────────────────────────────────────── */

function Kpi({ label, value, sub, tone, spark, fmt, loading }: { label: string; value: number; sub?: string; tone?: string; spark?: number[]; fmt?: (n: number) => string; loading: boolean }) {
  const animated = useCountUp(value)
  const display = fmt ? fmt(animated) : compact(animated)
  return (
    <div className="relative overflow-hidden rounded-xl border border-border bg-card p-5">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-1 text-3xl font-semibold tabular-nums" style={{ color: tone }}>
        {loading ? <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /> : display}
      </div>
      {sub && <div className="mt-0.5 text-xs text-muted-foreground">{sub}</div>}
      {spark && spark.length > 1 && <Sparkline points={spark} color={tone} className="mt-2 h-7 w-full" />}
    </div>
  )
}

/* ─── Breakdown: bars / treemap, top-N, share + last seen ──────────────── */

const INITIAL_N = 12
const MAX_N = 60

function Breakdown({ t, range, nonce }: { t: TFunction; range: RangeId; nonce: number }) {
  const [groupBy, setGroupBy] = useState<GroupBy>('dataSource')
  const [status, setStatus] = useState<IngestionStatus>('received')
  const [view, setView] = useState<'bars' | 'treemap'>('bars')
  const [showAll, setShowAll] = useState(false)
  const [buckets, setBuckets] = useState<IngestionBucket[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const w = rangeWindow(range)
      const s = await ingestionHttpService.stats({ groupBy, status, top: MAX_N, from: w.from, to: w.to })
      setBuckets(s.buckets ?? [])
      setTotal(s.total ?? 0)
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }, [groupBy, status, range, nonce])

  useEffect(() => {
    void load()
  }, [load])

  const shown = showAll ? buckets : buckets.slice(0, INITIAL_N)
  const max = Math.max(...buckets.map((b) => b.count), 1)
  const treemapItems: TreemapItem[] = useMemo(
    () => buckets.slice(0, 24).map((b, i) => ({ key: b.key || '—', value: b.count, color: PALETTE[i % PALETTE.length] })),
    [buckets],
  )

  return (
    <div className="mt-5 rounded-xl border border-border bg-card p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="text-sm font-semibold">{t('dataProcessing.breakdown.title')}</div>
        <div className="flex flex-wrap items-center gap-2">
          <select value={status} onChange={(e) => setStatus(e.target.value as IngestionStatus)} className="h-8 rounded-md border border-border bg-background px-2 text-xs">
            {(['received', 'parsing_dropped', 'analysis_dropped', 'correlation_dropped', 'all'] as IngestionStatus[]).map((s) => (
              <option key={s} value={s}>{t(`dataProcessing.ingestion.status.${s}`)}</option>
            ))}
          </select>
          <Segmented value={groupBy} onChange={(v) => setGroupBy(v as GroupBy)} options={(['dataSource', 'dataType'] as GroupBy[]).map((g) => ({ id: g, label: t(`dataProcessing.ingestion.groupBy.${g}`) }))} />
          <Segmented value={view} onChange={(v) => setView(v as 'bars' | 'treemap')} options={[{ id: 'bars', label: t('dataProcessing.breakdown.bars') }, { id: 'treemap', label: t('dataProcessing.breakdown.treemap') }]} />
        </div>
      </div>

      {error ? (
        <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
          <AlertTriangle size={15} className="text-amber-500" /> {t('dataProcessing.ingestion.loadError')}
        </div>
      ) : loading && buckets.length === 0 ? (
        <Center><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></Center>
      ) : buckets.length === 0 ? (
        <Center>{t('dataProcessing.ingestion.noData')}</Center>
      ) : view === 'treemap' ? (
        <div className="mt-4">
          <Treemap items={treemapItems} />
        </div>
      ) : (
        <>
          <ul className="mt-4 space-y-2.5">
            {shown.map((b, i) => (
              <BreakdownRow key={b.key || i} rank={i + 1} bucket={b} max={max} total={total} color={PALETTE[i % PALETTE.length]} delay={i * 35} t={t} />
            ))}
          </ul>
          {buckets.length > INITIAL_N && (
            <button onClick={() => setShowAll((v) => !v)} className="mt-3 text-xs text-primary hover:underline">
              {showAll ? t('dataProcessing.breakdown.showLess') : t('dataProcessing.breakdown.showAll', { count: buckets.length })}
            </button>
          )}
        </>
      )}
    </div>
  )
}

function BreakdownRow({ rank, bucket, max, total, color, delay, t }: { rank: number; bucket: IngestionBucket; max: number; total: number; color: string; delay: number; t: TFunction }) {
  const mounted = useMounted()
  const share = total > 0 ? (bucket.count / total) * 100 : 0
  const seen = relTime(bucket.lastSeen)
  return (
    <li className="flex items-center gap-3 text-xs">
      <span className="w-5 shrink-0 text-right font-mono text-[10px] text-muted-foreground">{rank}</span>
      <span className="w-48 shrink-0 truncate font-mono text-foreground/80" title={bucket.key}>{bucket.key || '—'}</span>
      <div className="relative h-2 flex-1 overflow-hidden rounded-full bg-muted">
        <div className="h-full rounded-full" style={{ width: mounted ? `${(bucket.count / max) * 100}%` : '0%', backgroundColor: color, transition: `width 700ms cubic-bezier(0.22,1,0.36,1) ${delay}ms` }} />
      </div>
      <span className="w-14 shrink-0 text-right font-mono tabular-nums">{compact(bucket.count)}</span>
      <span className="w-10 shrink-0 text-right font-mono text-[10px] text-muted-foreground">{share.toFixed(share < 10 ? 1 : 0)}%</span>
      <span className="hidden w-14 shrink-0 text-right font-mono text-[10px] text-muted-foreground sm:inline" title={t('dataProcessing.breakdown.lastSeen')}>
        {seen ? t('dataProcessing.breakdown.ago', { time: seen }) : '—'}
      </span>
    </li>
  )
}

/* ─── Small UI bits ────────────────────────────────────────────────────── */

function Segmented({ value, onChange, options }: { value: string; onChange: (v: string) => void; options: { id: string; label: string }[] }) {
  return (
    <div className="inline-flex rounded-md border border-border p-0.5">
      {options.map((o) => (
        <button
          key={o.id}
          onClick={() => onChange(o.id)}
          className={cn('rounded px-2.5 py-1 text-xs transition-colors', value === o.id ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
        >
          {o.label}
        </button>
      ))}
    </div>
  )
}

function Legend({ items }: { items: { color: string; label: string }[] }) {
  return (
    <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
      {items.map((i) => (
        <span key={i.label} className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-full" style={{ backgroundColor: i.color }} />
          {i.label}
        </span>
      ))}
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex h-48 items-center justify-center text-sm text-muted-foreground">{children}</div>
}
