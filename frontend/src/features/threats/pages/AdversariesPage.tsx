import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  Activity,
  AlertTriangle,
  Crosshair,
  Globe2,
  LayoutGrid,
  ListIcon,
  Loader2,
  RefreshCw,
  Search,
  Server,
  ShieldAlert,
  User,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { adversariesHttpService } from '../services/adversaries-http.service'
import { toAdversaries } from '../lib/adversary'
import type { Adversary, AdversaryKind, ThreatLevel } from '../types/adversary.types'

const THREAT: Record<ThreatLevel, { dot: string; text: string; ring: string }> = {
  critical: { dot: 'bg-red-500', text: 'text-red-600 dark:text-red-400', ring: 'ring-red-500/40' },
  high: { dot: 'bg-orange-500', text: 'text-orange-600 dark:text-orange-400', ring: 'ring-orange-500/40' },
  medium: { dot: 'bg-amber-500', text: 'text-amber-600 dark:text-amber-400', ring: 'ring-amber-500/40' },
  low: { dot: 'bg-sky-500', text: 'text-sky-600 dark:text-sky-400', ring: 'ring-sky-500/40' },
}

const KIND_ICON: Record<AdversaryKind, LucideIcon> = { ip: Globe2, host: Server, user: User }

type ViewId = 'all' | 'high' | 'ip' | 'host' | 'user'
const VIEWS: { id: ViewId; predicate: (a: Adversary) => boolean }[] = [
  { id: 'all', predicate: () => true },
  { id: 'high', predicate: (a) => a.maxSeverity >= 3 },
  { id: 'ip', predicate: (a) => a.kind === 'ip' },
  { id: 'host', predicate: (a) => a.kind === 'host' },
  { id: 'user', predicate: (a) => a.kind === 'user' },
]

export function AdversariesPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<ViewId>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'list' | 'cards'>('list')
  const [adversaries, setAdversaries] = useState<Adversary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [open, setOpen] = useState<Adversary | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const res = await adversariesHttpService.list([])
      setAdversaries(toAdversaries(res ?? []))
    } catch {
      setError(true)
      setAdversaries([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const counts = useMemo(
    () => Object.fromEntries(VIEWS.map((v) => [v.id, adversaries.filter(v.predicate).length])),
    [adversaries],
  )

  const filtered = useMemo(() => {
    const v = VIEWS.find((x) => x.id === view) ?? VIEWS[0]
    const q = search.trim().toLowerCase()
    return adversaries.filter(v.predicate).filter((a) =>
      q
        ? (a.identifier + ' ' + (a.geo?.country ?? '') + ' ' + (a.geo?.aso ?? '')).toLowerCase().includes(q)
        : true,
    )
  }, [adversaries, view, search])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="flex items-center gap-2 text-xl font-semibold">
              <ShieldAlert size={18} strokeWidth={1.75} />
              {t('adversaries.title')}
            </h1>
            <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
              {t('adversaries.countChip', { count: adversaries.length })}
            </span>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">{t('adversaries.subtitle')}</p>
        </div>
        <button
          onClick={() => void load()}
          title={t('adversaries.refresh')}
          className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </button>
      </header>

      <div className="mt-5">
        <OverviewCard adversaries={adversaries} t={t} />
      </div>

      <div className="mt-5 flex flex-wrap items-center gap-1 border-b border-border">
        {VIEWS.map((v) => {
          const active = view === v.id
          return (
            <button
              key={v.id}
              onClick={() => setView(v.id)}
              className={cn(
                'relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
                active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              <span>{t(`adversaries.views.${v.id}`)}</span>
              <span
                className={cn(
                  'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                  active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground',
                )}
              >
                {counts[v.id] ?? 0}
              </span>
              {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          )
        })}
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('adversaries.search')} className="h-9 pl-9" />
        </div>
        <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
          <button onClick={() => setLayout('list')} title={t('adversaries.layout.list')} className={cn('flex h-7 w-7 items-center justify-center rounded transition-colors', layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>
            <ListIcon size={13} />
          </button>
          <button onClick={() => setLayout('cards')} title={t('adversaries.layout.cards')} className={cn('flex h-7 w-7 items-center justify-center rounded transition-colors', layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>
            <LayoutGrid size={13} />
          </button>
        </div>
      </div>

      {error ? (
        <div className="mt-3 flex flex-col items-center gap-3 rounded-xl border border-border bg-card px-6 py-12 text-sm">
          <span className="inline-flex items-center gap-2 text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" />
            {t('adversaries.loadError')}
          </span>
          <Button variant="outline" size="sm" onClick={() => void load()}>
            {t('adversaries.retry')}
          </Button>
        </div>
      ) : loading ? (
        <div className="mt-3 flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="mt-3 rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
          {t('adversaries.empty')}
        </div>
      ) : layout === 'list' ? (
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
          <ListHeader t={t} />
          {filtered.map((a) => (
            <Row key={a.id} a={a} t={t} onOpen={() => setOpen(a)} />
          ))}
        </div>
      ) : (
        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((a) => (
            <Card key={a.id} a={a} t={t} onOpen={() => setOpen(a)} />
          ))}
        </div>
      )}

      {open && <Drawer a={open} t={t} onClose={() => setOpen(null)} />}
    </div>
  )
}

/* ─── Overview ─────────────────────────────────────────────────────────── */

function OverviewCard({ adversaries, t }: { adversaries: Adversary[]; t: TFunction }) {
  const summed = useMemo(() => {
    const b = new Array(24).fill(0)
    for (const a of adversaries) for (let i = 0; i < 24; i++) b[i] += a.activity[i] ?? 0
    return b
  }, [adversaries])
  const high = adversaries.filter((a) => a.maxSeverity >= 3).length
  const countries = new Set(adversaries.map((a) => a.geo?.country).filter(Boolean)).size
  const totalAlerts = adversaries.reduce((s, a) => s + a.alertsCount, 0)

  const w = 1000
  const h = 100
  const max = Math.max(...summed, 1) * 1.15
  const xs = summed.map((_, i) => (i * w) / (summed.length - 1))
  const ys = summed.map((v) => h - (v / max) * h)
  const line = summed.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const cx1 = xs[i - 1] + (xs[i] - xs[i - 1]) / 2
    const cx2 = xs[i] - (xs[i] - xs[i - 1]) / 2
    return `${acc} C ${cx1} ${ys[i - 1]}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const area = `${line} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('adversaries.overview.title')}</div>
      <div className="mt-1 flex items-baseline gap-3">
        <span className="text-3xl font-semibold tabular-nums">{adversaries.length}</span>
        <span className="text-sm text-muted-foreground">
          <span className="text-foreground">{high}</span> {t('adversaries.overview.high')}
          <span className="mx-2 text-border">·</span>
          <span className="text-foreground">{countries}</span> {t('adversaries.overview.countries')}
          <span className="mx-2 text-border">·</span>
          <span className="text-foreground">{totalAlerts}</span> {t('adversaries.overview.alerts')}
        </span>
      </div>
      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="advGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(239 68 68)" stopOpacity="0.25" />
            <stop offset="100%" stopColor="rgb(239 68 68)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={area} fill="url(#advGrad)" />
        <path d={line} fill="none" stroke="rgb(239 68 68)" strokeOpacity="0.8" strokeWidth="1.75" strokeLinejoin="round" strokeLinecap="round" />
      </svg>
      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>{t('adversaries.overview.ago24')}</span>
        <span>{t('adversaries.overview.ago12')}</span>
        <span>{t('adversaries.overview.now')}</span>
      </div>
    </div>
  )
}

/* ─── List + Card ──────────────────────────────────────────────────────── */

const COLS = '32px 1fr 100px 80px 90px 110px'

function ListHeader({ t }: { t: TFunction }) {
  return (
    <div className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
      <div />
      <div>{t('adversaries.col.adversary')}</div>
      <div>{t('adversaries.col.threat')}</div>
      <div className="text-right">{t('adversaries.col.alerts')}</div>
      <div className="text-right">{t('adversaries.col.targets')}</div>
      <div>{t('adversaries.col.lastSeen')}</div>
    </div>
  )
}

function Avatar({ a, large }: { a: Adversary; large?: boolean }) {
  const Icon = KIND_ICON[a.kind]
  return (
    <div className={cn('flex shrink-0 items-center justify-center rounded-lg bg-muted ring-2', large ? 'h-11 w-11' : 'h-7 w-7', THREAT[a.threatLevel].ring)}>
      <Icon size={large ? 18 : 13} className={THREAT[a.threatLevel].text} />
    </div>
  )
}

function Row({ a, t, onOpen }: { a: Adversary; t: TFunction; onOpen: () => void }) {
  return (
    <div onClick={onOpen} className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/40" style={{ gridTemplateColumns: COLS }}>
      <Avatar a={a} />
      <div className="min-w-0">
        <div className="truncate font-mono text-[13px] font-medium">{a.identifier}</div>
        <div className="truncate text-[11px] text-muted-foreground">
          {a.geo?.country ? `${flagEmoji(a.geo.countryCode)} ${a.geo.country}` : t(`adversaries.kind.${a.kind}`)}
          {a.geo?.aso ? ` · ${a.geo.aso}` : ''}
        </div>
      </div>
      <div>
        <span className={cn('inline-flex items-center gap-1.5 text-[11px]', THREAT[a.threatLevel].text)}>
          <span className={cn('h-1.5 w-1.5 rounded-full', THREAT[a.threatLevel].dot)} />
          {t(`adversaries.threat.${a.threatLevel}`)}
        </span>
      </div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">{a.alertsCount}</div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">{a.targets.length}</div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(a.lastSeen, t)}</div>
    </div>
  )
}

function Card({ a, t, onOpen }: { a: Adversary; t: TFunction; onOpen: () => void }) {
  return (
    <button onClick={onOpen} className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md">
      <div className="flex items-start gap-3">
        <Avatar a={a} large />
        <div className="min-w-0 flex-1">
          <div className="truncate font-mono text-sm font-semibold">{a.identifier}</div>
          <div className="truncate text-[11px] text-muted-foreground">
            {a.geo?.country ? `${flagEmoji(a.geo.countryCode)} ${a.geo.country}` : t(`adversaries.kind.${a.kind}`)}
          </div>
          <div className="mt-1">
            <span className={cn('inline-flex items-center gap-1.5 text-[11px]', THREAT[a.threatLevel].text)}>
              <span className={cn('h-1.5 w-1.5 rounded-full', THREAT[a.threatLevel].dot)} />
              {t(`adversaries.threat.${a.threatLevel}`)}
            </span>
          </div>
        </div>
      </div>
      <div className="grid grid-cols-3 gap-2 border-t border-border pt-3 text-[11px]">
        <Stat label={t('adversaries.card.alerts')} value={a.alertsCount} />
        <Stat label={t('adversaries.card.techniques')} value={a.techniques.length} />
        <Stat label={t('adversaries.card.targets')} value={a.targets.length} />
      </div>
      <div className="text-[11px] text-muted-foreground">
        {t('adversaries.col.lastSeen')} <span className="font-mono text-foreground">{relativeTime(a.lastSeen, t)}</span>
      </div>
    </button>
  )
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-2 py-1.5">
      <div className="text-[9px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-0.5 font-semibold tabular-nums">{value}</div>
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'profile' | 'techniques' | 'targets' | 'alerts'

function Drawer({ a, t, onClose }: { a: Adversary; t: TFunction; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('profile')
  const maxTech = Math.max(...a.techniques.map((x) => x.count), 1)
  const maxTgt = Math.max(...a.targets.map((x) => x.hits), 1)

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <Avatar a={a} large />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{t(`adversaries.kind.${a.kind}`)}</span>
                  {a.geo?.country && (
                    <>
                      <span>·</span>
                      <span>
                        {flagEmoji(a.geo.countryCode)} {a.geo.country}
                      </span>
                    </>
                  )}
                </div>
                <h2 className="mt-0.5 truncate font-mono text-xl font-semibold">{a.identifier}</h2>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', THREAT[a.threatLevel].text, THREAT[a.threatLevel].ring)}>
                    <span className={cn('h-1.5 w-1.5 rounded-full', THREAT[a.threatLevel].dot)} />
                    {t(`adversaries.threat.${a.threatLevel}`)}
                  </span>
                  <span className="text-muted-foreground">
                    {t('adversaries.card.alerts')}: <span className="font-mono text-foreground">{a.alertsCount}</span>
                  </span>
                </div>
              </div>
            </div>
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
              <X size={16} />
            </button>
          </div>
          <nav className="mt-4 flex items-center gap-1">
            {(['profile', 'techniques', 'targets', 'alerts'] as Tab[]).map((id) => (
              <DrawerTab key={id} id={id} current={tab} onChange={setTab} t={t} />
            ))}
          </nav>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto bg-muted/20 p-6">
          {tab === 'profile' && (
            <>
              <Section title={t('adversaries.drawer.activity')}>
                <ActivityChart data={a.activity} />
              </Section>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <Section title={t('adversaries.drawer.identification')}>
                  <dl className="grid grid-cols-[110px_1fr] gap-y-2 text-xs">
                    <Row2 k={t('adversaries.drawer.kindLabel')}>{t(`adversaries.kind.${a.kind}`)}</Row2>
                    <Row2 k={t('adversaries.drawer.identifier')}>
                      <span className="break-all font-mono">{a.identifier}</span>
                    </Row2>
                    <Row2 k={t('adversaries.drawer.firstSeen')}>{relativeTime(a.firstSeen, t)}</Row2>
                    <Row2 k={t('adversaries.drawer.lastSeen')}>{relativeTime(a.lastSeen, t)}</Row2>
                  </dl>
                </Section>
                {a.geo && (a.geo.country || a.geo.aso) && (
                  <Section title={t('adversaries.drawer.geolocation')}>
                    <dl className="grid grid-cols-[110px_1fr] gap-y-2 text-xs">
                      {a.geo.country && (
                        <Row2 k={t('adversaries.drawer.country')}>
                          {flagEmoji(a.geo.countryCode)} {a.geo.country}
                        </Row2>
                      )}
                      {a.geo.city && <Row2 k={t('adversaries.drawer.city')}>{a.geo.city}</Row2>}
                      {a.geo.asn ? (
                        <Row2 k="ASN">
                          <span className="font-mono">{a.geo.asn}</span>
                        </Row2>
                      ) : null}
                      {a.geo.aso && <Row2 k="ASO">{a.geo.aso}</Row2>}
                    </dl>
                  </Section>
                )}
              </div>
            </>
          )}

          {tab === 'techniques' && (
            <Section title={t('adversaries.drawer.topTechniques')}>
              {a.techniques.length === 0 ? (
                <div className="text-xs text-muted-foreground">{t('adversaries.drawer.noTechniques')}</div>
              ) : (
                <ul className="space-y-2">
                  {a.techniques.map((tch) => (
                    <li key={tch.id} className="flex items-center gap-3 text-xs">
                      <span className="w-48 shrink-0 truncate font-mono text-muted-foreground" title={tch.id}>
                        {tch.id}
                      </span>
                      <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                        <div className="h-full rounded-full bg-orange-500/70" style={{ width: `${(tch.count / maxTech) * 100}%` }} />
                      </div>
                      <span className="w-10 text-right font-mono tabular-nums text-muted-foreground">{tch.count}</span>
                    </li>
                  ))}
                </ul>
              )}
            </Section>
          )}

          {tab === 'targets' && (
            <Section title={t('adversaries.drawer.targets')}>
              {a.targets.length === 0 ? (
                <div className="text-xs text-muted-foreground">{t('adversaries.drawer.noTargets')}</div>
              ) : (
                <ul className="space-y-2">
                  {a.targets.map((tg) => {
                    const Icon = KIND_ICON[tg.kind]
                    return (
                      <li key={tg.label} className="flex items-center gap-3 text-xs">
                        <Icon size={13} className="shrink-0 text-muted-foreground" />
                        <span className="w-44 shrink-0 truncate font-mono">{tg.label}</span>
                        <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                          <div className="h-full rounded-full bg-sky-500/70" style={{ width: `${(tg.hits / maxTgt) * 100}%` }} />
                        </div>
                        <span className="w-10 text-right font-mono tabular-nums text-muted-foreground">{tg.hits}</span>
                      </li>
                    )
                  })}
                </ul>
              )}
            </Section>
          )}

          {tab === 'alerts' && (
            <Section title={t('adversaries.drawer.recentAlerts')}>
              {a.alerts.length === 0 ? (
                <div className="text-xs text-muted-foreground">{t('adversaries.drawer.noAlerts')}</div>
              ) : (
                <ul className="overflow-hidden rounded-lg border border-border">
                  {a.alerts.map((al) => (
                    <li key={al.id} className="flex items-center gap-3 border-b border-border/60 px-3 py-2 text-xs last:border-b-0">
                      <span className={cn('h-3 w-1 shrink-0 rounded-full', THREAT[threatFromSeverity(al.severity)].dot)} />
                      <div className="min-w-0 flex-1">
                        <div className="truncate font-medium">{al.name}</div>
                        <div className="mt-0.5 flex items-center gap-2 text-[11px] text-muted-foreground">
                          {al.technique && <span className="font-mono">{al.technique}</span>}
                          {al.target && <span>→ {al.target.label}</span>}
                          {al.echoes > 0 && <span>· {t('adversaries.drawer.echoes', { count: al.echoes })}</span>}
                        </div>
                      </div>
                      <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{relativeTime(al.timestamp, t)}</span>
                    </li>
                  ))}
                </ul>
              )}
            </Section>
          )}
        </div>
      </div>
    </div>
  )
}

function DrawerTab({ id, current, onChange, t }: { id: Tab; current: Tab; onChange: (t: Tab) => void; t: TFunction }) {
  const active = id === current
  const ICONS: Record<Tab, LucideIcon> = { profile: User, techniques: Activity, targets: Crosshair, alerts: AlertTriangle }
  const Icon = ICONS[id]
  return (
    <button onClick={() => onChange(id)} className={cn('relative flex items-center gap-1.5 px-3 py-2 text-xs transition-colors', active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}>
      <Icon size={13} />
      {t(`adversaries.tabs.${id}`)}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

function ActivityChart({ data }: { data: number[] }) {
  const w = 600
  const h = 90
  const max = Math.max(...data, 1)
  const bw = w / data.length
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-20 w-full">
      {data.map((v, i) => {
        const bh = (v / max) * (h - 14)
        return <rect key={i} x={i * bw + 1} y={h - bh - 10} width={bw - 2} height={bh} rx={1} className="fill-red-500/60" />
      })}
      <g className="fill-muted-foreground" fontSize="9">
        {[0, 6, 12, 18, 23].map((i) => (
          <text key={i} x={i * bw + bw / 2} y={h - 1} textAnchor="middle">{`${String(i).padStart(2, '0')}h`}</text>
        ))}
      </g>
    </svg>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}
function Row2({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function threatFromSeverity(sev: number): ThreatLevel {
  if (sev >= 4) return 'critical'
  if (sev === 3) return 'high'
  if (sev === 2) return 'medium'
  return 'low'
}

function relativeTime(iso: string | undefined, t: TFunction): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (Number.isNaN(diff)) return '—'
  const m = Math.round(diff / 60_000)
  if (m < 1) return t('adversaries.relative.justNow')
  if (m < 60) return t('adversaries.relative.minutesAgo', { count: m })
  const h = Math.round(m / 60)
  if (h < 24) return t('adversaries.relative.hoursAgo', { count: h })
  return t('adversaries.relative.daysAgo', { count: Math.round(h / 24) })
}

function flagEmoji(cc?: string): string {
  if (!cc || cc.length !== 2) return ''
  const A = 0x1f1e6
  return String.fromCodePoint(A - 65 + cc.toUpperCase().charCodeAt(0), A - 65 + cc.toUpperCase().charCodeAt(1))
}
