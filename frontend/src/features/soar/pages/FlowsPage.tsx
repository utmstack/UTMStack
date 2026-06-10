import { useMemo, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  BadgeCheck,
  Bell,
  Calendar,
  CheckCircle2,
  Clock,
  Code2,
  Copy,
  Database,
  Diamond,
  Download,
  FileText,
  GitBranch,
  History,
  LayoutGrid,
  ListFilter,
  ListIcon,
  Mail,
  MoreHorizontal,
  Network,
  PauseCircle,
  PlayCircle,
  Plus,
  RefreshCw,
  RotateCw,
  Search,
  Send,
  Server,
  Shield,
  Sparkles,
  Terminal,
  TrendingUp,
  Webhook,
  Workflow,
  XCircle,
  Zap,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type FlowStatus = 'active' | 'disabled' | 'draft' | 'error'
type TriggerKind = 'alert' | 'schedule' | 'manual' | 'webhook' | 'incident'
type NodeKind = 'trigger' | 'condition' | 'action' | 'notify' | 'end'

interface FlowNode {
  id: string
  kind: NodeKind
  label: string
  detail?: string
  icon?: LucideIcon
  branch?: 'true' | 'false'
}

interface FlowEdge {
  from: string
  to: string
  label?: string
}

interface Run {
  id: string
  ts: string
  status: 'success' | 'failed' | 'running' | 'skipped'
  duration: number
  triggeredBy: string
}

interface Flow {
  id: string
  name: string
  description: string
  status: FlowStatus
  trigger: { kind: TriggerKind; detail: string }
  platforms: ('windows' | 'linux' | 'macos' | 'cloud' | 'network')[]
  tags: string[]
  owner: { initials: string; tone: string; name: string }
  createdAt: string
  updatedAt: string
  runs24h: number
  successRate: number
  avgDuration: number
  timeSavedMinutes: number
  runHistory: number[]
  recentRuns: Run[]
  nodes: FlowNode[]
  edges: FlowEdge[]
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const sineRuns = (seed: number, len = 24) =>
  Array.from({ length: len }, (_, i) => Math.max(0, Math.round(Math.abs(Math.sin((i + seed) / 2.4)) * 9 + (i > 18 ? 5 : 2))))

const FLOWS: Flow[] = [
  {
    id: 'fl-isolate-host',
    name: 'Containment Lite — Isolate compromised host',
    description:
      'When a critical alert names a host, isolate it via firewall API, capture forensics, and notify the analyst on call.',
    status: 'active',
    trigger: { kind: 'alert', detail: 'severity = critical AND host present' },
    platforms: ['windows', 'linux'],
    tags: ['containment', 'auto'],
    owner: { initials: 'JS', tone: 'bg-sky-500', name: 'Jordan Smith' },
    createdAt: days(34),
    updatedAt: hr(8),
    runs24h: 12,
    successRate: 0.92,
    avgDuration: 41,
    timeSavedMinutes: 240,
    runHistory: sineRuns(2),
    recentRuns: [
      { id: 'r-1', ts: min(35), status: 'success', duration: 38, triggeredBy: 'Alert a-9f04c2' },
      { id: 'r-2', ts: hr(2), status: 'success', duration: 44, triggeredBy: 'Alert a-c014f0' },
      { id: 'r-3', ts: hr(4), status: 'failed', duration: 12, triggeredBy: 'Alert a-44a811' },
      { id: 'r-4', ts: hr(6), status: 'success', duration: 39, triggeredBy: 'Alert a-3e88a1' },
      { id: 'r-5', ts: hr(11), status: 'success', duration: 42, triggeredBy: 'Alert a-1ab32e' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'Critical alert', detail: 'severity = critical', icon: AlertTriangle },
      { id: 'n-cond', kind: 'condition', label: 'Host known?', detail: 'check inventory', icon: Diamond },
      { id: 'n-iso', kind: 'action', label: 'Isolate host', detail: 'firewall API', icon: Shield, branch: 'true' },
      { id: 'n-fore', kind: 'action', label: 'Snapshot host', detail: 'forensic capture', icon: Database },
      { id: 'n-not', kind: 'notify', label: 'Notify on-call', detail: 'Slack #soc', icon: Bell },
      { id: 'n-end', kind: 'end', label: 'Done' },
      { id: 'n-skip', kind: 'end', label: 'Skip · log', branch: 'false' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-cond' },
      { from: 'n-cond', to: 'n-iso', label: 'yes' },
      { from: 'n-cond', to: 'n-skip', label: 'no' },
      { from: 'n-iso', to: 'n-fore' },
      { from: 'n-fore', to: 'n-not' },
      { from: 'n-not', to: 'n-end' },
    ],
  },
  {
    id: 'fl-block-ip',
    name: 'Block C2 IP at perimeter',
    description: 'Push known-bad IP to edge firewall block-list and update IOC database.',
    status: 'active',
    trigger: { kind: 'alert', detail: 'IOC match · feed = C2' },
    platforms: ['network'],
    tags: ['network', 'threat-intel', 'auto'],
    owner: { initials: 'MT', tone: 'bg-emerald-500', name: 'Mara Thomas' },
    createdAt: days(21),
    updatedAt: hr(2),
    runs24h: 47,
    successRate: 0.99,
    avgDuration: 8,
    timeSavedMinutes: 480,
    runHistory: sineRuns(5),
    recentRuns: [
      { id: 'r-1', ts: min(8), status: 'success', duration: 7, triggeredBy: 'IOC feed sync' },
      { id: 'r-2', ts: min(38), status: 'success', duration: 9, triggeredBy: 'Alert a-c014f0' },
      { id: 'r-3', ts: hr(1), status: 'success', duration: 8, triggeredBy: 'Manual' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'IOC C2 match', icon: AlertTriangle },
      { id: 'n-block', kind: 'action', label: 'Push to firewall', detail: 'edge-fw-bsa-01', icon: Shield },
      { id: 'n-iocdb', kind: 'action', label: 'Update IOC db', icon: Database },
      { id: 'n-end', kind: 'end', label: 'Done' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-block' },
      { from: 'n-block', to: 'n-iocdb' },
      { from: 'n-iocdb', to: 'n-end' },
    ],
  },
  {
    id: 'fl-disable-account',
    name: 'Disable account on Kerberoasting',
    description: 'When Kerberoasting burst is detected, disable the source service account in AD.',
    status: 'active',
    trigger: { kind: 'alert', detail: 'technique = T1558.003 AND source.user matches svc_*' },
    platforms: ['windows'],
    tags: ['identity', 'campaign-44'],
    owner: { initials: 'JS', tone: 'bg-sky-500', name: 'Jordan Smith' },
    createdAt: days(12),
    updatedAt: hr(48),
    runs24h: 2,
    successRate: 1,
    avgDuration: 22,
    timeSavedMinutes: 120,
    runHistory: sineRuns(11).map((v, i) => (i > 19 ? v : 0)),
    recentRuns: [
      { id: 'r-1', ts: min(40), status: 'success', duration: 21, triggeredBy: 'Alert a-9f04c2' },
      { id: 'r-2', ts: hr(7), status: 'success', duration: 24, triggeredBy: 'Manual' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'Kerberoasting alert', icon: AlertTriangle },
      { id: 'n-cond', kind: 'condition', label: 'svc account?', icon: Diamond },
      { id: 'n-disable', kind: 'action', label: 'Disable AD account', detail: 'PowerShell', icon: Shield, branch: 'true' },
      { id: 'n-rotate', kind: 'action', label: 'Schedule krbtgt rotate', icon: RotateCw },
      { id: 'n-not', kind: 'notify', label: 'Email IT lead', icon: Mail },
      { id: 'n-end', kind: 'end', label: 'Done' },
      { id: 'n-skip', kind: 'end', label: 'Manual review', branch: 'false' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-cond' },
      { from: 'n-cond', to: 'n-disable', label: 'yes' },
      { from: 'n-cond', to: 'n-skip', label: 'no' },
      { from: 'n-disable', to: 'n-rotate' },
      { from: 'n-rotate', to: 'n-not' },
      { from: 'n-not', to: 'n-end' },
    ],
  },
  {
    id: 'fl-phish-triage',
    name: 'Phishing email — auto triage',
    description:
      'Pull reported message, detonate URLs in sandbox, classify with AI, quarantine if malicious.',
    status: 'active',
    trigger: { kind: 'webhook', detail: 'm365-report-message' },
    platforms: ['cloud'],
    tags: ['phishing', 'email'],
    owner: { initials: 'AL', tone: 'bg-violet-500', name: 'Ana Luna' },
    createdAt: days(60),
    updatedAt: hr(36),
    runs24h: 18,
    successRate: 0.88,
    avgDuration: 96,
    timeSavedMinutes: 540,
    runHistory: sineRuns(8),
    recentRuns: [
      { id: 'r-1', ts: min(11), status: 'success', duration: 92, triggeredBy: 'Webhook · jdoe@corp' },
      { id: 'r-2', ts: min(45), status: 'failed', duration: 14, triggeredBy: 'Webhook · ksmith@corp' },
      { id: 'r-3', ts: hr(2), status: 'success', duration: 102, triggeredBy: 'Webhook · finance@corp' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'Reported phish', icon: Webhook },
      { id: 'n-pull', kind: 'action', label: 'Pull message', detail: 'Graph API', icon: Mail },
      { id: 'n-deton', kind: 'action', label: 'Detonate URL', detail: 'sandbox', icon: Zap },
      { id: 'n-ai', kind: 'action', label: 'AI classify', detail: 'gpt-class-v3', icon: Sparkles },
      { id: 'n-cond', kind: 'condition', label: 'Malicious?', icon: Diamond },
      { id: 'n-quar', kind: 'action', label: 'Quarantine + alert IT', icon: Shield, branch: 'true' },
      { id: 'n-mark', kind: 'action', label: 'Mark FP', branch: 'false', icon: BadgeCheck },
      { id: 'n-end', kind: 'end', label: 'Done' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-pull' },
      { from: 'n-pull', to: 'n-deton' },
      { from: 'n-deton', to: 'n-ai' },
      { from: 'n-ai', to: 'n-cond' },
      { from: 'n-cond', to: 'n-quar', label: 'yes' },
      { from: 'n-cond', to: 'n-mark', label: 'no' },
      { from: 'n-quar', to: 'n-end' },
      { from: 'n-mark', to: 'n-end' },
    ],
  },
  {
    id: 'fl-daily-vuln',
    name: 'Daily vuln-scan summary',
    description: 'Pull yesterday\'s scan results, summarize new criticals, post to leadership channel.',
    status: 'active',
    trigger: { kind: 'schedule', detail: 'every day · 08:00 local' },
    platforms: ['cloud'],
    tags: ['scheduled', 'reporting'],
    owner: { initials: 'TF', tone: 'bg-amber-500', name: 'Tomás Ferguson' },
    createdAt: days(90),
    updatedAt: days(2),
    runs24h: 1,
    successRate: 0.97,
    avgDuration: 220,
    timeSavedMinutes: 30,
    runHistory: Array(24).fill(0).map((_, i) => (i === 8 ? 1 : 0)),
    recentRuns: [
      { id: 'r-1', ts: hr(8), status: 'success', duration: 218, triggeredBy: 'Schedule · 08:00' },
      { id: 'r-2', ts: hr(32), status: 'success', duration: 224, triggeredBy: 'Schedule · 08:00' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: '08:00 daily', icon: Calendar },
      { id: 'n-pull', kind: 'action', label: 'Fetch scan report', icon: Database },
      { id: 'n-sum', kind: 'action', label: 'Summarize', detail: 'AI digest', icon: Sparkles },
      { id: 'n-post', kind: 'notify', label: 'Post to channel', detail: 'Slack #leadership', icon: Send },
      { id: 'n-end', kind: 'end', label: 'Done' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-pull' },
      { from: 'n-pull', to: 'n-sum' },
      { from: 'n-sum', to: 'n-post' },
      { from: 'n-post', to: 'n-end' },
    ],
  },
  {
    id: 'fl-onboard',
    name: 'New asset onboarding',
    description:
      'When a new agent registers, validate config, install baseline rules, post to inventory channel.',
    status: 'disabled',
    trigger: { kind: 'webhook', detail: 'agent-register' },
    platforms: ['linux', 'windows'],
    tags: ['onboarding'],
    owner: { initials: 'MT', tone: 'bg-emerald-500', name: 'Mara Thomas' },
    createdAt: days(45),
    updatedAt: days(7),
    runs24h: 0,
    successRate: 0.85,
    avgDuration: 130,
    timeSavedMinutes: 0,
    runHistory: Array(24).fill(0),
    recentRuns: [
      { id: 'r-1', ts: days(2), status: 'success', duration: 138, triggeredBy: 'Webhook' },
      { id: 'r-2', ts: days(3), status: 'success', duration: 121, triggeredBy: 'Webhook' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'Agent registered', icon: Webhook },
      { id: 'n-val', kind: 'action', label: 'Validate config', icon: BadgeCheck },
      { id: 'n-rule', kind: 'action', label: 'Push baseline rules', icon: Shield },
      { id: 'n-end', kind: 'end', label: 'Done' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-val' },
      { from: 'n-val', to: 'n-rule' },
      { from: 'n-rule', to: 'n-end' },
    ],
  },
  {
    id: 'fl-draft-mfa',
    name: 'Force MFA on suspicious login (DRAFT)',
    description:
      'When a login from an impossible-travel pattern is detected, force MFA challenge on next request.',
    status: 'draft',
    trigger: { kind: 'alert', detail: 'pattern = impossible-travel' },
    platforms: ['cloud'],
    tags: ['identity', 'in-design'],
    owner: { initials: 'JS', tone: 'bg-sky-500', name: 'Jordan Smith' },
    createdAt: days(3),
    updatedAt: hr(5),
    runs24h: 0,
    successRate: 0,
    avgDuration: 0,
    timeSavedMinutes: 0,
    runHistory: Array(24).fill(0),
    recentRuns: [],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'Impossible travel', icon: AlertTriangle },
      { id: 'n-mfa', kind: 'action', label: 'Force MFA', icon: Shield },
      { id: 'n-end', kind: 'end', label: 'Done' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-mfa' },
      { from: 'n-mfa', to: 'n-end' },
    ],
  },
  {
    id: 'fl-failing',
    name: 'Slack notify on critical incident',
    description: 'Post a structured message to #soc whenever a critical incident is opened.',
    status: 'error',
    trigger: { kind: 'incident', detail: 'severity = critical' },
    platforms: ['cloud'],
    tags: ['notification'],
    owner: { initials: 'AL', tone: 'bg-violet-500', name: 'Ana Luna' },
    createdAt: days(120),
    updatedAt: hr(1),
    runs24h: 8,
    successRate: 0.25,
    avgDuration: 4,
    timeSavedMinutes: 12,
    runHistory: sineRuns(15).map((v, i) => (i > 18 ? Math.min(v, 3) : v)),
    recentRuns: [
      { id: 'r-1', ts: min(20), status: 'failed', duration: 3, triggeredBy: 'Incident #1142' },
      { id: 'r-2', ts: min(45), status: 'failed', duration: 4, triggeredBy: 'Incident #1141' },
      { id: 'r-3', ts: hr(1), status: 'success', duration: 4, triggeredBy: 'Incident #1140' },
    ],
    nodes: [
      { id: 'n-trig', kind: 'trigger', label: 'Critical incident', icon: AlertTriangle },
      { id: 'n-form', kind: 'action', label: 'Format message', icon: FileText },
      { id: 'n-slack', kind: 'notify', label: 'Post to Slack', detail: '#soc', icon: Send },
      { id: 'n-end', kind: 'end', label: 'Done' },
    ],
    edges: [
      { from: 'n-trig', to: 'n-form' },
      { from: 'n-form', to: 'n-slack' },
      { from: 'n-slack', to: 'n-end' },
    ],
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const STATUS_META: Record<FlowStatus, { label: string; pill: string; dot: string }> = {
  active: { label: 'Active', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
  disabled: { label: 'Disabled', pill: 'bg-muted text-muted-foreground ring-border', dot: 'bg-muted-foreground' },
  draft: { label: 'Draft', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
  error: { label: 'Failing', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500 animate-pulse' },
}

const TRIGGER_META: Record<TriggerKind, { label: string; icon: LucideIcon; tone: string }> = {
  alert: { label: 'Alert match', icon: AlertTriangle, tone: 'text-orange-500' },
  schedule: { label: 'Schedule', icon: Calendar, tone: 'text-sky-500' },
  manual: { label: 'Manual', icon: PlayCircle, tone: 'text-violet-500' },
  webhook: { label: 'Webhook', icon: Webhook, tone: 'text-fuchsia-500' },
  incident: { label: 'Incident', icon: AlertTriangle, tone: 'text-red-500' },
}

const NODE_META: Record<NodeKind, { tone: string; ring: string; icon: LucideIcon }> = {
  trigger: { tone: 'text-orange-500', ring: 'ring-orange-500/40', icon: Zap },
  condition: { tone: 'text-amber-500', ring: 'ring-amber-500/40', icon: Diamond },
  action: { tone: 'text-sky-500', ring: 'ring-sky-500/40', icon: Workflow },
  notify: { tone: 'text-violet-500', ring: 'ring-violet-500/40', icon: Bell },
  end: { tone: 'text-emerald-500', ring: 'ring-emerald-500/40', icon: CheckCircle2 },
}

const PLATFORM_ICON: Record<string, LucideIcon> = {
  windows: Server,
  linux: Terminal,
  macos: Server,
  cloud: Network,
  network: Network,
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (f: Flow) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All flows', count: FLOWS.length, predicate: () => true },
  { id: 'active', label: 'Active', count: FLOWS.filter((f) => f.status === 'active').length, predicate: (f) => f.status === 'active' },
  { id: 'failing', label: 'Failing', count: FLOWS.filter((f) => f.status === 'error').length, predicate: (f) => f.status === 'error' },
  { id: 'draft', label: 'Drafts', count: FLOWS.filter((f) => f.status === 'draft').length, predicate: (f) => f.status === 'draft' },
  { id: 'mine', label: 'Mine', count: FLOWS.filter((f) => f.owner.initials === 'JS').length, predicate: (f) => f.owner.initials === 'JS' },
  { id: 'auto', label: 'Auto-trigger', count: FLOWS.filter((f) => f.trigger.kind !== 'manual').length, predicate: (f) => f.trigger.kind !== 'manual' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function FlowsPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'grid' | 'list'>('list')
  const [openFlow, setOpenFlow] = useState<Flow | null>(null)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return FLOWS.filter(v.predicate).filter((f) =>
      search ? (f.name + f.description + f.tags.join(' ')).toLowerCase().includes(search.toLowerCase()) : true
    )
  }, [view, search])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <Header total={filtered.length} />

      <div className="mt-5">
        <RunsOverviewCard />
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} layout={layout} onLayoutChange={setLayout} />

      {layout === 'grid' ? (
        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((f) => (
            <FlowCard key={f.id} flow={f} onOpen={() => setOpenFlow(f)} />
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center">
              <Workflow className="mx-auto mb-3 text-muted-foreground" size={32} strokeWidth={1.5} />
              <h4 className="text-sm font-semibold">No flows match this view</h4>
              <p className="mt-1 text-xs text-muted-foreground">
                Start from a template or build a new flow from scratch.
              </p>
              <Button size="sm" className="mt-4">
                <Plus size={13} className="mr-1.5" />
                New flow
              </Button>
            </div>
          )}
        </div>
      ) : (
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
          <ListHeader />
          {filtered.map((f) => (
            <FlowListRow key={f.id} flow={f} onOpen={() => setOpenFlow(f)} />
          ))}
          {filtered.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">No flows match this view.</div>
          )}
        </div>
      )}

      {openFlow && <FlowDrawer flow={openFlow} onClose={() => setOpenFlow(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">SOAR · Flows</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} flows
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Automation playbooks. Each flow listens for a trigger and runs a sequence of actions.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <FileText size={14} className="mr-2" />
          Templates
        </Button>
        <Button variant="default" size="sm">
          <Plus size={14} className="mr-2" />
          New flow
        </Button>
      </div>
    </header>
  )
}

/* ─── Runs overview card ──────────────────────────────────────────────── */

function RunsOverviewCard() {
  const buckets = 48
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.round(Math.abs(Math.sin(i / 4)) * 18 + Math.abs(Math.sin(i / 9)) * 6 + 4)
      ),
    []
  )
  const total = 412
  const successRate = 93.2
  const failed = 28
  const avg = '47s'

  // Smooth area chart geometry
  const w = 1000
  const h = 120
  const pl = 0
  const pr = 0
  const pt = 6
  const pb = 0
  const innerW = w - pl - pr
  const innerH = h - pt - pb
  const max = Math.max(...data) * 1.1
  const xs = data.map((_, i) => pl + (i * innerW) / (data.length - 1))
  const ys = data.map((v) => pt + innerH - (v / max) * innerH)

  // Build a smooth catmull-rom-ish path
  const linePath = data.reduce((acc, _, i) => {
    if (i === 0) return `M ${xs[i]} ${ys[i]}`
    const prevX = xs[i - 1]
    const prevY = ys[i - 1]
    const cx1 = prevX + (xs[i] - prevX) / 2
    const cx2 = xs[i] - (xs[i] - prevX) / 2
    return `${acc} C ${cx1} ${prevY}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
  }, '')
  const areaPath = `${linePath} L ${xs[xs.length - 1]} ${pt + innerH} L ${xs[0]} ${pt + innerH} Z`

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      {/* Top row — title + metric inline, like Linear/Vercel */}
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Runs · last 24 hours</div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{total.toLocaleString()}</span>
            <span className="text-sm text-muted-foreground">
              <span className="text-emerald-500">{successRate}%</span> success
              <span className="mx-2 text-border">·</span>
              <span className="text-red-500">{failed}</span> failed
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{avg}</span> avg
            </span>
          </div>
        </div>
      </div>

      {/* Smooth area chart */}
      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-28 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="runsGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(139 92 246)" stopOpacity="0.32" />
            <stop offset="100%" stopColor="rgb(139 92 246)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#runsGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(139 92 246)"
          strokeOpacity="0.85"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>

      {/* Tiny period markers — minimal */}
      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>24h ago</span>
        <span>12h ago</span>
        <span>now</span>
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

function Toolbar({
  search,
  onSearch,
  layout,
  onLayoutChange,
}: {
  search: string
  onSearch: (s: string) => void
  layout: 'grid' | 'list'
  onLayoutChange: (l: 'grid' | 'list') => void
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search by name, description, tag…"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9 pr-24"
        />
      </div>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
      </Button>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
        <button
          onClick={() => onLayoutChange('grid')}
          title="Grid"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'grid' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <LayoutGrid size={13} />
        </button>
        <button
          onClick={() => onLayoutChange('list')}
          title="List"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <ListIcon size={13} />
        </button>
      </div>
      <button
        title="Refresh"
        className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <RefreshCw size={14} />
      </button>
    </div>
  )
}

/* ─── Flow card ────────────────────────────────────────────────────────── */

function FlowCard({ flow, onOpen }: { flow: Flow; onOpen: () => void }) {
  const st = STATUS_META[flow.status]
  return (
    <button
      onClick={onOpen}
      className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-center gap-2">
        <span className={cn('h-2 w-2 rounded-full', st.dot)} />
        <span className="text-[11px] uppercase tracking-wider text-muted-foreground">{st.label}</span>
      </div>

      <div className="min-w-0">
        <h4 className="truncate text-sm font-semibold">{flow.name}</h4>
        <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{flow.description}</p>
      </div>

      <div className="mt-auto flex items-center justify-between border-t border-border pt-3 text-[11px] text-muted-foreground">
        <span className="flex items-center gap-1.5">
          <span className={cn('flex h-5 w-5 items-center justify-center rounded-full text-[9px] font-semibold text-white', flow.owner.tone)}>
            {flow.owner.initials}
          </span>
          {flow.owner.name}
        </span>
        <span className="font-mono tabular-nums">
          {flow.runs24h > 0 ? `${flow.runs24h} runs · ${Math.round(flow.successRate * 100)}%` : 'no runs'}
        </span>
      </div>
    </button>
  )
}


/* ─── List view ────────────────────────────────────────────────────────── */

const LIST_COLS = '90px 1fr 130px 90px 80px 80px 110px 36px'

function ListHeader() {
  return (
    <div className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: LIST_COLS }}>
      <div>Status</div>
      <div>Flow</div>
      <div>Trigger</div>
      <div className="text-right">24h</div>
      <div className="text-right">Success</div>
      <div className="text-right">Avg</div>
      <div>Updated</div>
      <div />
    </div>
  )
}

function FlowListRow({ flow, onOpen }: { flow: Flow; onOpen: () => void }) {
  const st = STATUS_META[flow.status]
  const tg = TRIGGER_META[flow.trigger.kind]
  const TgIcon = tg.icon
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div>
        <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', st.pill)}>
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>
      <div className="min-w-0">
        <div className="truncate font-medium">{flow.name}</div>
        <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className={cn('flex h-4 w-4 items-center justify-center rounded-full text-[8px] font-semibold text-white', flow.owner.tone)}>
            {flow.owner.initials}
          </span>
          <span>{flow.owner.name}</span>
          {flow.tags.length > 0 && (
            <>
              <span>·</span>
              {flow.tags.slice(0, 2).map((t) => (
                <span key={t} className="rounded bg-muted px-1.5 py-0.5 text-[10px]">
                  {t}
                </span>
              ))}
            </>
          )}
        </div>
      </div>
      <div>
        <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5 text-[10px]">
          <TgIcon size={10} className={tg.tone} />
          {tg.label}
        </span>
      </div>
      <div className="text-right tabular-nums">{flow.runs24h}</div>
      <div className={cn('text-right tabular-nums', flow.successRate >= 0.9 ? 'text-emerald-500' : flow.successRate >= 0.7 ? 'text-amber-500' : 'text-red-500')}>
        {flow.runs24h > 0 ? `${Math.round(flow.successRate * 100)}%` : '—'}
      </div>
      <div className="text-right font-mono text-[11px] text-muted-foreground">{flow.avgDuration > 0 ? `${flow.avgDuration}s` : '—'}</div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(flow.updatedAt)}</div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button onClick={(e) => e.stopPropagation()} className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground hover:bg-background hover:text-foreground">
          <MoreHorizontal size={13} />
        </button>
      </div>
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'overview' | 'canvas' | 'runs' | 'audit'

function FlowDrawer({ flow, onClose }: { flow: Flow; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('overview')
  const st = STATUS_META[flow.status]
  const tg = TRIGGER_META[flow.trigger.kind]
  const TgIcon = tg.icon

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[1080px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className="font-mono">{flow.id}</span>
                <span>·</span>
                <span>created {relativeTime(flow.createdAt)}</span>
                <span>·</span>
                <span>v3 · last edited {relativeTime(flow.updatedAt)} by {flow.owner.name}</span>
              </div>
              <div className="mt-1 flex items-center gap-2">
                <Workflow size={20} className="text-violet-500" />
                <h2 className="truncate text-lg font-semibold">{flow.name}</h2>
              </div>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', st.pill)}>
                  <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                  {st.label}
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5 font-medium">
                  <TgIcon size={10} className={tg.tone} />
                  {tg.label} · {flow.trigger.detail}
                </span>
                {flow.platforms.map((p) => {
                  const Icon = PLATFORM_ICON[p] ?? Server
                  return (
                    <span key={p} className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5 capitalize">
                      <Icon size={10} />
                      {p}
                    </span>
                  )
                })}
                {flow.tags.map((t) => (
                  <span key={t} className="rounded-md bg-muted px-1.5 py-0.5">{t}</span>
                ))}
              </div>
            </div>
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
              <XCircle size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button size="sm">
              <PlayCircle size={14} className="mr-1.5" />
              Run now
            </Button>
            <Button size="sm" variant="outline">
              <Code2 size={14} className="mr-1.5" />
              Open in editor
            </Button>
            {flow.status === 'active' ? (
              <Button size="sm" variant="outline">
                <PauseCircle size={14} className="mr-1.5" />
                Disable
              </Button>
            ) : (
              <Button size="sm" variant="outline">
                <PlayCircle size={14} className="mr-1.5 text-emerald-500" />
                Enable
              </Button>
            )}
            <Button size="sm" variant="outline">
              <Copy size={14} className="mr-1.5" />
              Duplicate
            </Button>
            <Button size="sm" variant="outline">
              <Download size={14} className="mr-1.5" />
              Export
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={14} />
            </Button>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="overview" current={tab} onChange={setTab} icon={Sparkles}>Overview</DrawerTab>
          <DrawerTab id="canvas" current={tab} onChange={setTab} icon={Workflow}>Canvas</DrawerTab>
          <DrawerTab id="runs" current={tab} onChange={setTab} icon={History}>
            Runs
            <span className="ml-1.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
              {flow.recentRuns.length}
            </span>
          </DrawerTab>
          <DrawerTab id="audit" current={tab} onChange={setTab} icon={GitBranch}>Audit</DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'overview' && <OverviewTab flow={flow} />}
          {tab === 'canvas' && <CanvasTab flow={flow} />}
          {tab === 'runs' && <RunsTab flow={flow} />}
          {tab === 'audit' && <AuditTab flow={flow} />}
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

function OverviewTab({ flow }: { flow: Flow }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        {flow.status === 'error' && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-4 text-xs">
            <div className="flex items-center gap-2 font-medium text-red-600 dark:text-red-300">
              <XCircle size={13} />
              This flow is failing
            </div>
            <p className="mt-1 text-muted-foreground">
              Last 4 runs returned <span className="font-mono">401 Unauthorized</span> from Slack.
              Likely an expired bot token. <button className="text-primary hover:underline">Re-authenticate</button>
            </p>
          </div>
        )}

        <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium">
            <Sparkles size={14} className="text-fuchsia-500" />
            Flow at a glance
          </div>
          <p className="text-sm leading-relaxed">
            {flow.description} This flow saves an estimated{' '}
            <strong>{Math.round(flow.timeSavedMinutes / 60)}h {flow.timeSavedMinutes % 60}m</strong> of analyst time per
            week. Based on its trigger pattern, it would also be a good fit for handling{' '}
            <span className="font-medium">incident #1135</span>.
          </p>
        </div>

        <Section title="Description">
          <p className="text-sm text-foreground/90">{flow.description}</p>
        </Section>

        <Section title="Run history · 24h">
          <SuccessFailureChart history={flow.runHistory} />
        </Section>

        <Section title="Recent runs">
          <ul className="space-y-1.5">
            {flow.recentRuns.slice(0, 5).map((r) => (
              <li key={r.id} className="flex items-center justify-between rounded border border-border bg-background/40 px-3 py-1.5 text-xs">
                <div className="flex items-center gap-2">
                  <RunStatusDot status={r.status} />
                  <span className="font-mono text-[10px] text-muted-foreground">{r.id}</span>
                  <span className="truncate">{r.triggeredBy}</span>
                </div>
                <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
                  <span className="font-mono">{r.duration}s</span>
                  <span>{relativeTime(r.ts)}</span>
                </div>
              </li>
            ))}
            {flow.recentRuns.length === 0 && (
              <li className="rounded border border-dashed border-border bg-background/30 px-3 py-3 text-center text-xs italic text-muted-foreground">
                No runs yet — this flow has not been triggered.
              </li>
            )}
          </ul>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">Performance</div>
          <div className="space-y-3 text-xs">
            <Metric label="24h runs" value={flow.runs24h.toString()} icon={Activity} />
            <Metric label="Success rate" value={flow.runs24h > 0 ? `${Math.round(flow.successRate * 100)}%` : '—'} icon={CheckCircle2} tone={flow.successRate >= 0.9 ? 'text-emerald-500' : 'text-amber-500'} />
            <Metric label="Avg duration" value={flow.avgDuration > 0 ? `${flow.avgDuration}s` : '—'} icon={Clock} />
            <Metric label="Time saved (7d)" value={`${Math.round(flow.timeSavedMinutes / 60)}h`} icon={TrendingUp} tone="text-emerald-500" />
          </div>
        </div>

        <KeyValueCard
          title="Detail"
          rows={[
            ['Owner', <span className="inline-flex items-center gap-1.5">
              <span className={cn('flex h-5 w-5 items-center justify-center rounded-full text-[9px] font-semibold text-white', flow.owner.tone)}>
                {flow.owner.initials}
              </span>
              {flow.owner.name}
            </span>],
            ['Trigger', flow.trigger.detail],
            ['Platforms', flow.platforms.map((p) => p).join(', ')],
            ['Created', absTimestamp(flow.createdAt)],
            ['Updated', absTimestamp(flow.updatedAt)],
            ['Version', 'v3'],
          ]}
        />

        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
            Connected systems
          </div>
          <ul className="space-y-1.5 text-xs">
            <li className="flex items-center gap-2">
              <Shield size={11} className="text-sky-500" />
              edge-fw-bsa-01 <span className="text-muted-foreground">· firewall API</span>
            </li>
            <li className="flex items-center gap-2">
              <Database size={11} className="text-violet-500" />
              IOC database
            </li>
            <li className="flex items-center gap-2">
              <Bell size={11} className="text-fuchsia-500" />
              Slack #soc
            </li>
          </ul>
        </div>
      </aside>
    </div>
  )
}

function Metric({ label, value, icon: Icon, tone }: { label: string; value: string; icon: LucideIcon; tone?: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="inline-flex items-center gap-2 text-muted-foreground">
        <Icon size={11} />
        {label}
      </span>
      <span className={cn('font-mono tabular-nums', tone)}>{value}</span>
    </div>
  )
}

function SuccessFailureChart({ history }: { history: number[] }) {
  const w = 600
  const h = 70
  const max = Math.max(...history, 1)
  const bw = w / history.length
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-16 w-full">
      {history.map((v, i) => {
        const bh = (v / max) * (h - 8)
        return (
          <rect
            key={i}
            x={i * bw + 1}
            y={h - bh - 4}
            width={bw - 2}
            height={bh}
            rx={1}
            className="fill-emerald-500/70"
          />
        )
      })}
    </svg>
  )
}

function RunStatusDot({ status }: { status: Run['status'] }) {
  const tone =
    status === 'success'
      ? 'bg-emerald-500'
      : status === 'failed'
        ? 'bg-red-500'
        : status === 'running'
          ? 'bg-sky-500 animate-pulse'
          : 'bg-muted-foreground'
  return <span className={cn('h-2 w-2 shrink-0 rounded-full', tone)} />
}

/* ─── Drawer: Canvas ───────────────────────────────────────────────────── */

function CanvasTab({ flow }: { flow: Flow }) {
  return (
    <div className="space-y-4 p-6">
      <div className="rounded-lg border border-border bg-card p-2">
        <div className="flex items-center justify-between border-b border-border pb-2 px-2">
          <div className="text-[10px] uppercase tracking-wider text-muted-foreground">
            Read-only preview
          </div>
          <Button size="sm" variant="outline">
            <Code2 size={12} className="mr-1.5" />
            Open in editor
          </Button>
        </div>
        <FlowCanvas flow={flow} />
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-5">
        {(['trigger', 'condition', 'action', 'notify', 'end'] as NodeKind[]).map((k) => {
          const meta = NODE_META[k]
          const Icon = meta.icon
          return (
            <div key={k} className="rounded-md border border-border bg-card px-3 py-2 text-[11px]">
              <div className="flex items-center gap-1.5">
                <Icon size={12} className={meta.tone} />
                <span className="font-medium capitalize">{k}</span>
              </div>
              <div className="mt-0.5 text-[10px] text-muted-foreground">
                {flow.nodes.filter((n) => n.kind === k).length} in this flow
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function FlowCanvas({ flow }: { flow: Flow }) {
  // Compute layered layout: BFS from trigger, group by depth.
  const depth = new Map<string, number>()
  const trig = flow.nodes.find((n) => n.kind === 'trigger')
  if (trig) depth.set(trig.id, 0)
  let changed = true
  while (changed) {
    changed = false
    for (const e of flow.edges) {
      const f = depth.get(e.from)
      if (f === undefined) continue
      const cur = depth.get(e.to)
      if (cur === undefined || cur < f + 1) {
        depth.set(e.to, f + 1)
        changed = true
      }
    }
  }
  const maxDepth = Math.max(...Array.from(depth.values()), 0)
  const groups: string[][] = Array.from({ length: maxDepth + 1 }, () => [])
  flow.nodes.forEach((n) => {
    const d = depth.get(n.id)
    if (d !== undefined) groups[d].push(n.id)
  })

  const colW = 180
  const rowH = 70
  const padX = 30
  const padY = 30
  const w = padX * 2 + (maxDepth + 1) * colW
  const maxRows = Math.max(...groups.map((g) => g.length), 1)
  const h = padY * 2 + maxRows * rowH

  const pos = new Map<string, { x: number; y: number }>()
  groups.forEach((ids, d) => {
    const offset = (maxRows - ids.length) / 2
    ids.forEach((id, r) => {
      pos.set(id, {
        x: padX + d * colW + colW / 2,
        y: padY + (offset + r) * rowH + rowH / 2,
      })
    })
  })

  return (
    <div className="overflow-x-auto p-4">
      <svg viewBox={`0 0 ${w} ${h}`} className="block" style={{ minWidth: w, height: h }}>
        <defs>
          <pattern id="grid-canvas" width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M 20 0 L 0 0 0 20" className="fill-none stroke-border/40" strokeWidth="0.5" />
          </pattern>
          <marker id="arrow" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
            <path d="M0,0 L0,6 L6,3 Z" className="fill-muted-foreground/60" />
          </marker>
        </defs>
        <rect width={w} height={h} fill="url(#grid-canvas)" />

        {/* Edges */}
        {flow.edges.map((e, i) => {
          const a = pos.get(e.from)
          const b = pos.get(e.to)
          if (!a || !b) return null
          const cx = (a.x + b.x) / 2
          return (
            <g key={i}>
              <path
                d={`M ${a.x + 70} ${a.y} C ${cx} ${a.y}, ${cx} ${b.y}, ${b.x - 70} ${b.y}`}
                fill="none"
                className="stroke-muted-foreground/50"
                strokeWidth="1.5"
                markerEnd="url(#arrow)"
              />
              {e.label && (
                <text x={cx} y={(a.y + b.y) / 2 - 4} textAnchor="middle" fontSize="9" className="fill-muted-foreground">
                  {e.label}
                </text>
              )}
            </g>
          )
        })}

        {/* Nodes */}
        {flow.nodes.map((n) => {
          const p = pos.get(n.id)
          if (!p) return null
          const meta = NODE_META[n.kind]
          const Icon = n.icon ?? meta.icon
          const fill =
            n.kind === 'trigger'
              ? 'rgb(249 115 22)'
              : n.kind === 'condition'
                ? 'rgb(245 158 11)'
                : n.kind === 'action'
                  ? 'rgb(14 165 233)'
                  : n.kind === 'notify'
                    ? 'rgb(168 85 247)'
                    : 'rgb(16 185 129)'
          return (
            <g key={n.id} transform={`translate(${p.x - 70}, ${p.y - 22})`}>
              <rect
                width="140"
                height="44"
                rx="6"
                className="fill-card stroke-border"
                strokeWidth="1.25"
              />
              <rect width="3" height="44" rx="1.5" fill={fill} />
              <foreignObject x="10" y="6" width="124" height="32">
                <div className="flex h-full items-center gap-2">
                  <Icon size={13} className={meta.tone} />
                  <div className="min-w-0 leading-tight">
                    <div className="truncate text-[11px] font-medium text-foreground">{n.label}</div>
                    {n.detail && <div className="truncate text-[9px] text-muted-foreground">{n.detail}</div>}
                  </div>
                </div>
              </foreignObject>
            </g>
          )
        })}
      </svg>
    </div>
  )
}

/* ─── Drawer: Runs ─────────────────────────────────────────────────────── */

function RunsTab({ flow }: { flow: Flow }) {
  return (
    <div className="p-6">
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <div className="grid grid-cols-[80px_120px_1fr_90px_110px] gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div>Status</div>
          <div>Run ID</div>
          <div>Triggered by</div>
          <div className="text-right">Duration</div>
          <div>Time</div>
        </div>
        {flow.recentRuns.map((r) => (
          <div key={r.id} className="grid cursor-pointer grid-cols-[80px_120px_1fr_90px_110px] gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0">
            <div className="flex items-center gap-1.5">
              <RunStatusDot status={r.status} />
              <span className="capitalize">{r.status}</span>
            </div>
            <div className="font-mono text-muted-foreground">{r.id}</div>
            <div className="truncate">{r.triggeredBy}</div>
            <div className="text-right font-mono tabular-nums">{r.duration}s</div>
            <div className="font-mono text-muted-foreground">{relativeTime(r.ts)}</div>
          </div>
        ))}
        {flow.recentRuns.length === 0 && (
          <div className="px-4 py-12 text-center text-xs italic text-muted-foreground">No runs yet.</div>
        )}
      </div>
    </div>
  )
}

/* ─── Drawer: Audit ────────────────────────────────────────────────────── */

function AuditTab({ flow }: { flow: Flow }) {
  const events = [
    { ts: hr(8), who: flow.owner.name, what: 'updated condition node "svc account?"', kind: 'edit' as const },
    { ts: hr(48), who: flow.owner.name, what: 'enabled flow', kind: 'state' as const },
    { ts: days(3), who: 'Mara Thomas', what: 'reviewed and approved v3', kind: 'review' as const },
    { ts: days(7), who: flow.owner.name, what: 'created flow', kind: 'create' as const },
  ]
  const tone: Record<string, string> = {
    edit: 'bg-sky-500',
    state: 'bg-amber-500',
    review: 'bg-emerald-500',
    create: 'bg-violet-500',
  }
  return (
    <div className="p-6">
      <ol className="relative space-y-4 border-l border-border pl-6">
        {events.map((e, i) => (
          <li key={i} className="relative">
            <span className={cn('absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-card', tone[e.kind])} />
            <div className="text-xs">
              <span className="font-medium">{e.who}</span> {e.what}
            </div>
            <div className="text-[10px] text-muted-foreground">{absTimestamp(e.ts)}</div>
          </li>
        ))}
      </ol>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-2 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

function KeyValueCard({ title, rows }: { title: string; rows: [string, React.ReactNode][] }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      <dl className="grid grid-cols-[100px_1fr] gap-y-2 text-xs">
        {rows.map(([k, v]) => (
          <div key={k} className="contents">
            <dt className="text-muted-foreground">{k}</dt>
            <dd className="break-words">{v}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.round(h / 24)}d ago`
}

function absTimestamp(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
