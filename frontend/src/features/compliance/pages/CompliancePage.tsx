import { useMemo, useState } from 'react'
import {
  ArrowDownRight,
  ArrowUpRight,
  BadgeCheck,
  Calendar,
  ChevronDown,
  ChevronRight,
  Clock,
  Download,
  ExternalLink,
  FileText,
  ListFilter,
  MoreHorizontal,
  Pause,
  Play,
  Plus,
  Search,
  Sparkles,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Framework = {
  id: string
  name: string
  shortName: string
  score: number
  prevScore: number
  status: 'passing' | 'attention' | 'failing'
  lastEval: string
  controlsTotal: number
  controlsPassing: number
  enabled: boolean
  description: string
}

type Report = {
  id: string
  framework: string
  shortName: string
  score: number
  ts: string
  by: string
  status: 'completed' | 'running' | 'failed'
}

type Schedule = {
  id: string
  framework: string
  shortName: string
  cadence: string
  cron: string
  nextRun: string
  lastRun?: string
  recipients: string[]
  enabled: boolean
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()
const future = (m: number) => new Date(NOW + m * 60_000).toISOString()

const FRAMEWORKS: Framework[] = [
  { id: 'pci', name: 'PCI DSS 4.0', shortName: 'PCI', score: 92, prevScore: 89, status: 'passing', lastEval: hr(8), controlsTotal: 281, controlsPassing: 258, enabled: true, description: 'Payment Card Industry Data Security Standard for cardholder data.' },
  { id: 'hipaa', name: 'HIPAA Security Rule', shortName: 'HIPAA', score: 88, prevScore: 84, status: 'passing', lastEval: hr(8), controlsTotal: 78, controlsPassing: 69, enabled: true, description: 'Health Insurance Portability and Accountability Act for PHI.' },
  { id: 'sox', name: 'SOX (IT Controls)', shortName: 'SOX', score: 76, prevScore: 78, status: 'attention', lastEval: hr(32), controlsTotal: 42, controlsPassing: 32, enabled: true, description: 'Sarbanes-Oxley Act IT general controls.' },
  { id: 'iso27001', name: 'ISO/IEC 27001:2022', shortName: 'ISO 27001', score: 81, prevScore: 80, status: 'passing', lastEval: hr(32), controlsTotal: 93, controlsPassing: 75, enabled: true, description: 'Information security management system standard.' },
  { id: 'nist', name: 'NIST 800-53 Rev 5', shortName: 'NIST 800-53', score: 91, prevScore: 88, status: 'passing', lastEval: days(2), controlsTotal: 414, controlsPassing: 376, enabled: true, description: 'Security and privacy controls for federal information systems.' },
  { id: 'gdpr', name: 'GDPR', shortName: 'GDPR', score: 84, prevScore: 82, status: 'passing', lastEval: days(2), controlsTotal: 36, controlsPassing: 30, enabled: true, description: 'EU General Data Protection Regulation.' },
  { id: 'soc2', name: 'SOC 2 Type II', shortName: 'SOC 2', score: 64, prevScore: 71, status: 'failing', lastEval: days(7), controlsTotal: 64, controlsPassing: 41, enabled: true, description: 'Trust Services Criteria for service organizations.' },
  { id: 'glba', name: 'GLBA Safeguards Rule', shortName: 'GLBA', score: 0, prevScore: 0, status: 'attention', lastEval: '', controlsTotal: 18, controlsPassing: 0, enabled: false, description: 'Gramm-Leach-Bliley Act for financial institutions.' },
]

const REPORTS: Report[] = [
  { id: 'r-2401', framework: 'PCI DSS 4.0', shortName: 'PCI', score: 92, ts: hr(8), by: 'jsmith', status: 'completed' },
  { id: 'r-2400', framework: 'HIPAA Security Rule', shortName: 'HIPAA', score: 88, ts: hr(8), by: 'jsmith', status: 'completed' },
  { id: 'r-2398', framework: 'SOX (IT Controls)', shortName: 'SOX', score: 76, ts: hr(32), by: 'system · scheduled', status: 'completed' },
  { id: 'r-2397', framework: 'ISO/IEC 27001:2022', shortName: 'ISO 27001', score: 81, ts: hr(32), by: 'system · scheduled', status: 'completed' },
  { id: 'r-2393', framework: 'NIST 800-53 Rev 5', shortName: 'NIST 800-53', score: 91, ts: days(2), by: 'mthomas', status: 'completed' },
  { id: 'r-2392', framework: 'GDPR', shortName: 'GDPR', score: 84, ts: days(2), by: 'mthomas', status: 'completed' },
  { id: 'r-2390', framework: 'PCI DSS 4.0', shortName: 'PCI', score: 89, ts: days(7), by: 'system · scheduled', status: 'completed' },
  { id: 'r-2389', framework: 'SOC 2 Type II', shortName: 'SOC 2', score: 64, ts: days(7), by: 'aluna', status: 'completed' },
  { id: 'r-2402', framework: 'PCI DSS 4.0', shortName: 'PCI', score: 0, ts: min(2), by: 'jsmith', status: 'running' },
]

const SCHEDULES: Schedule[] = [
  { id: 's-1', framework: 'PCI DSS 4.0', shortName: 'PCI', cadence: 'Weekly · Monday 06:00', cron: '0 6 * * 1', nextRun: future(60 * 14), lastRun: hr(8), recipients: ['compliance@utmstack.com', 'cfo@utmstack.com'], enabled: true },
  { id: 's-2', framework: 'HIPAA Security Rule', shortName: 'HIPAA', cadence: 'Weekly · Monday 06:00', cron: '0 6 * * 1', nextRun: future(60 * 14), lastRun: hr(8), recipients: ['compliance@utmstack.com'], enabled: true },
  { id: 's-3', framework: 'SOX (IT Controls)', shortName: 'SOX', cadence: 'Monthly · 1st 06:00', cron: '0 6 1 * *', nextRun: future(60 * 24 * 14), lastRun: hr(32), recipients: ['compliance@utmstack.com'], enabled: true },
  { id: 's-4', framework: 'ISO/IEC 27001:2022', shortName: 'ISO 27001', cadence: 'Monthly · 1st 06:00', cron: '0 6 1 * *', nextRun: future(60 * 24 * 14), lastRun: hr(32), recipients: ['compliance@utmstack.com'], enabled: true },
  { id: 's-5', framework: 'GDPR', shortName: 'GDPR', cadence: 'Quarterly', cron: '0 6 1 1,4,7,10 *', nextRun: future(60 * 24 * 60), lastRun: days(2), recipients: ['dpo@utmstack.com'], enabled: true },
  { id: 's-6', framework: 'SOC 2 Type II', shortName: 'SOC 2', cadence: 'Monthly · 1st 06:00', cron: '0 6 1 * *', nextRun: future(60 * 24 * 14), lastRun: days(7), recipients: ['audit@utmstack.com'], enabled: false },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const STATUS_META: Record<Framework['status'], { label: string; tone: string; dot: string }> = {
  passing: { label: 'Passing', tone: 'text-emerald-500', dot: 'bg-emerald-500' },
  attention: { label: 'Needs attention', tone: 'text-amber-500', dot: 'bg-amber-500' },
  failing: { label: 'Failing', tone: 'text-red-500', dot: 'bg-red-500' },
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

type Tab = 'posture' | 'reports' | 'schedule' | 'standards'

export function CompliancePage() {
  const [tab, setTab] = useState<Tab>('posture')

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <Header />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <PostureOverviewCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <ComplianceInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <TabsRow current={tab} onChange={setTab} />
      </div>

      <div className="mt-3">
        {tab === 'posture' && <PostureTab />}
        {tab === 'reports' && <ReportsTab />}
        {tab === 'schedule' && <ScheduleTab />}
        {tab === 'standards' && <StandardsTab />}
      </div>
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header() {
  const [runOpen, setRunOpen] = useState(false)
  const enabled = FRAMEWORKS.filter((f) => f.enabled)
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Compliance</h1>
        <p className="mt-1 text-xs text-muted-foreground">
          Continuously evaluated against {enabled.length} active frameworks.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Download size={14} className="mr-2" />
          Export
        </Button>
        <div className="relative">
          <Button size="sm" onClick={() => setRunOpen((v) => !v)}>
            <Plus size={14} className="mr-2" />
            Run report
            <ChevronDown size={12} className="ml-1.5 opacity-80" />
          </Button>
          {runOpen && (
            <div className="absolute right-0 top-full z-30 mt-1 w-72 overflow-hidden rounded-md border border-border bg-popover shadow-lg">
              <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
                Run a new evaluation
              </div>
              {enabled.map((f) => (
                <button
                  key={f.id}
                  onClick={() => setRunOpen(false)}
                  className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm hover:bg-muted"
                >
                  <span>{f.name}</span>
                  <span className="text-[10px] text-muted-foreground">{f.controlsTotal} controls</span>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </header>
  )
}

/* ─── Posture overview chart ───────────────────────────────────────────── */

function PostureOverviewCard() {
  // Build a 90-day overall compliance trend
  const buckets = 90
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.round(72 + Math.sin(i / 12) * 6 + Math.sin(i / 5) * 2 + (i / buckets) * 8)
      ),
    []
  )
  const overall = Math.round(
    FRAMEWORKS.filter((f) => f.enabled).reduce((s, f) => s + f.score, 0) /
      FRAMEWORKS.filter((f) => f.enabled).length
  )
  const prevOverall = Math.round(
    FRAMEWORKS.filter((f) => f.enabled).reduce((s, f) => s + f.prevScore, 0) /
      FRAMEWORKS.filter((f) => f.enabled).length
  )
  const delta = overall - prevOverall

  const w = 1000
  const h = 100
  const max = 100
  const min0 = Math.min(...data) - 4
  const range = max - min0
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - ((v - min0) / range) * h)
  const linePath = data.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const prevX = xs[i - 1]
    const prevY = ys[i - 1]
    const cx1 = prevX + (xs[i] - prevX) / 2
    const cx2 = xs[i] - (xs[i] - prevX) / 2
    return `${acc} C ${cx1} ${prevY}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Overall posture · last 90 days
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{overall}%</span>
            <span
              className={cn(
                'inline-flex items-center gap-0.5 text-sm',
                delta >= 0 ? 'text-emerald-500' : 'text-red-500'
              )}
            >
              {delta >= 0 ? <ArrowUpRight size={14} /> : <ArrowDownRight size={14} />}
              {Math.abs(delta)} pts
            </span>
            <span className="text-sm text-muted-foreground">vs previous evaluation</span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="postureGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(16 185 129)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(16 185 129)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#postureGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(16 185 129)"
          strokeOpacity="0.85"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>

      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>90d ago</span>
        <span>60d ago</span>
        <span>30d ago</span>
        <span>now</span>
      </div>
    </div>
  )
}

/* ─── Insights ─────────────────────────────────────────────────────────── */

function ComplianceInsightsCard() {
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">SOC AI insights</h3>
      </div>

      <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">SOC 2 dropped 7 pts</span> after 3 access-review
        controls failed. <button className="text-primary hover:underline">View failing controls</button>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">PCI 4.0 ready in ~12 days</span> at the current
        improvement rate (+3 pts/week). <button className="text-primary hover:underline">View gaps</button>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">GLBA not enabled</span> — recommended for
        financial workloads detected in your env. <button className="text-primary hover:underline">Enable</button>
      </div>
    </div>
  )
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

const TABS: { id: Tab; label: string; count?: number }[] = [
  { id: 'posture', label: 'Posture' },
  { id: 'reports', label: 'Reports', count: REPORTS.length },
  { id: 'schedule', label: 'Schedule', count: SCHEDULES.length },
  { id: 'standards', label: 'Standards', count: FRAMEWORKS.length },
]

function TabsRow({ current, onChange }: { current: Tab; onChange: (t: Tab) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const active = current === t.id
        return (
          <button
            key={t.id}
            onClick={() => onChange(t.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{t.label}</span>
            {t.count !== undefined && (
              <span
                className={cn(
                  'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                  active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
                )}
              >
                {t.count}
              </span>
            )}
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Posture tab ──────────────────────────────────────────────────────── */

function PostureTab() {
  return (
    <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
      {FRAMEWORKS.map((f) => (
        <FrameworkCard key={f.id} framework={f} />
      ))}
    </div>
  )
}

function FrameworkCard({ framework: f }: { framework: Framework }) {
  const st = STATUS_META[f.status]
  const delta = f.score - f.prevScore
  if (!f.enabled) {
    return (
      <div className="flex flex-col gap-3 rounded-xl border border-dashed border-border bg-card/40 p-4">
        <div className="flex items-start justify-between">
          <div>
            <div className="text-[10px] uppercase tracking-wider text-muted-foreground">
              {f.shortName}
            </div>
            <h4 className="mt-0.5 text-sm font-medium text-foreground/70">{f.name}</h4>
          </div>
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
            disabled
          </span>
        </div>
        <p className="line-clamp-2 text-xs text-muted-foreground">{f.description}</p>
        <button className="self-start text-[11px] text-primary hover:underline">Enable</button>
      </div>
    )
  }
  return (
    <button className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
            <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
              {f.shortName}
            </span>
          </div>
          <h4 className="mt-0.5 truncate text-sm font-semibold">{f.name}</h4>
        </div>
        <ScoreCircle value={f.score} status={f.status} />
      </div>

      <div className="flex items-center gap-2 text-[11px]">
        <span className={cn('font-medium', st.tone)}>{st.label}</span>
        <span className="text-border">·</span>
        <span
          className={cn(
            'inline-flex items-center gap-0.5',
            delta >= 0 ? 'text-emerald-500' : 'text-red-500'
          )}
        >
          {delta >= 0 ? <ArrowUpRight size={11} /> : <ArrowDownRight size={11} />}
          {Math.abs(delta)} pts
        </span>
      </div>

      <div className="text-[11px] text-muted-foreground">
        <span className="font-mono text-foreground">{f.controlsPassing}</span> of{' '}
        <span className="font-mono">{f.controlsTotal}</span> controls passing
        <div className="mt-1 h-1 overflow-hidden rounded-full bg-muted">
          <div
            className={cn(
              'h-full rounded-full',
              f.status === 'passing' && 'bg-emerald-500',
              f.status === 'attention' && 'bg-amber-500',
              f.status === 'failing' && 'bg-red-500'
            )}
            style={{ width: `${(f.controlsPassing / f.controlsTotal) * 100}%` }}
          />
        </div>
      </div>

      <div className="flex items-center justify-between border-t border-border pt-2.5 text-[11px] text-muted-foreground">
        <span>Last eval {relativeTime(f.lastEval)}</span>
        <ChevronRight size={12} className="opacity-0 transition-opacity group-hover:opacity-100" />
      </div>
    </button>
  )
}

function ScoreCircle({ value, status }: { value: number; status: Framework['status'] }) {
  const tone =
    status === 'passing'
      ? 'stroke-emerald-500'
      : status === 'attention'
        ? 'stroke-amber-500'
        : 'stroke-red-500'
  const r = 16
  const c = 2 * Math.PI * r
  const off = c - (value / 100) * c
  return (
    <div className="relative inline-flex h-11 w-11 items-center justify-center">
      <svg viewBox="0 0 40 40" className="h-11 w-11 -rotate-90">
        <circle cx="20" cy="20" r={r} className="fill-none stroke-muted" strokeWidth="3" />
        <circle
          cx="20"
          cy="20"
          r={r}
          className={cn('fill-none', tone)}
          strokeWidth="3"
          strokeDasharray={c}
          strokeDashoffset={off}
          strokeLinecap="round"
        />
      </svg>
      <span className="absolute inset-0 flex items-center justify-center font-mono text-xs font-semibold tabular-nums">
        {value}
      </span>
    </div>
  )
}

/* ─── Reports tab ──────────────────────────────────────────────────────── */

function ReportsTab() {
  const [search, setSearch] = useState('')
  const filtered = REPORTS.filter((r) =>
    search ? (r.framework + r.id + r.by).toLowerCase().includes(search.toLowerCase()) : true
  )

  return (
    <div>
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative min-w-[280px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by framework, ID, or generated by…"
            className="h-9 pl-9"
          />
        </div>
        <Button variant="outline" size="sm">
          <ListFilter size={14} className="mr-2" />
          Filters
        </Button>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: '90px 1fr 100px 80px 160px 110px 36px' }}
        >
          <div>ID</div>
          <div>Framework</div>
          <div>Status</div>
          <div className="text-right">Score</div>
          <div>Generated by</div>
          <div>When</div>
          <div />
        </div>
        {filtered.map((r) => (
          <ReportRow key={r.id} report={r} />
        ))}
        {filtered.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">No reports match.</div>
        )}
      </div>
    </div>
  )
}

function ReportRow({ report: r }: { report: Report }) {
  return (
    <div
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: '90px 1fr 100px 80px 160px 110px 36px' }}
    >
      <div className="font-mono text-muted-foreground">{r.id}</div>
      <div className="min-w-0">
        <div className="truncate font-medium">{r.framework}</div>
        <div className="text-[10px] text-muted-foreground">{r.shortName}</div>
      </div>
      <div className="flex items-center gap-1.5 text-[11px]">
        {r.status === 'running' && (
          <>
            <span className="h-1.5 w-1.5 rounded-full bg-sky-500 animate-pulse" />
            <span className="text-sky-500">Running</span>
          </>
        )}
        {r.status === 'completed' && (
          <>
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
            <span className="text-muted-foreground">Done</span>
          </>
        )}
        {r.status === 'failed' && (
          <>
            <span className="h-1.5 w-1.5 rounded-full bg-red-500" />
            <span className="text-red-500">Failed</span>
          </>
        )}
      </div>
      <div className="text-right">
        {r.status === 'completed' ? (
          <span className="font-mono tabular-nums">{r.score}%</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </div>
      <div className="truncate text-muted-foreground">
        <span className="font-mono">{r.by}</span>
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(r.ts)}</div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground">
          <ExternalLink size={12} />
        </button>
      </div>
    </div>
  )
}

/* ─── Schedule tab ─────────────────────────────────────────────────────── */

function ScheduleTab() {
  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Recurring evaluations. Reports are emailed to the configured recipients on each run.
        </p>
        <Button size="sm">
          <Plus size={13} className="mr-1.5" />
          New schedule
        </Button>
      </div>
      <div className="overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: '12px 1fr 200px 1fr 130px 110px 36px' }}
        >
          <div />
          <div>Framework</div>
          <div>Cadence</div>
          <div>Recipients</div>
          <div>Next run</div>
          <div>Last run</div>
          <div />
        </div>
        {SCHEDULES.map((s) => (
          <ScheduleRow key={s.id} schedule={s} />
        ))}
      </div>
    </div>
  )
}

function ScheduleRow({ schedule: s }: { schedule: Schedule }) {
  return (
    <div
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: '12px 1fr 200px 1fr 130px 110px 36px' }}
    >
      <span
        className={cn(
          'h-2 w-2 rounded-full',
          s.enabled ? 'bg-emerald-500' : 'bg-muted-foreground'
        )}
      />
      <div className="min-w-0">
        <div className="truncate font-medium">{s.framework}</div>
        <div className="font-mono text-[10px] text-muted-foreground">{s.shortName}</div>
      </div>
      <div className="text-[11px]">
        <div>{s.cadence}</div>
        <div className="font-mono text-[10px] text-muted-foreground">{s.cron}</div>
      </div>
      <div className="min-w-0 text-[11px] text-muted-foreground">
        {s.recipients.slice(0, 1).map((r) => (
          <div key={r} className="truncate font-mono">{r}</div>
        ))}
        {s.recipients.length > 1 && (
          <div className="text-[10px] text-muted-foreground/70">
            +{s.recipients.length - 1} more
          </div>
        )}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {s.enabled ? relativeFuture(s.nextRun) : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {s.lastRun ? relativeTime(s.lastRun) : '—'}
      </div>
      <div className="flex justify-end gap-0.5 opacity-0 group-hover:opacity-100">
        <button
          title={s.enabled ? 'Pause' : 'Resume'}
          className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-background hover:text-foreground"
        >
          {s.enabled ? <Pause size={11} /> : <Play size={11} />}
        </button>
        <button className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-background hover:text-foreground">
          <MoreHorizontal size={12} />
        </button>
      </div>
    </div>
  )
}

/* ─── Standards tab ────────────────────────────────────────────────────── */

function StandardsTab() {
  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Enable or disable frameworks, edit custom controls, and configure scope.
        </p>
        <Button size="sm" variant="outline">
          <Plus size={13} className="mr-1.5" />
          Add custom standard
        </Button>
      </div>
      <div className="overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: '60px 1fr 100px 110px 110px 36px' }}
        >
          <div>Status</div>
          <div>Framework</div>
          <div className="text-right">Controls</div>
          <div className="text-right">Score</div>
          <div>Last eval</div>
          <div />
        </div>
        {FRAMEWORKS.map((f) => (
          <StandardRow key={f.id} framework={f} />
        ))}
      </div>
    </div>
  )
}

function StandardRow({ framework: f }: { framework: Framework }) {
  return (
    <div
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: '60px 1fr 100px 110px 110px 36px' }}
    >
      <Toggle enabled={f.enabled} />
      <div className="min-w-0">
        <div className="truncate font-medium">{f.name}</div>
        <div className="truncate text-[10px] text-muted-foreground">{f.description}</div>
      </div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">
        {f.controlsTotal}
      </div>
      <div className="text-right">
        {f.enabled && f.lastEval ? (
          <span className="font-mono tabular-nums">{f.score}%</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {f.lastEval ? relativeTime(f.lastEval) : '—'}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-background hover:text-foreground">
          <MoreHorizontal size={12} />
        </button>
      </div>
    </div>
  )
}

function Toggle({ enabled }: { enabled: boolean }) {
  return (
    <button
      onClick={(e) => e.stopPropagation()}
      className={cn(
        'relative h-4 w-7 rounded-full transition-colors',
        enabled ? 'bg-primary' : 'bg-muted'
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-3 w-3 rounded-full bg-white shadow transition-all',
          enabled ? 'left-3.5' : 'left-0.5'
        )}
      />
    </button>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  if (!iso) return '—'
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.round(h / 24)}d ago`
}

function relativeFuture(iso: string) {
  if (!iso) return '—'
  const diff = new Date(iso).getTime() - NOW
  const m = Math.round(diff / 60_000)
  if (m < 60) return `in ${m}m`
  const h = Math.round(m / 60)
  if (h < 24) return `in ${h}h`
  return `in ${Math.round(h / 24)}d`
}

// Suppress unused-import warnings
void BadgeCheck
void Calendar
void Clock
void FileText
void X
void ChevronRight as unknown as LucideIcon
