import { useMemo, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  ChevronDown,
  Clock,
  Copy,
  Crosshair,
  Database,
  Download,
  ExternalLink,
  History,
  ListFilter,
  MoreHorizontal,
  Play,
  Plus,
  Search,
  Sparkles,
  Upload,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Severity = 'critical' | 'high' | 'medium' | 'low'
type DetectionKind = 'signature' | 'threshold' | 'correlation' | 'anomaly' | 'ml'
type Author = 'built-in' | 'custom'

interface Rule {
  id: string
  name: string
  description: string
  severity: Severity
  enabled: boolean
  kind: DetectionKind
  author: Author
  category: string
  technique?: { id: string; name: string }
  sources: string[]
  fires24h: number
  firesTotal: number
  lastTriggered?: string
  createdAt: string
  modifiedAt: string
  modifiedBy: string
  query: string
  threshold?: { count: number; window: string; groupBy?: string }
  tags: string[]
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const RULES: Rule[] = [
  {
    id: 'r-kerberoasting',
    name: 'Kerberoasting burst against domain controller',
    description: 'Detects unusually many TGS-REQ ticket requests with RC4 encryption from a single source — a strong signal of Kerberoasting credential extraction.',
    severity: 'critical',
    enabled: true,
    kind: 'threshold',
    author: 'built-in',
    category: 'Credential Access',
    technique: { id: 'T1558.003', name: 'Kerberoasting' },
    sources: ['Windows', 'AD'],
    fires24h: 12,
    firesTotal: 184,
    lastTriggered: min(4),
    createdAt: days(180),
    modifiedAt: hr(48),
    modifiedBy: 'system',
    query: `event.code: 4769
AND kerberos.ticket.encryption_type: "RC4-HMAC"
AND NOT user.name: ("krbtgt" OR "MSOL_*")`,
    threshold: { count: 10, window: '2 minutes', groupBy: 'source.user.name' },
    tags: ['credential-access', 'mitre', 'kerberos'],
  },
  {
    id: 'r-bruteforce',
    name: 'Multiple failed logins followed by success',
    description: 'Identifies brute-force attempts: more than 50 failed logins within 5 minutes followed by a successful authentication from the same source.',
    severity: 'high',
    enabled: true,
    kind: 'correlation',
    author: 'built-in',
    category: 'Credential Access',
    technique: { id: 'T1110', name: 'Brute Force' },
    sources: ['AD', 'M365', 'Okta'],
    fires24h: 8,
    firesTotal: 1240,
    lastTriggered: min(23),
    createdAt: days(220),
    modifiedAt: days(7),
    modifiedBy: 'jsmith',
    query: `sequence by source.ip
[ event.outcome: "failure" ] count >= 50
[ event.outcome: "success" ]
within 5 minutes`,
    tags: ['credential-access', 'mitre'],
  },
  {
    id: 'r-c2-ioc',
    name: 'Outbound connection to known C2 IP',
    description: 'Triggers when a host inside the network connects to an IP address present in any active threat-intel feed tagged as C2 infrastructure.',
    severity: 'critical',
    enabled: true,
    kind: 'signature',
    author: 'built-in',
    category: 'Command and Control',
    technique: { id: 'T1071.001', name: 'Web Protocols' },
    sources: ['Firewall', 'NetFlow', 'Sysmon'],
    fires24h: 5,
    firesTotal: 142,
    lastTriggered: min(46),
    createdAt: days(360),
    modifiedAt: days(2),
    modifiedBy: 'system',
    query: `destination.ip: IN(threat_intel.c2_feed)
AND NOT destination.ip: IN(allowlist.cdn)`,
    tags: ['c2', 'mitre', 'threat-intel'],
  },
  {
    id: 'r-mimikatz',
    name: 'Mimikatz signature match',
    description: 'AV / EDR signature match for Mimikatz family. Indicates credential dumping tooling on the host.',
    severity: 'critical',
    enabled: true,
    kind: 'signature',
    author: 'built-in',
    category: 'Credential Access',
    technique: { id: 'T1003', name: 'OS Credential Dumping' },
    sources: ['Windows Defender', 'CrowdStrike', 'SentinelOne'],
    fires24h: 1,
    firesTotal: 18,
    lastTriggered: min(62),
    createdAt: days(420),
    modifiedAt: days(30),
    modifiedBy: 'system',
    query: `threat.indicator.name: ("HackTool:Win32/Mimikatz*" OR "Mimikatz*")`,
    tags: ['malware', 'mitre'],
  },
  {
    id: 'r-impossible-travel',
    name: 'Impossible travel between authentications',
    description: 'A user authenticates from two geographically distant locations within a time window that makes physical travel implausible.',
    severity: 'high',
    enabled: true,
    kind: 'anomaly',
    author: 'built-in',
    category: 'Initial Access',
    technique: { id: 'T1078', name: 'Valid Accounts' },
    sources: ['M365', 'Azure AD', 'Okta'],
    fires24h: 3,
    firesTotal: 88,
    lastTriggered: min(8),
    createdAt: days(180),
    modifiedAt: days(12),
    modifiedBy: 'mthomas',
    query: `from auth.success
calc travel_distance(prev.geo, curr.geo) / time_delta
where speed > 800 km/h`,
    tags: ['identity', 'anomaly', 'mitre'],
  },
  {
    id: 'r-encoded-powershell',
    name: 'Encoded PowerShell execution',
    description: 'PowerShell execution with -EncodedCommand or -e flag. Common technique for living-off-the-land payloads.',
    severity: 'medium',
    enabled: true,
    kind: 'signature',
    author: 'built-in',
    category: 'Execution',
    technique: { id: 'T1059.001', name: 'PowerShell' },
    sources: ['Sysmon', 'Windows'],
    fires24h: 47,
    firesTotal: 4_412,
    lastTriggered: min(11),
    createdAt: days(540),
    modifiedAt: hr(8),
    modifiedBy: 'jsmith',
    query: `process.name: "powershell.exe"
AND process.command_line: (-enc OR -encodedcommand OR -e)
AND NOT process.parent.name: ("ScheduledTasks.exe")`,
    tags: ['execution', 'mitre', 'tuned'],
  },
  {
    id: 'r-smb-enum',
    name: 'Suspicious SMB share enumeration',
    description: 'A non-administrator account enumerates SMB shares across multiple hosts in a short time window.',
    severity: 'medium',
    enabled: true,
    kind: 'threshold',
    author: 'built-in',
    category: 'Discovery',
    technique: { id: 'T1135', name: 'Network Share Discovery' },
    sources: ['Windows', 'AD'],
    fires24h: 2,
    firesTotal: 34,
    lastTriggered: min(38),
    createdAt: days(120),
    modifiedAt: days(45),
    modifiedBy: 'system',
    query: `event.code: 5145
AND NOT user.is_administrator
AND share.name NOT IN allowlist`,
    threshold: { count: 5, window: '10 minutes', groupBy: 'user.name' },
    tags: ['discovery', 'mitre'],
  },
  {
    id: 'r-data-exfil-volume',
    name: 'Unusual data transfer volume',
    description: 'Outbound network traffic from a host exceeds 5× its 30-day baseline. ML-driven, learns each host\'s normal behavior.',
    severity: 'high',
    enabled: true,
    kind: 'ml',
    author: 'built-in',
    category: 'Exfiltration',
    technique: { id: 'T1041', name: 'C2 Exfiltration' },
    sources: ['NetFlow', 'Firewall'],
    fires24h: 1,
    firesTotal: 22,
    lastTriggered: min(74),
    createdAt: days(120),
    modifiedAt: days(7),
    modifiedBy: 'mthomas',
    query: `model: outbound_volume_baseline
where current.bytes / baseline.p99 > 5.0
and time_window = "1 hour"`,
    tags: ['exfiltration', 'ml', 'mitre'],
  },
  {
    id: 'r-gpo-mod',
    name: 'GPO modification by non-administrator',
    description: 'A user without explicit Group Policy edit rights modifies a Group Policy Object.',
    severity: 'high',
    enabled: true,
    kind: 'signature',
    author: 'built-in',
    category: 'Persistence',
    technique: { id: 'T1484', name: 'Domain Policy Modification' },
    sources: ['AD'],
    fires24h: 1,
    firesTotal: 6,
    lastTriggered: hr(2),
    createdAt: days(540),
    modifiedAt: days(180),
    modifiedBy: 'system',
    query: `event.code: 5136
AND object.class: "groupPolicyContainer"
AND NOT user.name: IN(domain_admins, gpo_editors)`,
    tags: ['persistence', 'mitre'],
  },
  {
    id: 'r-scheduled-task',
    name: 'New scheduled task on critical asset',
    description: 'A new scheduled task is created on a host tagged as critical-asset. Triggers regardless of who creates it.',
    severity: 'medium',
    enabled: true,
    kind: 'signature',
    author: 'custom',
    category: 'Persistence',
    technique: { id: 'T1053.005', name: 'Scheduled Task' },
    sources: ['Sysmon'],
    fires24h: 3,
    firesTotal: 84,
    lastTriggered: min(58),
    createdAt: days(45),
    modifiedAt: days(10),
    modifiedBy: 'jsmith',
    query: `event.code: 4698
AND host.tags: "critical-asset"`,
    tags: ['persistence', 'custom'],
  },
  {
    id: 'r-tor-exit',
    name: 'Authentication from Tor exit node',
    description: 'A successful authentication occurs from an IP listed as a Tor exit node. Disabled by default until you decide whether to allow Tor.',
    severity: 'medium',
    enabled: false,
    kind: 'signature',
    author: 'built-in',
    category: 'Initial Access',
    technique: { id: 'T1078', name: 'Valid Accounts' },
    sources: ['M365', 'Azure AD', 'Okta'],
    fires24h: 0,
    firesTotal: 142,
    createdAt: days(720),
    modifiedAt: days(360),
    modifiedBy: 'system',
    query: `event.outcome: "success"
AND source.ip: IN(threat_intel.tor_exit_nodes)`,
    tags: ['identity', 'mitre', 'disabled-by-default'],
  },
  {
    id: 'r-stale-pw-svc',
    name: 'Service account with stale password (>180d)',
    description: 'Service account whose password has not rotated in over 180 days. Long-lived passwords on service accounts increase Kerberoasting risk.',
    severity: 'low',
    enabled: true,
    kind: 'threshold',
    author: 'custom',
    category: 'Hygiene',
    sources: ['AD'],
    fires24h: 0,
    firesTotal: 12,
    lastTriggered: days(2),
    createdAt: days(60),
    modifiedAt: days(15),
    modifiedBy: 'mthomas',
    query: `user.kind: "service"
AND time_since(pwd_last_set) > 180d`,
    tags: ['hygiene', 'custom'],
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const SEV_STYLE: Record<Severity, { bar: string; pill: string; tone: string; label: string }> = {
  critical: { bar: 'bg-red-500', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', tone: 'text-red-500', label: 'Critical' },
  high: { bar: 'bg-orange-500', pill: 'bg-orange-500/15 text-orange-600 ring-orange-500/30 dark:text-orange-300', tone: 'text-orange-500', label: 'High' },
  medium: { bar: 'bg-amber-500', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', tone: 'text-amber-500', label: 'Medium' },
  low: { bar: 'bg-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300', tone: 'text-sky-500', label: 'Low' },
}

const KIND_META: Record<DetectionKind, { label: string; tone: string }> = {
  signature: { label: 'Signature', tone: 'text-sky-500' },
  threshold: { label: 'Threshold', tone: 'text-amber-500' },
  correlation: { label: 'Correlation', tone: 'text-violet-500' },
  anomaly: { label: 'Anomaly', tone: 'text-fuchsia-500' },
  ml: { label: 'ML', tone: 'text-emerald-500' },
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (r: Rule) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All rules', count: RULES.length, predicate: () => true },
  { id: 'enabled', label: 'Enabled', count: RULES.filter((r) => r.enabled).length, predicate: (r) => r.enabled },
  { id: 'disabled', label: 'Disabled', count: RULES.filter((r) => !r.enabled).length, predicate: (r) => !r.enabled },
  { id: 'critical', label: 'Critical & High', count: RULES.filter((r) => r.severity === 'critical' || r.severity === 'high').length, predicate: (r) => r.severity === 'critical' || r.severity === 'high' },
  { id: 'custom', label: 'Custom', count: RULES.filter((r) => r.author === 'custom').length, predicate: (r) => r.author === 'custom' },
  { id: 'firing', label: 'Firing today', count: RULES.filter((r) => r.fires24h > 0).length, predicate: (r) => r.fires24h > 0 },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AlertingRulesPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [openRule, setOpenRule] = useState<Rule | null>(null)
  // Local optimistic toggle state (so you can see the UI react)
  const [enabledOverrides, setEnabledOverrides] = useState<Record<string, boolean>>({})

  const isEnabled = (r: Rule) => enabledOverrides[r.id] ?? r.enabled

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return RULES.filter(v.predicate).filter((r) =>
      search
        ? (r.name + r.description + r.category + r.tags.join(' ') + (r.technique?.id ?? ''))
            .toLowerCase()
            .includes(search.toLowerCase())
        : true
    )
  }, [view, search])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <Header total={filtered.length} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <FiringOverviewCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <RuleInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} />

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <ListHeader />
        {filtered.map((r) => (
          <RuleRow
            key={r.id}
            rule={r}
            enabled={isEnabled(r)}
            onToggle={() =>
              setEnabledOverrides((prev) => ({ ...prev, [r.id]: !isEnabled(r) }))
            }
            onOpen={() => setOpenRule(r)}
          />
        ))}
        {filtered.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            No rules match this view.
          </div>
        )}
      </div>

      {openRule && <RuleDrawer rule={openRule} enabled={isEnabled(openRule)} onClose={() => setOpenRule(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Alerting rules</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} rules
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Detection logic that runs continuously against your event stream.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Upload size={14} className="mr-2" />
          Import
        </Button>
        <Button size="sm">
          <Plus size={14} className="mr-2" />
          New rule
        </Button>
      </div>
    </header>
  )
}

/* ─── Firing overview ──────────────────────────────────────────────────── */

function FiringOverviewCard() {
  const buckets = 60
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.max(0, Math.round(Math.abs(Math.sin(i / 6)) * 22 + Math.abs(Math.sin(i / 11)) * 8 + (i > 50 ? 14 : 0) + 4))
      ),
    []
  )
  const totalFires = RULES.reduce((s, r) => s + r.fires24h, 0)
  const enabled = RULES.filter((r) => r.enabled).length

  const w = 1000
  const h = 100
  const max = Math.max(...data) * 1.15
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - (v / max) * h)
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
            Detections fired · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{totalFires}</span>
            <span className="text-sm text-muted-foreground">
              from <span className="text-foreground">{enabled}</span> enabled rules
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="rulesGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(244 63 94)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(244 63 94)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#rulesGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(244 63 94)"
          strokeOpacity="0.85"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>

      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>24h ago</span>
        <span>12h ago</span>
        <span>now</span>
      </div>
    </div>
  )
}

/* ─── Insights ─────────────────────────────────────────────────────────── */

function RuleInsightsCard() {
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">SOC AI insights</h3>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">Encoded PowerShell</span> rule fired 47× in 24h —{' '}
        <span className="text-foreground">93% benign</span>.{' '}
        <button className="text-primary hover:underline">Tune threshold</button>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">14 MITRE techniques</span> uncovered — most-common
        gap is <span className="font-medium">Defense Evasion</span>.{' '}
        <button className="text-primary hover:underline">Add coverage</button>
      </div>

      <div className="rounded-lg border border-border bg-background/40 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">3 rules</span> haven't fired in 90 days.{' '}
        <button className="text-primary hover:underline">Review</button>
      </div>
    </div>
  )
}

/* ─── Saved view tabs ──────────────────────────────────────────────────── */

function SavedViewTabs({ current, onChange }: { current: string; onChange: (id: string) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {SAVED_VIEWS.map((v) => {
        const active = current === v.id
        return (
          <button
            key={v.id}
            onClick={() => onChange(v.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{v.label}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {v.count}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Toolbar ──────────────────────────────────────────────────────────── */

function Toolbar({ search, onSearch }: { search: string; onSearch: (s: string) => void }) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          placeholder="Search by name, description, MITRE, source, tag…"
          className="h-9 pl-9"
        />
      </div>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
      </Button>
    </div>
  )
}

/* ─── List ─────────────────────────────────────────────────────────────── */

const COLS = '4px 44px 1fr 110px 130px 80px 110px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <div />
      <div />
      <div>Rule</div>
      <div>Severity</div>
      <div>Type</div>
      <div className="text-right">24h</div>
      <div>Last fired</div>
      <div>Modified</div>
      <div />
    </div>
  )
}

function RuleRow({
  rule: r,
  enabled,
  onToggle,
  onOpen,
}: {
  rule: Rule
  enabled: boolean
  onToggle: () => void
  onOpen: () => void
}) {
  const sev = SEV_STYLE[r.severity]
  const km = KIND_META[r.kind]
  return (
    <div
      onClick={onOpen}
      className={cn(
        'group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs transition-colors hover:bg-muted/40 last:border-b-0',
        !enabled && 'opacity-60'
      )}
      style={{ gridTemplateColumns: COLS }}
    >
      <span className={cn('h-7 w-1 rounded-full', enabled ? sev.bar : 'bg-muted')} />
      <Toggle
        enabled={enabled}
        onChange={(e) => {
          e.stopPropagation()
          onToggle()
        }}
      />
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{r.name}</span>
          {r.author === 'custom' && (
            <span className="rounded bg-emerald-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-emerald-600 ring-1 ring-emerald-500/30 dark:text-emerald-300">
              custom
            </span>
          )}
        </div>
        <div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
          {r.technique && (
            <>
              <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
                {r.technique.id}
              </span>
              <span>{r.technique.name}</span>
              <span>·</span>
            </>
          )}
          <span>{r.category}</span>
          <span>·</span>
          <span className="font-mono">{r.sources.slice(0, 2).join(', ')}</span>
          {r.sources.length > 2 && <span className="text-[10px]">+{r.sources.length - 2}</span>}
        </div>
      </div>
      <div>
        <span className={cn('text-[11px] font-medium', sev.tone)}>{sev.label}</span>
      </div>
      <div className="text-[11px]">
        <span className={km.tone}>{km.label}</span>
      </div>
      <div className={cn('text-right tabular-nums', r.fires24h > 0 ? 'text-foreground' : 'text-muted-foreground')}>
        {r.fires24h > 0 ? r.fires24h.toLocaleString() : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {r.lastTriggered ? relativeTime(r.lastTriggered) : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {relativeTime(r.modifiedAt)}
        <div className="text-[10px] text-muted-foreground/70">{r.modifiedBy}</div>
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button
          onClick={(e) => e.stopPropagation()}
          className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground"
        >
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}

function Toggle({
  enabled,
  onChange,
}: {
  enabled: boolean
  onChange: (e: React.MouseEvent) => void
}) {
  return (
    <button
      onClick={onChange}
      className={cn(
        'relative h-4 w-7 shrink-0 rounded-full transition-colors',
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

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'overview' | 'logic' | 'history'

function RuleDrawer({
  rule: r,
  enabled,
  onClose,
}: {
  rule: Rule
  enabled: boolean
  onClose: () => void
}) {
  const [tab, setTab] = useState<Tab>('overview')
  const sev = SEV_STYLE[r.severity]
  const km = KIND_META[r.kind]

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[920px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className="font-mono">{r.id}</span>
                <span>·</span>
                <span className="capitalize">{r.author}</span>
                <span>·</span>
                <span>{r.category}</span>
              </div>
              <div className="mt-1 flex items-center gap-2">
                <span className={cn('h-5 w-1.5 rounded-full', enabled ? sev.bar : 'bg-muted')} />
                <h2 className="truncate text-lg font-semibold">{r.name}</h2>
              </div>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span className={cn('rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', sev.pill)}>
                  {sev.label}
                </span>
                <span className={cn('inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5 font-medium', km.tone)}>
                  {km.label}
                </span>
                {r.technique && (
                  <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5 font-mono">
                    <Crosshair size={10} className="text-rose-500" />
                    {r.technique.id} {r.technique.name}
                  </span>
                )}
                {r.tags.map((t) => (
                  <span key={t} className="rounded-md bg-muted px-1.5 py-0.5">
                    {t}
                  </span>
                ))}
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button size="sm">
              <Play size={13} className="mr-1.5" />
              Test against last 24h
            </Button>
            <Button size="sm" variant="outline">
              <Copy size={13} className="mr-1.5" />
              Duplicate
            </Button>
            <Button size="sm" variant="outline">
              <ExternalLink size={13} className="mr-1.5" />
              View matched alerts
            </Button>
            <span className="ml-auto inline-flex items-center gap-2 text-[11px] text-muted-foreground">
              <Toggle enabled={enabled} onChange={(e) => e.stopPropagation()} />
              {enabled ? 'Enabled' : 'Disabled'}
            </span>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="overview" current={tab} onChange={setTab} icon={Sparkles}>
            Overview
          </DrawerTab>
          <DrawerTab id="logic" current={tab} onChange={setTab} icon={Crosshair}>
            Logic
          </DrawerTab>
          <DrawerTab id="history" current={tab} onChange={setTab} icon={History}>
            Firing history
          </DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'overview' && <OverviewTab rule={r} />}
          {tab === 'logic' && <LogicTab rule={r} />}
          {tab === 'history' && <HistoryTab rule={r} />}
        </div>
      </div>
    </div>
  )
}

function DrawerTab({
  id,
  current,
  onChange,
  icon: Icon,
  children,
}: {
  id: Tab
  current: Tab
  onChange: (t: Tab) => void
  icon: LucideIcon
  children: React.ReactNode
}) {
  const active = id === current
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative flex items-center gap-1.5 px-3 py-2.5 text-xs transition-colors',
        active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
      )}
    >
      <Icon size={13} />
      {children}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

/* ─── Drawer: Overview ─────────────────────────────────────────────────── */

function OverviewTab({ rule: r }: { rule: Rule }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        <Section title="What this rule detects">
          <p className="text-sm leading-relaxed">{r.description}</p>
        </Section>

        <Section title="Required data sources">
          <div className="flex flex-wrap gap-2">
            {r.sources.map((s) => (
              <span
                key={s}
                className="inline-flex items-center gap-1.5 rounded-md border border-border bg-background/40 px-2 py-1 text-xs"
              >
                <Database size={11} className="text-muted-foreground" />
                {s}
              </span>
            ))}
          </div>
          <p className="mt-2 text-[11px] text-muted-foreground">
            Rule will only evaluate when at least one of these sources is connected.
          </p>
        </Section>

        <Section title="Fire trend · 14 days">
          <SparkCalendar />
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Stats"
          rows={[
            ['24h fires', r.fires24h.toLocaleString()],
            ['Total fires', r.firesTotal.toLocaleString()],
            ['Last fired', r.lastTriggered ? relativeTime(r.lastTriggered) : '—'],
          ]}
        />

        <KeyValueCard
          title="Lifecycle"
          rows={[
            ['Created', relativeTime(r.createdAt)],
            ['Modified', relativeTime(r.modifiedAt)],
            ['Modified by', <span className="font-mono">{r.modifiedBy}</span>],
            ['Author', <span className="capitalize">{r.author}</span>],
          ]}
        />
      </aside>
    </div>
  )
}

function SparkCalendar() {
  // 14-day mini cells
  const days14 = Array.from({ length: 14 }, (_, i) =>
    Math.max(0, Math.round(Math.abs(Math.sin((i + 3) / 1.7)) * 10 + (i === 13 ? 6 : 0)))
  )
  const max = Math.max(...days14, 1)
  return (
    <div className="flex items-end gap-1.5">
      {days14.map((v, i) => {
        const intensity = v / max
        return (
          <div key={i} className="flex flex-1 flex-col items-center gap-1">
            <div
              className={cn(
                'w-full rounded-sm',
                v === 0 ? 'bg-muted' : intensity > 0.7 ? 'bg-red-500/80' : intensity > 0.4 ? 'bg-orange-500/70' : 'bg-amber-500/60'
              )}
              style={{ height: `${4 + intensity * 36}px` }}
              title={`${v} fires`}
            />
            <span className="text-[9px] tabular-nums text-muted-foreground">{i === 13 ? 'now' : `-${13 - i}d`}</span>
          </div>
        )
      })}
    </div>
  )
}

/* ─── Drawer: Logic ────────────────────────────────────────────────────── */

function LogicTab({ rule: r }: { rule: Rule }) {
  return (
    <div className="space-y-4 p-6">
      <Section title="Detection query">
        <pre className="overflow-x-auto rounded-md bg-muted/40 p-3 font-mono text-[11px] leading-relaxed">
          {r.query}
        </pre>
        <div className="mt-2 flex items-center gap-3">
          <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
            <Copy size={11} /> Copy
          </button>
          <button className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
            <ExternalLink size={11} /> Open in editor
          </button>
        </div>
      </Section>

      {r.threshold && (
        <Section title="Threshold">
          <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
            <Row k="Count">
              <span className="font-mono">{r.threshold.count}</span>
            </Row>
            <Row k="Window">
              <span className="font-mono">{r.threshold.window}</span>
            </Row>
            {r.threshold.groupBy && (
              <Row k="Group by">
                <span className="font-mono">{r.threshold.groupBy}</span>
              </Row>
            )}
          </dl>
        </Section>
      )}

      <Section title="What happens when this rule matches">
        <ol className="list-decimal pl-5 text-sm leading-relaxed">
          <li>An alert is generated with severity <strong>{SEV_STYLE[r.severity].label}</strong>.</li>
          <li>The alert is auto-tagged with <span className="font-mono text-xs">{r.tags.join(', ')}</span>.</li>
          {r.fires24h >= 8 && (
            <li>If more than 5 alerts fire from this rule within 1 hour, they are auto-grouped into an incident.</li>
          )}
        </ol>
      </Section>
    </div>
  )
}

/* ─── Drawer: History ──────────────────────────────────────────────────── */

function HistoryTab({ rule: r }: { rule: Rule }) {
  // Build mock fire events
  const events = [
    { ts: r.lastTriggered ?? min(60), alert: 'Alert a-9f04c2', host: 'WIN-7F3K2L9Q8X' },
    { ts: hr(2), alert: 'Alert a-c014f0', host: 'fin-app-01' },
    { ts: hr(7), alert: 'Alert a-44a811', host: 'WIN-7F3K2L9Q8X' },
    { ts: hr(14), alert: 'Alert a-3e88a1', host: 'WIN-7F3K2L9Q8X' },
    { ts: hr(20), alert: 'Alert a-1ab32e', host: 'fin-app-01' },
  ].filter(() => r.fires24h > 0)
  if (events.length === 0) {
    return (
      <div className="p-6">
        <div className="rounded-lg border border-dashed border-border bg-card px-6 py-12 text-center text-sm text-muted-foreground">
          This rule has not fired in the selected time range.
        </div>
      </div>
    )
  }
  return (
    <div className="p-6">
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div
          className="grid grid-cols-[140px_1fr_180px_36px] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        >
          <div>When</div>
          <div>Alert</div>
          <div>Host</div>
          <div />
        </div>
        {events.map((e, i) => (
          <div
            key={i}
            className="grid cursor-pointer grid-cols-[140px_1fr_180px_36px] gap-3 border-b border-border/60 px-4 py-2 text-xs hover:bg-muted/30 last:border-b-0"
          >
            <div className="font-mono text-[11px] text-muted-foreground">
              {relativeTime(e.ts)}
            </div>
            <div className="truncate font-mono text-[11px]">{e.alert}</div>
            <div className="truncate font-mono text-[11px]">{e.host}</div>
            <div className="flex justify-end">
              <ExternalLink size={11} className="text-muted-foreground" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

function KeyValueCard({
  title,
  rows,
}: {
  title: string
  rows: [string, React.ReactNode][]
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      <dl className="grid grid-cols-[100px_1fr] gap-y-2 text-xs">
        {rows.map(([k, v]) => (
          <Row key={k} k={k}>
            {v}
          </Row>
        ))}
      </dl>
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
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

// Suppress unused-import warnings
void Activity
void AlertTriangle
void ChevronDown
void Clock
void Download
