import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Activity,
  AlertTriangle,
  AtSign,
  BadgeCheck,
  Crosshair,
  Database,
  Flame,
  Paperclip,
  Server,
  ShieldAlert,
  Sparkles,
  TrendingUp,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useSocAi } from '@/features/soc-ai/SocAiProvider'
import { useSocAiConfigured } from '@/features/soc-ai/lib/useSocAiConfig'
import {
  useActivePlaybooks,
  useAlertKpis,
  useComplianceScores,
  useDatasourcesOverview,
  useOpenIncidents,
  useSystemHealth,
  useTopAssets,
  useTopTechniques,
  type HealthStatus,
} from '../hooks/use-overview'

type TFn = ReturnType<typeof useTranslation>['t']

export function HomePage() {
  const { t } = useTranslation()
  const alerts = useAlertKpis()
  const incidents = useOpenIncidents()
  const playbooks = useActivePlaybooks()

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-10">
      <ChatHero />

      <div className="mt-10 grid grid-cols-12 gap-4">
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={AlertTriangle}
            label={t('home.kpis.alerts24h')}
            value={fmtCount(alerts.alerts24h)}
            loading={alerts.isLoading}
            sparkline={alerts.sparkline}
            accent="text-red-500"
            href="/threat-management/alerts"
          />
        </div>
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={Flame}
            label={t('home.kpis.openIncidents')}
            value={fmtCount(incidents.count)}
            loading={incidents.isLoading}
            sublabel={t('home.kpis.openIncidentsSub')}
            accent="text-amber-500"
            href="/threat-management/incidents"
          />
        </div>
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={ShieldAlert}
            label={t('home.kpis.highSeverity24h')}
            value={fmtCount(alerts.high24h)}
            loading={alerts.isLoading}
            sublabel={t('home.kpis.highSeveritySub', { n: fmtCount(alerts.alerts24h) })}
            accent="text-rose-500"
            href="/threat-management/alerts"
          />
        </div>
        <div className="col-span-6 md:col-span-3">
          <KpiTile
            icon={Workflow}
            label={t('home.kpis.activePlaybooks')}
            value={fmtCount(playbooks.count)}
            loading={playbooks.isLoading}
            sublabel={t('home.kpis.activePlaybooksSub')}
            accent="text-violet-500"
          />
        </div>

        <div className="col-span-12 lg:col-span-8">
          <DatasourcesCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <SystemHealthCard />
        </div>

        <div className="col-span-12 lg:col-span-6">
          <MitreTechniquesCard />
        </div>
        <div className="col-span-12 lg:col-span-6">
          <TopAssetsCard />
        </div>

        <div className="col-span-12">
          <ComplianceCard />
        </div>
      </div>
    </div>
  )
}

/* ─── Number / time helpers ────────────────────────────────────────────────── */

function fmtCount(n: number): string {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return n.toLocaleString()
}

function spanLabel(t: TFn, from?: string, to?: string): string {
  if (!from || !to) return t('home.span.recent')
  const ms = new Date(to).getTime() - new Date(from).getTime()
  if (!Number.isFinite(ms) || ms <= 0) return t('home.span.recent')
  const days = Math.round(ms / 86_400_000)
  if (days >= 1) return t('home.span.days', { count: days })
  const hours = Math.max(1, Math.round(ms / 3_600_000))
  return t('home.span.hours', { n: hours })
}

/* ─── AI chat hero ─────────────────────────────────────────────────────────── */

const SUGGESTION_KEYS = ['threatHunt', 'observableTriage', 'writeRule', 'systemStatus'] as const

function ChatHero() {
  const { t } = useTranslation()
  const configured = useSocAiConfigured()
  const { submit } = useSocAi()
  const [value, setValue] = useState('')

  // Hidden entirely until SOC-AI has a provider configured — same gate the
  // floating composer/panel use, so we never show a dead chat box.
  if (!configured) return null

  const send = () => {
    const text = value.trim()
    if (!text) return
    submit(text)
    setValue('')
  }

  return (
    <section className="flex flex-col items-center">
      <h2 className="mb-4 text-center text-base font-medium text-foreground">{t('home.hero.title')}</h2>
      <div className="relative w-full max-w-3xl">
        <div
          aria-hidden
          className="absolute -inset-px rounded-xl bg-[conic-gradient(from_180deg_at_50%_50%,#1a8cff_0deg,#0ea5e9_120deg,#2563eb_240deg,#1a8cff_360deg)] opacity-50 blur-[1px]"
        />
        <div className="relative rounded-xl bg-card p-4">
          <textarea
            rows={3}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            placeholder={t('home.hero.placeholder')}
            className="w-full resize-none border-0 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
          />
          <div className="mt-2 flex items-center justify-between">
            <div className="flex items-center gap-2 text-muted-foreground">
              <IconBtn label={t('home.hero.mention')}><AtSign size={16} strokeWidth={1.75} /></IconBtn>
              <IconBtn label={t('home.hero.attach')}><Paperclip size={16} strokeWidth={1.75} /></IconBtn>
            </div>
            <button
              onClick={send}
              disabled={!value.trim()}
              className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm hover:opacity-90 disabled:opacity-40"
              aria-label={t('home.hero.send')}
            >
              <Sparkles size={14} strokeWidth={2} />
            </button>
          </div>
        </div>
      </div>
      <div className="mt-4 flex flex-wrap justify-center gap-2">
        {SUGGESTION_KEYS.map((key) => {
          const label = t(`home.hero.suggestions.${key}`)
          return (
            <button
              key={key}
              onClick={() => submit(label)}
              className="rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              {label}
            </button>
          )
        })}
      </div>
    </section>
  )
}

function IconBtn({ children, label }: { children: React.ReactNode; label: string }) {
  return (
    <button
      aria-label={label}
      title={label}
      className="flex h-7 w-7 items-center justify-center rounded-md hover:bg-muted hover:text-foreground"
    >
      {children}
    </button>
  )
}

/* ─── KPI tiles ────────────────────────────────────────────────────────────── */

function KpiTile({
  icon: Icon,
  label,
  value,
  sublabel,
  sparkline,
  accent,
  loading,
  href,
}: {
  icon: LucideIcon
  label: string
  value: string
  sublabel?: string
  sparkline?: number[]
  accent: string
  loading?: boolean
  href?: string
}) {
  const inner = (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-4 transition-colors hover:border-primary/40">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon size={14} strokeWidth={1.75} className={accent} />
        {label}
      </div>
      {loading ? (
        <div className="mt-2 h-8 w-16 animate-pulse rounded bg-muted" />
      ) : (
        <div className="mt-2 text-3xl font-semibold tracking-tight tabular-nums">{value}</div>
      )}
      <div className="mt-3 -mx-1 min-h-[36px]">
        {sparkline && sparkline.length > 1 ? (
          <Sparkline data={sparkline} accent={accent} />
        ) : sublabel ? (
          <span className="text-xs text-muted-foreground">{sublabel}</span>
        ) : null}
      </div>
    </div>
  )
  return href ? (
    <Link to={href} className="block h-full">
      {inner}
    </Link>
  ) : (
    inner
  )
}

function Sparkline({ data, accent }: { data: number[]; accent: string }) {
  const w = 200
  const h = 36
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - ((v - min) / range) * (h - 4) - 2)
  const path = data.map((_, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys[i]}`).join(' ')
  const area = `${path} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-9 w-full" preserveAspectRatio="none">
      <path d={area} className={cn('opacity-15', accent.replace('text-', 'fill-'))} />
      <path d={path} fill="none" strokeWidth="1.75" strokeLinejoin="round" className={cn('stroke-current', accent)} />
    </svg>
  )
}

/* ─── Datasources ──────────────────────────────────────────────────────────── */

function DatasourcesCard() {
  const { t } = useTranslation()
  const ds = useDatasourcesOverview()
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.datasources.title')} icon={Database} action={{ label: t('home.datasources.viewAll'), href: '/datasources' }} />
      <div className="flex flex-wrap items-end gap-x-10 gap-y-4 px-6 pt-2">
        <div>
          {ds.isLoading ? (
            <div className="h-12 w-20 animate-pulse rounded bg-muted" />
          ) : (
            <div className="text-5xl font-semibold leading-none tracking-tight text-emerald-500 tabular-nums dark:text-emerald-400">
              {fmtCount(ds.activeSources)}
            </div>
          )}
          <div className="mt-2 text-xs text-muted-foreground">
            <span className="text-foreground">{t('home.datasources.ofTotal', { n: fmtCount(ds.total) })}</span>{' '}
            {t('home.datasources.datasources')}
            <br />
            {t('home.datasources.sendingData')}
          </div>
        </div>
        <div>
          <div className="flex items-baseline gap-2">
            {ds.isLoading ? (
              <div className="h-12 w-28 animate-pulse rounded bg-muted" />
            ) : (
              <span className="text-5xl font-semibold leading-none tracking-tight tabular-nums">
                {fmtCount(ds.events)}
              </span>
            )}
            <TrendingUp size={20} strokeWidth={2} className="text-emerald-500 dark:text-emerald-400" />
          </div>
          <div className="mt-2 text-xs text-muted-foreground">
            {t('home.datasources.totalEvents')}
            <br />
            {spanLabel(t, ds.from, ds.to)}
          </div>
        </div>
      </div>
      <div className="px-2 pb-2 pt-4">
        <EventsChart points={ds.points} loading={ds.isLoading} />
      </div>
    </div>
  )
}

function EventsChart({ points, loading }: { points: { t: string; count: number }[]; loading: boolean }) {
  const { t } = useTranslation()
  const w = 700
  const h = 180
  const pl = 44
  const pr = 16
  const pt = 10
  const pb = 24
  const innerW = w - pl - pr
  const innerH = h - pt - pb

  if (loading) {
    return <div className="mx-4 h-44 animate-pulse rounded-lg bg-muted" />
  }
  if (points.length < 2) {
    return (
      <div className="flex h-44 items-center justify-center text-xs text-muted-foreground">
        {t('home.datasources.noData')}
      </div>
    )
  }

  const counts = points.map((p) => p.count)
  const max = Math.max(...counts, 1)
  const xs = points.map((_, i) => pl + (i * innerW) / (points.length - 1))
  const ys = points.map((v) => pt + innerH - (v.count / max) * innerH)
  const linePath = points.map((_, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys[i]}`).join(' ')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${pt + innerH} L ${xs[0]} ${pt + innerH} Z`

  // Two reference gridlines (half + full scale) with compact labels.
  const grids = [max / 2, max]
  // A handful of evenly-spaced x labels so the axis stays readable.
  const labelEvery = Math.max(1, Math.ceil(points.length / 6))

  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-44 w-full" preserveAspectRatio="none">
      <defs>
        <pattern id="hatch" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <line x1="0" y1="0" x2="0" y2="6" className="stroke-foreground/5" strokeWidth="1" />
        </pattern>
        <linearGradient id="lineGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stopColor="#22d3ee" />
          <stop offset="1" stopColor="#1a8cff" />
        </linearGradient>
      </defs>
      <rect x={pl} y={pt} width={innerW} height={innerH} fill="url(#hatch)" />
      <g className="fill-muted-foreground" fontSize="10">
        {grids.map((v) => (
          <text key={v} x={pl - 6} y={pt + innerH - (v / max) * innerH + 3} textAnchor="end">
            {fmtCount(Math.round(v))}
          </text>
        ))}
      </g>
      {grids.map((v) => (
        <line
          key={v}
          x1={pl}
          x2={pl + innerW}
          y1={pt + innerH - (v / max) * innerH}
          y2={pt + innerH - (v / max) * innerH}
          className="stroke-border/40"
          strokeDasharray="2 4"
        />
      ))}
      <path d={areaPath} className="fill-primary/10" />
      <path d={linePath} fill="none" stroke="url(#lineGrad)" strokeWidth="2" strokeLinejoin="round" />
      <g className="fill-muted-foreground" fontSize="10">
        {points.map((p, i) =>
          i % labelEvery === 0 ? (
            <text key={i} x={xs[i]} y={h - 6} textAnchor="middle">
              {shortTime(p.t)}
            </text>
          ) : null,
        )}
      </g>
    </svg>
  )
}

function shortTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric' })
}

/* ─── MITRE ATT&CK top techniques ──────────────────────────────────────────── */

function MitreTechniquesCard() {
  const { t } = useTranslation()
  const { items, isLoading } = useTopTechniques()
  const max = Math.max(...items.map((m) => m.count), 1)
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader
        title={t('home.mitre.title')}
        icon={Crosshair}
        iconClass="text-rose-500"
        action={{ label: t('home.mitre.explore'), href: '/threat-management/alerts' }}
      />
      <div className="space-y-2.5 px-6 pb-5">
        {isLoading ? (
          <SkeletonRows rows={5} />
        ) : items.length === 0 ? (
          <EmptyState text={t('home.mitre.empty')} />
        ) : (
          items.map((m) => (
            <div key={m.value} className="grid grid-cols-[1fr_auto] items-center gap-3 text-xs">
              <div>
                <span className="font-medium">{m.value}</span>
                <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-sky-400 to-blue-600"
                    style={{ width: `${(m.count / max) * 100}%` }}
                  />
                </div>
              </div>
              <div className="font-mono text-sm tabular-nums">{fmtCount(m.count)}</div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

/* ─── System health (real backend / OpenSearch / SOC-AI) ───────────────────── */

const HEALTH_META: Record<HealthStatus, { dot: string; badge: string }> = {
  up: { dot: 'bg-emerald-500', badge: 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300' },
  degraded: { dot: 'bg-amber-500', badge: 'bg-amber-500/15 text-amber-600 dark:text-amber-300' },
  down: { dot: 'bg-red-500', badge: 'bg-red-500/15 text-red-500 dark:text-red-300' },
  unknown: { dot: 'bg-muted-foreground/50', badge: 'bg-muted text-muted-foreground' },
}

function SystemHealthCard() {
  const { t } = useTranslation()
  const { services, isLoading } = useSystemHealth()
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.health.title')} icon={Activity} iconClass="text-emerald-500" action={{ label: t('home.health.details'), href: '/settings/about' }} />
      <div className="divide-y divide-border">
        {isLoading ? (
          <div className="px-6 py-4">
            <SkeletonRows rows={3} />
          </div>
        ) : (
          services.map((p) => {
            const meta = HEALTH_META[p.status]
            return (
              <div key={p.name} className="flex items-center justify-between px-6 py-2.5 text-xs">
                <div className="flex items-center gap-2">
                  <span className={cn('h-2 w-2 rounded-full', meta.dot)} />
                  <span>{p.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[10px] text-muted-foreground">{p.detail}</span>
                  <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none', meta.badge)}>
                    {t(`about.status.${p.status}`)}
                  </span>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}

/* ─── Top targeted assets (by alert volume) ────────────────────────────────── */

function TopAssetsCard() {
  const { t } = useTranslation()
  const { items, isLoading } = useTopAssets()
  const max = Math.max(...items.map((a) => a.count), 1)
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.assets.title')} icon={Server} action={{ label: t('home.assets.viewAll'), href: '/datasources' }} />
      <div className="grid grid-cols-[1fr_70px] gap-3 px-6 pb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>{t('home.assets.asset')}</div>
        <div className="text-right">{t('home.assets.alerts')}</div>
      </div>
      <div className="border-t border-border">
        {isLoading ? (
          <div className="px-6 py-4">
            <SkeletonRows rows={5} />
          </div>
        ) : items.length === 0 ? (
          <div className="px-6 py-6">
            <EmptyState text={t('home.assets.empty')} />
          </div>
        ) : (
          items.map((a) => (
            <div
              key={a.value}
              className="grid grid-cols-[1fr_70px] items-center gap-3 border-b border-border px-6 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
            >
              <div className="min-w-0">
                <div className="truncate font-mono text-[12px]">{a.value}</div>
                <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-rose-400 to-red-600"
                    style={{ width: `${(a.count / max) * 100}%` }}
                  />
                </div>
              </div>
              <div className="text-right font-mono tabular-nums">{fmtCount(a.count)}</div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

/* ─── Compliance score (latest snapshot per framework) ─────────────────────── */

function ComplianceCard() {
  const { t } = useTranslation()
  const { items, isLoading } = useComplianceScores()
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.compliance.title')} icon={BadgeCheck} iconClass="text-emerald-500" action={{ label: t('home.compliance.reports'), href: '/compliance' }} />
      <div className="px-6 pb-6">
        {isLoading ? (
          <div className="h-32 animate-pulse rounded-lg bg-muted" />
        ) : items.length === 0 ? (
          <EmptyState text={t('home.compliance.empty')} />
        ) : (
          <>
            <div
              className="grid items-end gap-3 h-32"
              style={{ gridTemplateColumns: `repeat(${items.length}, minmax(0, 1fr))` }}
            >
              {items.map((c) => (
                <div key={c.key} className="flex h-full flex-col items-center justify-end">
                  <div className="mb-1 font-mono text-[10px] tabular-nums">{c.score}%</div>
                  <div
                    className={cn(
                      'w-full rounded-t-md',
                      c.score >= 90
                        ? 'bg-gradient-to-t from-emerald-600 to-emerald-400'
                        : c.score >= 80
                          ? 'bg-gradient-to-t from-sky-600 to-sky-400'
                          : c.score >= 50
                            ? 'bg-gradient-to-t from-amber-600 to-amber-400'
                            : 'bg-gradient-to-t from-red-600 to-red-400',
                    )}
                    style={{ height: `${Math.max(c.score, 2)}%` }}
                  />
                </div>
              ))}
            </div>
            <div
              className="mt-2 grid gap-3 text-center text-[10px] uppercase tracking-wider text-muted-foreground"
              style={{ gridTemplateColumns: `repeat(${items.length}, minmax(0, 1fr))` }}
            >
              {items.map((c) => (
                <div key={c.key} className="truncate" title={c.name}>
                  {c.name}
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

/* ─── Shared bits ──────────────────────────────────────────────────────────── */

function SkeletonRows({ rows }: { rows: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="h-3 w-full animate-pulse rounded bg-muted" />
      ))}
    </div>
  )
}

function EmptyState({ text }: { text: string }) {
  return <div className="py-4 text-center text-xs text-muted-foreground">{text}</div>
}

function CardHeader({
  title,
  icon: Icon,
  iconClass,
  action,
}: {
  title: string
  icon?: LucideIcon
  iconClass?: string
  action?: { label: string; href: string }
}) {
  return (
    <div className="flex items-center justify-between px-6 pt-5 pb-4">
      <h3 className="flex items-center gap-2 text-sm font-semibold">
        {Icon && <Icon size={16} strokeWidth={1.75} className={iconClass ?? 'text-muted-foreground'} />}
        {title}
      </h3>
      {action && (
        <Link to={action.href} className="text-xs font-medium text-primary hover:underline">
          {action.label}
        </Link>
      )}
    </div>
  )
}
