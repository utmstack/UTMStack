import {
  Activity,
  AlertTriangle,
  ArrowDownRight,
  ArrowUpRight,
  AtSign,
  BadgeCheck,
  Crosshair,
  Database,
  Flame,
  LockKeyhole,
  Paperclip,
  Radar,
  Server,
  Sparkles,
  TrendingUp,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'

export function HomePage() {
  return (
    <div className="mx-auto w-full max-w-[1400px] px-6 py-10">
      <ChatHero />

      <div className="mt-10 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-6">
          <DatasourcesCard />
        </div>
        <div className="col-span-12 lg:col-span-6">
          <LiveCasesCard />
        </div>

        <div className="col-span-6 md:col-span-3 lg:col-span-3">
          <KpiTile icon={AlertTriangle} label="Alerts (24h)" value="1,284" trend="+12%" trendUp data={[12, 18, 14, 22, 17, 25, 28]} accent="text-red-500" />
        </div>
        <div className="col-span-6 md:col-span-3 lg:col-span-3">
          <KpiTile icon={Flame} label="Open incidents" value="47" trend="-8%" trendUp={false} data={[55, 52, 48, 50, 47, 45, 47]} accent="text-amber-500" />
        </div>
        <div className="col-span-6 md:col-span-3 lg:col-span-3">
          <KpiTile icon={LockKeyhole} label="Failed logins (24h)" value="312" trend="+24%" trendUp data={[120, 140, 180, 200, 240, 280, 312]} accent="text-sky-500" />
        </div>
        <div className="col-span-6 md:col-span-3 lg:col-span-3">
          <KpiTile icon={Workflow} label="Active playbooks" value="9" trend="+2" trendUp data={[6, 7, 7, 8, 8, 9, 9]} accent="text-violet-500" />
        </div>

        <div className="col-span-12 lg:col-span-8">
          <MitreTechniquesCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <SystemHealthCard />
        </div>

        <div className="col-span-12 lg:col-span-6">
          <TopAssetsCard />
        </div>
        <div className="col-span-12 lg:col-span-6">
          <ComplianceCard />
        </div>

        <div className="col-span-12">
          <ThreatIntelCard />
        </div>
      </div>
    </div>
  )
}

const SUGGESTIONS = ['Threat hunt', 'Observable triage', 'Write a rule', 'Current system status']

function ChatHero() {
  return (
    <section className="flex flex-col items-center">
      <h2 className="mb-4 text-center text-base font-medium text-foreground">
        What would you like to know from your security data?
      </h2>
      <div className="relative w-full max-w-3xl">
        <div
          aria-hidden
          className="absolute -inset-px rounded-xl bg-[conic-gradient(from_180deg_at_50%_50%,#a855f7_0deg,#ec4899_120deg,#3b82f6_240deg,#a855f7_360deg)] opacity-60 blur-[1px]"
        />
        <div className="relative rounded-xl bg-card p-4">
          <textarea
            rows={3}
            placeholder="Research Scattered Spider, does it apply to us, have we ever been hit, build a detection rule and keep it tuned."
            className="w-full resize-none border-0 bg-transparent text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
          />
          <div className="mt-2 flex items-center justify-between">
            <div className="flex items-center gap-2 text-muted-foreground">
              <IconBtn label="Mention"><AtSign size={16} strokeWidth={1.75} /></IconBtn>
              <IconBtn label="Attach"><Paperclip size={16} strokeWidth={1.75} /></IconBtn>
            </div>
            <button
              className="flex h-7 w-7 items-center justify-center rounded-md bg-gradient-to-br from-fuchsia-500 to-purple-600 text-white shadow-sm hover:opacity-90"
              aria-label="Send"
            >
              <Sparkles size={14} strokeWidth={2} />
            </button>
          </div>
        </div>
      </div>
      <div className="mt-4 flex flex-wrap justify-center gap-2">
        {SUGGESTIONS.map((s) => (
          <button
            key={s}
            className="rounded-full border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            {s}
          </button>
        ))}
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

/* ─── KPI tiles ────────────────────────────────────────────────────────── */

function KpiTile({
  icon: Icon,
  label,
  value,
  trend,
  trendUp,
  data,
  accent,
}: {
  icon: LucideIcon
  label: string
  value: string
  trend: string
  trendUp: boolean
  data: number[]
  accent: string
}) {
  return (
    <div className="flex h-full flex-col rounded-xl border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Icon size={14} strokeWidth={1.75} className={accent} />
          {label}
        </div>
        <div
          className={cn(
            'flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium',
            trendUp
              ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'
              : 'bg-red-500/15 text-red-500 dark:text-red-400'
          )}
        >
          {trendUp ? <ArrowUpRight size={10} /> : <ArrowDownRight size={10} />}
          {trend}
        </div>
      </div>
      <div className="mt-2 text-3xl font-semibold tracking-tight">{value}</div>
      <div className="mt-3 -mx-1">
        <Sparkline data={data} accent={accent} />
      </div>
    </div>
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

/* ─── Datasources ──────────────────────────────────────────────────────── */

function DatasourcesCard() {
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title="Datasources" icon={Database} action={{ label: 'View all', href: '/datasources' }} />
      <div className="flex flex-wrap items-end gap-x-10 gap-y-4 px-6 pt-2">
        <div>
          <div className="text-5xl font-semibold leading-none tracking-tight text-emerald-500 dark:text-emerald-400">
            47
          </div>
          <div className="mt-2 text-xs text-muted-foreground">
            <span className="text-foreground">of 47</span> healthy
            <br />
            datasources
          </div>
        </div>
        <div>
          <div className="flex items-baseline gap-2">
            <span className="text-5xl font-semibold leading-none tracking-tight">79.4 M</span>
            <TrendingUp size={20} strokeWidth={2} className="text-emerald-500 dark:text-emerald-400" />
          </div>
          <div className="mt-2 text-xs text-muted-foreground">
            Total events
            <br />
            last 7 days
          </div>
        </div>
      </div>
      <div className="px-2 pb-2 pt-4">
        <EventsChart />
      </div>
    </div>
  )
}

function EventsChart() {
  const days = ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN']
  const points = [9, 14, 8, 18, 6, 13, 11]
  const max = 22
  const w = 700
  const h = 180
  const pl = 36
  const pr = 16
  const pt = 10
  const pb = 24
  const innerW = w - pl - pr
  const innerH = h - pt - pb
  const xs = points.map((_, i) => pl + (i * innerW) / (points.length - 1))
  const ys = points.map((v) => pt + innerH - (v / max) * innerH)
  const linePath = points.map((_, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys[i]}`).join(' ')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${pt + innerH} L ${xs[0]} ${pt + innerH} Z`
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-44 w-full" preserveAspectRatio="none">
      <defs>
        <pattern id="hatch" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <line x1="0" y1="0" x2="0" y2="6" className="stroke-foreground/5" strokeWidth="1" />
        </pattern>
        <linearGradient id="lineGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stopColor="#22d3ee" />
          <stop offset="1" stopColor="#a855f7" />
        </linearGradient>
      </defs>
      <rect x={pl} y={pt} width={innerW} height={innerH} fill="url(#hatch)" />
      <g className="fill-muted-foreground" fontSize="10">
        <text x={pl - 6} y={pt + innerH - (10 / max) * innerH + 3} textAnchor="end">10M</text>
        <text x={pl - 6} y={pt + innerH - (20 / max) * innerH + 3} textAnchor="end">20M</text>
      </g>
      {[10, 20].map((v) => (
        <line key={v} x1={pl} x2={pl + innerW} y1={pt + innerH - (v / max) * innerH} y2={pt + innerH - (v / max) * innerH} className="stroke-border/40" strokeDasharray="2 4" />
      ))}
      <path d={areaPath} className="fill-primary/10" />
      <path d={linePath} fill="none" stroke="url(#lineGrad)" strokeWidth="2" strokeLinejoin="round" />
      {points.map((_, i) => (
        <circle key={i} cx={xs[i]} cy={ys[i]} r="3.5" className="fill-card stroke-primary" strokeWidth="2" />
      ))}
      <g className="fill-muted-foreground" fontSize="10">
        {days.map((d, i) => (
          <text key={d} x={xs[i]} y={h - 6} textAnchor="middle">{d}</text>
        ))}
      </g>
    </svg>
  )
}

/* ─── Live Agent cases ─────────────────────────────────────────────────── */

type CaseStatus = 'Completed' | 'Investigating' | 'On Hold' | 'Queued'
type Conclusion = 'Benign' | 'Malicious' | 'Needs Input' | null

interface CaseRow {
  id: number
  status: CaseStatus
  title: string
  conclusion: Conclusion
}

const CASES: CaseRow[] = [
  { id: 1, status: 'Completed', title: 'Possible Lateral Movement via PsExec from Unusual Host', conclusion: 'Benign' },
  { id: 2, status: 'Completed', title: 'Routine False Positive scan on Rule ID 992', conclusion: 'Benign' },
  { id: 3, status: 'Completed', title: 'Link related cases #488 and #491', conclusion: 'Malicious' },
  { id: 4, status: 'Investigating', title: 'Researching new threat intel related to CVE-2024-38077', conclusion: null },
  { id: 5, status: 'Investigating', title: 'Investigating log gap from Firewall-Z', conclusion: null },
  { id: 6, status: 'On Hold', title: 'Verify a suspicious PowerShell execution on Host-A', conclusion: 'Needs Input' },
  { id: 7, status: 'Queued', title: 'Routine False Positive scan on Rule ID 992', conclusion: null },
  { id: 8, status: 'Queued', title: 'WAF rate limiting surge: IP 70.39.165.194 recently having excessive…', conclusion: null },
  { id: 9, status: 'Queued', title: 'Multi-technique threat campaign: 2 MITRE techniques observed…', conclusion: null },
]

const STATUS_STYLES: Record<CaseStatus, string> = {
  Completed: 'bg-violet-500/15 text-violet-500 dark:text-violet-300 ring-1 ring-inset ring-violet-500/30',
  Investigating: 'bg-sky-500/15 text-sky-500 dark:text-sky-300 ring-1 ring-inset ring-sky-500/30',
  'On Hold': 'bg-amber-500/15 text-amber-600 dark:text-amber-300 ring-1 ring-inset ring-amber-500/30',
  Queued: 'bg-muted text-muted-foreground ring-1 ring-inset ring-border',
}

const CONCLUSION_STYLES: Record<Exclude<Conclusion, null>, string> = {
  Benign: 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300 ring-1 ring-inset ring-emerald-500/30',
  Malicious: 'bg-red-500/15 text-red-500 dark:text-red-300 ring-1 ring-inset ring-red-500/30',
  'Needs Input': 'bg-amber-500/15 text-amber-600 dark:text-amber-300 ring-1 ring-inset ring-amber-500/30',
}

function Pill({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-medium leading-none',
        className
      )}
    >
      {children}
    </span>
  )
}

function LiveCasesCard() {
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title="Live Agent case investigations" icon={Sparkles} iconClass="text-fuchsia-500" />
      <div className="grid grid-cols-[110px_1fr_110px] gap-3 px-6 pb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>Status</div>
        <div>Case</div>
        <div className="text-right">Conclusion</div>
      </div>
      <div className="border-t border-border">
        {CASES.map((c) => (
          <div
            key={c.id}
            className="grid grid-cols-[110px_1fr_110px] items-center gap-3 border-b border-border px-6 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
          >
            <div>
              <Pill className={STATUS_STYLES[c.status]}>
                {c.status === 'Investigating' ? 'Investigating…' : c.status}
              </Pill>
            </div>
            <div className="truncate text-foreground/90">{c.title}</div>
            <div className="flex justify-end">
              {c.conclusion && <Pill className={CONCLUSION_STYLES[c.conclusion]}>{c.conclusion}</Pill>}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ─── MITRE ATT&CK Top techniques ──────────────────────────────────────── */

const MITRE: { tid: string; name: string; tactic: string; count: number }[] = [
  { tid: 'T1059.001', name: 'PowerShell', tactic: 'Execution', count: 124 },
  { tid: 'T1110', name: 'Brute Force', tactic: 'Credential Access', count: 87 },
  { tid: 'T1021.002', name: 'SMB / Admin Shares', tactic: 'Lateral Movement', count: 56 },
  { tid: 'T1547.001', name: 'Registry Run Keys', tactic: 'Persistence', count: 41 },
  { tid: 'T1027', name: 'Obfuscated Files', tactic: 'Defense Evasion', count: 34 },
  { tid: 'T1078', name: 'Valid Accounts', tactic: 'Initial Access', count: 22 },
]

function MitreTechniquesCard() {
  const max = Math.max(...MITRE.map((m) => m.count))
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader
        title="Top MITRE ATT&CK techniques"
        icon={Crosshair}
        iconClass="text-rose-500"
        action={{ label: 'Explore', href: '/threat-management/alerts' }}
      />
      <div className="space-y-2.5 px-6 pb-5">
        {MITRE.map((m) => (
          <div key={m.tid} className="grid grid-cols-[1fr_auto] items-center gap-3 text-xs">
            <div>
              <div className="flex items-baseline gap-2">
                <span className="font-medium">{m.name}</span>
                <span className="font-mono text-[10px] text-muted-foreground">{m.tid}</span>
              </div>
              <div className="mt-1 text-[11px] text-muted-foreground">{m.tactic}</div>
              <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-fuchsia-500 to-rose-500"
                  style={{ width: `${(m.count / max) * 100}%` }}
                />
              </div>
            </div>
            <div className="font-mono text-sm tabular-nums">{m.count}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ─── System Health ────────────────────────────────────────────────────── */

const PIPELINES: { name: string; status: 'ok' | 'warn' | 'error'; throughput: string }[] = [
  { name: 'Event ingestion', status: 'ok', throughput: '4.2k/s' },
  { name: 'Correlation engine', status: 'ok', throughput: '38ms' },
  { name: 'Threat intel sync', status: 'warn', throughput: 'lag 2m' },
  { name: 'SOC AI inference', status: 'ok', throughput: 'p95 1.1s' },
  { name: 'Index writer', status: 'ok', throughput: '99.98%' },
  { name: 'Email notifier', status: 'error', throughput: 'down' },
]

const HEALTH_DOT: Record<'ok' | 'warn' | 'error', string> = {
  ok: 'bg-emerald-500',
  warn: 'bg-amber-500',
  error: 'bg-red-500',
}

const HEALTH_LABEL: Record<'ok' | 'warn' | 'error', string> = {
  ok: 'Healthy',
  warn: 'Degraded',
  error: 'Down',
}

function SystemHealthCard() {
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title="System health" icon={Activity} iconClass="text-emerald-500" />
      <div className="divide-y divide-border">
        {PIPELINES.map((p) => (
          <div key={p.name} className="flex items-center justify-between px-6 py-2.5 text-xs">
            <div className="flex items-center gap-2">
              <span className={cn('h-2 w-2 rounded-full', HEALTH_DOT[p.status])} />
              <span>{p.name}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="font-mono text-[10px] text-muted-foreground">{p.throughput}</span>
              <span
                className={cn(
                  'rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none',
                  p.status === 'ok' && 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300',
                  p.status === 'warn' && 'bg-amber-500/15 text-amber-600 dark:text-amber-300',
                  p.status === 'error' && 'bg-red-500/15 text-red-500 dark:text-red-300'
                )}
              >
                {HEALTH_LABEL[p.status]}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ─── Top Targeted Assets ──────────────────────────────────────────────── */

const ASSETS: { name: string; type: string; alerts: number; risk: number }[] = [
  { name: 'WIN-7F3K2L9Q8X', type: 'Windows Server', alerts: 142, risk: 96 },
  { name: 'fin-app-01.corp.local', type: 'Application', alerts: 88, risk: 91 },
  { name: 'jsmith@utmstack.com', type: 'User', alerts: 64, risk: 84 },
  { name: 'db-prod-04', type: 'Linux Server', alerts: 47, risk: 73 },
  { name: '198.51.100.14', type: 'External IP', alerts: 38, risk: 68 },
  { name: 'edge-fw-bsa-01', type: 'Firewall', alerts: 22, risk: 51 },
]

function TopAssetsCard() {
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title="Top targeted assets" icon={Server} action={{ label: 'View all', href: '/datasources' }} />
      <div className="grid grid-cols-[1fr_120px_60px_60px] gap-3 px-6 pb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>Asset</div>
        <div>Type</div>
        <div className="text-right">Alerts</div>
        <div className="text-right">Risk</div>
      </div>
      <div className="border-t border-border">
        {ASSETS.map((a) => (
          <div
            key={a.name}
            className="grid grid-cols-[1fr_120px_60px_60px] items-center gap-3 border-b border-border px-6 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
          >
            <div className="truncate font-mono text-[12px]">{a.name}</div>
            <div className="text-muted-foreground">{a.type}</div>
            <div className="text-right tabular-nums">{a.alerts}</div>
            <div className="flex justify-end">
              <RiskBadge value={a.risk} />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function RiskBadge({ value }: { value: number }) {
  const tone =
    value >= 90
      ? 'bg-red-500/15 text-red-500 dark:text-red-300'
      : value >= 70
        ? 'bg-amber-500/15 text-amber-600 dark:text-amber-300'
        : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300'
  return (
    <span className={cn('inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-mono leading-none', tone)}>
      {value}
    </span>
  )
}

/* ─── Compliance Score ─────────────────────────────────────────────────── */

const COMPLIANCE: { name: string; score: number }[] = [
  { name: 'PCI DSS', score: 92 },
  { name: 'HIPAA', score: 88 },
  { name: 'SOX', score: 76 },
  { name: 'ISO 27001', score: 81 },
  { name: 'NIST 800-53', score: 91 },
  { name: 'GDPR', score: 84 },
]

function ComplianceCard() {
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title="Compliance score" icon={BadgeCheck} iconClass="text-emerald-500" action={{ label: 'Reports', href: '/compliance/manage' }} />
      <div className="px-6 pb-6">
        <div className="grid grid-cols-6 items-end gap-3 h-32">
          {COMPLIANCE.map((c) => (
            <div key={c.name} className="flex h-full flex-col items-center justify-end">
              <div
                className={cn(
                  'w-full rounded-t-md',
                  c.score >= 90
                    ? 'bg-gradient-to-t from-emerald-600 to-emerald-400'
                    : c.score >= 80
                      ? 'bg-gradient-to-t from-sky-600 to-sky-400'
                      : 'bg-gradient-to-t from-amber-600 to-amber-400'
                )}
                style={{ height: `${c.score}%` }}
              />
              <div className="mt-1 font-mono text-[10px] tabular-nums">{c.score}</div>
            </div>
          ))}
        </div>
        <div className="mt-2 grid grid-cols-6 gap-3 text-center text-[10px] uppercase tracking-wider text-muted-foreground">
          {COMPLIANCE.map((c) => (
            <div key={c.name} className="truncate">{c.name}</div>
          ))}
        </div>
      </div>
    </div>
  )
}

/* ─── Threat Intel Feed ────────────────────────────────────────────────── */

const IOCS: { type: string; value: string; source: string; tags: string[]; time: string }[] = [
  { type: 'IPv4', value: '198.51.100.14', source: 'AlienVault OTX', tags: ['C2', 'Scattered Spider'], time: '2m ago' },
  { type: 'Domain', value: 'malicious-update[.]com', source: 'AbuseCH', tags: ['Phishing'], time: '14m ago' },
  { type: 'SHA256', value: '9f86d…3b0c4 (truncated)', source: 'VirusTotal', tags: ['Ransomware', 'BlackBasta'], time: '38m ago' },
  { type: 'URL', value: 'hxxp://exfil[.]bad/upload', source: 'Internal', tags: ['Exfiltration'], time: '1h ago' },
  { type: 'CVE', value: 'CVE-2024-38077', source: 'NVD', tags: ['RCE', 'Windows'], time: '2h ago' },
]

const IOC_TYPE_STYLES: Record<string, string> = {
  IPv4: 'bg-sky-500/15 text-sky-500 dark:text-sky-300',
  Domain: 'bg-violet-500/15 text-violet-500 dark:text-violet-300',
  SHA256: 'bg-fuchsia-500/15 text-fuchsia-500 dark:text-fuchsia-300',
  URL: 'bg-amber-500/15 text-amber-600 dark:text-amber-300',
  CVE: 'bg-red-500/15 text-red-500 dark:text-red-300',
}

function ThreatIntelCard() {
  return (
    <div className="rounded-xl border border-border bg-card">
      <CardHeader title="Threat intelligence — recent IOCs" icon={Radar} iconClass="text-fuchsia-500" action={{ label: 'View feed', href: '/threat-intelligence' }} />
      <div className="grid grid-cols-[80px_1fr_180px_1fr_80px] gap-3 px-6 pb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>Type</div>
        <div>Indicator</div>
        <div>Source</div>
        <div>Tags</div>
        <div className="text-right">Seen</div>
      </div>
      <div className="border-t border-border">
        {IOCS.map((i, idx) => (
          <div
            key={idx}
            className="grid grid-cols-[80px_1fr_180px_1fr_80px] items-center gap-3 border-b border-border px-6 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
          >
            <Pill className={cn('justify-self-start font-mono', IOC_TYPE_STYLES[i.type])}>{i.type}</Pill>
            <div className="truncate font-mono text-[12px]">{i.value}</div>
            <div className="truncate text-muted-foreground">{i.source}</div>
            <div className="flex flex-wrap gap-1">
              {i.tags.map((t) => (
                <span key={t} className="rounded-md bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                  {t}
                </span>
              ))}
            </div>
            <div className="text-right text-[10px] text-muted-foreground">{i.time}</div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ─── Shared card header ───────────────────────────────────────────────── */

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
        <a href={action.href} className="text-xs font-medium text-primary hover:underline">
          {action.label}
        </a>
      )}
    </div>
  )
}
