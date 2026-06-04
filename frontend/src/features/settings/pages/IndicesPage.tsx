import { useMemo, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  Boxes,
  Calendar,
  ChevronRight,
  Copy,
  Database,
  Flame,
  HardDriveDownload,
  LayoutGrid,
  Lock,
  MoreHorizontal,
  Plus,
  RefreshCw,
  Search,
  Snowflake,
  Star,
  Thermometer,
  Trash2,
  X,
  Zap,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Tier = 'hot' | 'warm' | 'cold' | 'frozen'
type IndexStatus = 'open' | 'closed' | 'reindexing' | 'error'
type FieldType = 'date' | 'keyword' | 'text' | 'ip' | 'number' | 'boolean' | 'geo' | 'object'

interface DataView {
  id: string
  pattern: string
  description: string
  integrations: string[]
  timeField: string
  fieldCount: number
  docs: number
  bytes: number
  indices: number
  lastUpdate: string
  starred?: boolean
  systemOwner?: boolean
  fields: { name: string; type: FieldType; sample?: string }[]
}

interface Idx {
  id: string
  name: string
  pattern: string // which data view it belongs to
  tier: Tier
  status: IndexStatus
  docs: number
  bytes: number
  shards: number
  replicas: number
  createdAt: string
  closed?: boolean
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-02T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()
const GB = 1024 ** 3

const VIEWS: DataView[] = [
  {
    id: 'v-logs-windows',
    pattern: 'logs-windows-*',
    description: 'Sysmon, Security, Application, and File Audit events from Windows agents.',
    integrations: ['Windows agent'],
    timeField: '@timestamp',
    fieldCount: 142,
    docs: 18_220_412,
    bytes: 220 * GB,
    indices: 30,
    lastUpdate: min(0),
    starred: true,
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'event.code', type: 'number', sample: '4624' },
      { name: 'event.action', type: 'keyword', sample: 'Successful logon' },
      { name: 'host.name', type: 'keyword', sample: 'WIN-7F3K2L9Q8X' },
      { name: 'host.ip', type: 'ip', sample: '10.4.18.22' },
      { name: 'user.name', type: 'keyword', sample: 'svc_backup' },
      { name: 'process.name', type: 'keyword', sample: 'powershell.exe' },
      { name: 'process.command_line', type: 'text' },
      { name: 'kerberos.ticket.encryption_type', type: 'keyword' },
      { name: 'log.level', type: 'keyword', sample: 'info' },
    ],
  },
  {
    id: 'v-logs-linux',
    pattern: 'logs-linux-*',
    description: 'auditd, journald, syslog, and tailed /var/log files from Linux agents.',
    integrations: ['Linux agent'],
    timeField: '@timestamp',
    fieldCount: 88,
    docs: 28_104_882,
    bytes: 410 * GB,
    indices: 30,
    lastUpdate: min(0),
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'host.name', type: 'keyword', sample: 'fin-app-01' },
      { name: 'host.ip', type: 'ip', sample: '10.0.4.91' },
      { name: 'process.name', type: 'keyword', sample: 'sshd' },
      { name: 'log.level', type: 'keyword' },
      { name: 'message', type: 'text' },
    ],
  },
  {
    id: 'v-alerts',
    pattern: 'alerts-*',
    description: 'All security alerts emitted by detection rules.',
    integrations: ['All sources'],
    timeField: '@timestamp',
    fieldCount: 42,
    docs: 412_004,
    bytes: 18 * GB,
    indices: 90,
    lastUpdate: min(0),
    starred: true,
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'alert.id', type: 'keyword' },
      { name: 'alert.name', type: 'keyword' },
      { name: 'alert.severity', type: 'keyword', sample: 'critical' },
      { name: 'alert.status', type: 'keyword' },
      { name: 'rule.id', type: 'keyword' },
      { name: 'technique.id', type: 'keyword', sample: 'T1558.003' },
      { name: 'source.ip', type: 'ip' },
      { name: 'adversary.ip', type: 'ip' },
      { name: 'adversary.geo.country', type: 'keyword' },
    ],
  },
  {
    id: 'v-incidents',
    pattern: 'incidents-*',
    description: 'Incident records, status changes, and case history.',
    integrations: ['UTMStack'],
    timeField: '@timestamp',
    fieldCount: 28,
    docs: 8_204,
    bytes: 600 * 1024 * 1024,
    indices: 12,
    lastUpdate: hr(2),
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'incident.id', type: 'keyword' },
      { name: 'incident.severity', type: 'keyword' },
      { name: 'incident.status', type: 'keyword' },
      { name: 'incident.assignee', type: 'keyword' },
    ],
  },
  {
    id: 'v-audit',
    pattern: 'audit-*',
    description: 'Tamper-evident application audit log. Required for compliance reporting.',
    integrations: ['UTMStack'],
    timeField: '@timestamp',
    fieldCount: 18,
    docs: 2_204_120,
    bytes: 4 * GB,
    indices: 365,
    lastUpdate: min(2),
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'audit.action', type: 'keyword', sample: 'auth.login' },
      { name: 'audit.actor', type: 'keyword' },
      { name: 'audit.resource_type', type: 'keyword' },
      { name: 'audit.outcome', type: 'keyword' },
    ],
  },
  {
    id: 'v-firewall',
    pattern: 'firewall-*',
    description: 'pfSense, FortiGate, Palo Alto, SonicWall, Cisco ASA, and Meraki traffic.',
    integrations: ['pfSense', 'Palo Alto', 'FortiGate', 'Cisco ASA', 'Meraki'],
    timeField: '@timestamp',
    fieldCount: 67,
    docs: 380_220_044,
    bytes: 1_200 * GB,
    indices: 90,
    lastUpdate: min(0),
    starred: true,
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'source.ip', type: 'ip' },
      { name: 'destination.ip', type: 'ip' },
      { name: 'destination.port', type: 'number' },
      { name: 'network.bytes', type: 'number' },
      { name: 'network.transport', type: 'keyword' },
      { name: 'event.outcome', type: 'keyword' },
    ],
  },
  {
    id: 'v-cloud-aws',
    pattern: 'cloud-aws-*',
    description: 'CloudTrail, GuardDuty, VPC Flow, IAM, RDS audit, ECS, and Lambda.',
    integrations: ['AWS'],
    timeField: '@timestamp',
    fieldCount: 124,
    docs: 84_220_044,
    bytes: 480 * GB,
    indices: 60,
    lastUpdate: min(0),
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'cloud.provider', type: 'keyword', sample: 'aws' },
      { name: 'cloud.account.id', type: 'keyword' },
      { name: 'event.action', type: 'keyword', sample: 'ConsoleLogin' },
    ],
  },
  {
    id: 'v-cloud-m365',
    pattern: 'cloud-m365-*',
    description: 'M365 Unified Audit Log: Exchange, SharePoint, Teams, OneDrive, Defender.',
    integrations: ['Microsoft 365'],
    timeField: '@timestamp',
    fieldCount: 96,
    docs: 31_004_888,
    bytes: 220 * GB,
    indices: 60,
    lastUpdate: min(2),
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'user.email', type: 'keyword' },
      { name: 'event.action', type: 'keyword' },
      { name: 'object.path', type: 'keyword' },
    ],
  },
  {
    id: 'v-netflow',
    pattern: 'network-flows-*',
    description: 'NetFlow v5/v9, IPFIX, and sFlow records.',
    integrations: ['NetFlow'],
    timeField: '@timestamp',
    fieldCount: 32,
    docs: 380_222_004,
    bytes: 920 * GB,
    indices: 30,
    lastUpdate: min(0),
    systemOwner: true,
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'source.ip', type: 'ip' },
      { name: 'destination.ip', type: 'ip' },
      { name: 'network.bytes', type: 'number' },
      { name: 'network.packets', type: 'number' },
    ],
  },
  {
    id: 'v-custom',
    pattern: 'custom-app-*',
    description: 'Custom application logs ingested via the JSON intake.',
    integrations: ['Custom syslog / JSON'],
    timeField: '@timestamp',
    fieldCount: 12,
    docs: 1_204_882,
    bytes: 8 * GB,
    indices: 14,
    lastUpdate: min(8),
    fields: [
      { name: '@timestamp', type: 'date' },
      { name: 'service.name', type: 'keyword' },
      { name: 'level', type: 'keyword' },
      { name: 'message', type: 'text' },
    ],
  },
]

const INDICES: Idx[] = [
  // Windows logs — 30 daily indices
  { id: 'i-w-1', name: 'logs-windows-2026.05.02', pattern: 'logs-windows-*', tier: 'hot', status: 'open', docs: 4_204_412, bytes: 18 * GB, shards: 3, replicas: 1, createdAt: hr(8) },
  { id: 'i-w-2', name: 'logs-windows-2026.05.01', pattern: 'logs-windows-*', tier: 'hot', status: 'open', docs: 8_204_120, bytes: 24 * GB, shards: 3, replicas: 1, createdAt: hr(32) },
  { id: 'i-w-3', name: 'logs-windows-2026.04.30', pattern: 'logs-windows-*', tier: 'hot', status: 'open', docs: 7_812_004, bytes: 22 * GB, shards: 3, replicas: 1, createdAt: hr(56) },
  { id: 'i-w-old', name: 'logs-windows-2026.04.01', pattern: 'logs-windows-*', tier: 'warm', status: 'open', docs: 9_104_004, bytes: 28 * GB, shards: 1, replicas: 0, createdAt: days(31) },
  { id: 'i-w-cold', name: 'logs-windows-2026.02.15', pattern: 'logs-windows-*', tier: 'cold', status: 'open', docs: 6_204_004, bytes: 18 * GB, shards: 1, replicas: 0, createdAt: days(76) },

  // Linux logs
  { id: 'i-l-1', name: 'logs-linux-2026.05.02', pattern: 'logs-linux-*', tier: 'hot', status: 'open', docs: 12_204_004, bytes: 32 * GB, shards: 3, replicas: 1, createdAt: hr(8) },
  { id: 'i-l-2', name: 'logs-linux-2026.05.01', pattern: 'logs-linux-*', tier: 'hot', status: 'open', docs: 18_104_882, bytes: 48 * GB, shards: 3, replicas: 1, createdAt: hr(32) },
  { id: 'i-l-old', name: 'logs-linux-2026.04.05', pattern: 'logs-linux-*', tier: 'warm', status: 'open', docs: 22_004_120, bytes: 56 * GB, shards: 1, replicas: 0, createdAt: days(27) },

  // Alerts
  { id: 'i-a-1', name: 'alerts-2026.05', pattern: 'alerts-*', tier: 'hot', status: 'open', docs: 12_482, bytes: 600 * 1024 * 1024, shards: 1, replicas: 1, createdAt: days(1) },
  { id: 'i-a-2', name: 'alerts-2026.04', pattern: 'alerts-*', tier: 'warm', status: 'open', docs: 84_204, bytes: 4 * GB, shards: 1, replicas: 1, createdAt: days(31) },
  { id: 'i-a-3', name: 'alerts-2025-q4', pattern: 'alerts-*', tier: 'cold', status: 'open', docs: 412_004, bytes: 18 * GB, shards: 1, replicas: 0, createdAt: days(120) },

  // Audit
  { id: 'i-au-1', name: 'audit-2026.05', pattern: 'audit-*', tier: 'hot', status: 'open', docs: 84_204, bytes: 200 * 1024 * 1024, shards: 1, replicas: 1, createdAt: days(1) },
  { id: 'i-au-2', name: 'audit-2025', pattern: 'audit-*', tier: 'cold', status: 'open', docs: 1_204_004, bytes: 2 * GB, shards: 1, replicas: 0, createdAt: days(120) },
  { id: 'i-au-3', name: 'audit-2024', pattern: 'audit-*', tier: 'frozen', status: 'closed', docs: 884_120, bytes: 1 * GB, shards: 1, replicas: 0, createdAt: days(450), closed: true },

  // Firewall
  { id: 'i-fw-1', name: 'firewall-2026.05.02', pattern: 'firewall-*', tier: 'hot', status: 'open', docs: 18_204_882, bytes: 64 * GB, shards: 5, replicas: 1, createdAt: hr(8) },
  { id: 'i-fw-2', name: 'firewall-2026.05.01', pattern: 'firewall-*', tier: 'hot', status: 'open', docs: 22_104_882, bytes: 72 * GB, shards: 5, replicas: 1, createdAt: hr(32) },
  { id: 'i-fw-old', name: 'firewall-2026.04.20', pattern: 'firewall-*', tier: 'warm', status: 'open', docs: 28_204_882, bytes: 88 * GB, shards: 1, replicas: 0, createdAt: days(12) },

  // Cloud AWS
  { id: 'i-aws-1', name: 'cloud-aws-2026.05', pattern: 'cloud-aws-*', tier: 'hot', status: 'open', docs: 14_204_882, bytes: 78 * GB, shards: 3, replicas: 1, createdAt: days(1) },
  { id: 'i-aws-2', name: 'cloud-aws-2026.04', pattern: 'cloud-aws-*', tier: 'warm', status: 'open', docs: 22_104_882, bytes: 142 * GB, shards: 1, replicas: 0, createdAt: days(31) },

  // M365
  { id: 'i-m-1', name: 'cloud-m365-2026.05', pattern: 'cloud-m365-*', tier: 'hot', status: 'open', docs: 6_504_882, bytes: 38 * GB, shards: 3, replicas: 1, createdAt: days(1) },
  { id: 'i-m-2', name: 'cloud-m365-2026.04', pattern: 'cloud-m365-*', tier: 'warm', status: 'reindexing', docs: 14_104_882, bytes: 88 * GB, shards: 1, replicas: 0, createdAt: days(31) },

  // NetFlow
  { id: 'i-nf-1', name: 'network-flows-2026.05.02', pattern: 'network-flows-*', tier: 'hot', status: 'open', docs: 88_204_882, bytes: 142 * GB, shards: 5, replicas: 0, createdAt: hr(8) },

  // Custom — broken
  { id: 'i-c-1', name: 'custom-app-2026.05', pattern: 'custom-app-*', tier: 'hot', status: 'error', docs: 0, bytes: 0, shards: 1, replicas: 0, createdAt: hr(2) },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const TIER_META: Record<Tier, { label: string; icon: LucideIcon; tone: string; bg: string }> = {
  hot: { label: 'Hot', icon: Flame, tone: 'text-rose-500', bg: 'bg-rose-500' },
  warm: { label: 'Warm', icon: Thermometer, tone: 'text-amber-500', bg: 'bg-amber-500' },
  cold: { label: 'Cold', icon: HardDriveDownload, tone: 'text-sky-500', bg: 'bg-sky-500' },
  frozen: { label: 'Frozen', icon: Snowflake, tone: 'text-violet-500', bg: 'bg-violet-500' },
}

const STATUS_META: Record<IndexStatus, { label: string; tone: string; dot: string }> = {
  open: { label: 'Open', tone: 'text-emerald-500', dot: 'bg-emerald-500' },
  closed: { label: 'Closed', tone: 'text-muted-foreground', dot: 'bg-muted-foreground' },
  reindexing: { label: 'Reindexing', tone: 'text-sky-500', dot: 'bg-sky-500 animate-pulse' },
  error: { label: 'Error', tone: 'text-red-500', dot: 'bg-red-500 animate-pulse' },
}

const FIELD_TYPE_TONE: Record<FieldType, string> = {
  date: 'bg-violet-500',
  keyword: 'bg-sky-500',
  text: 'bg-emerald-500',
  ip: 'bg-fuchsia-500',
  number: 'bg-amber-500',
  boolean: 'bg-orange-500',
  geo: 'bg-cyan-500',
  object: 'bg-pink-500',
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

type Tab = 'views' | 'indices'

export function IndicesPage() {
  const [tab, setTab] = useState<Tab>('views')
  const [search, setSearch] = useState('')
  const [openView, setOpenView] = useState<DataView | null>(null)

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header />

      <div className="mt-6">
        <ClusterOverviewCard />
      </div>

      <div className="mt-6">
        <TabsRow current={tab} onChange={setTab} />
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[280px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={tab === 'views' ? 'Search data views, integrations, fields…' : 'Search indices by name…'}
            className="h-9 pl-9"
          />
        </div>
        <button
          title="Refresh"
          className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <RefreshCw size={14} />
        </button>
        {tab === 'views' && (
          <Button size="sm">
            <Plus size={13} className="mr-1.5" />
            New data view
          </Button>
        )}
      </div>

      <div className="mt-4">
        {tab === 'views' && <DataViewsTable search={search} onOpen={setOpenView} />}
        {tab === 'indices' && <IndicesTable search={search} />}
      </div>

      {openView && <ViewDrawer view={openView} onClose={() => setOpenView(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header() {
  return (
    <header>
      <h1 className="flex items-center gap-2 text-xl font-semibold">
        <Boxes size={18} strokeWidth={1.75} />
        Indices &amp; data views
      </h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Data views are logical groupings of indices used for searching. Indices are the physical
        layer behind them.
      </p>
    </header>
  )
}

/* ─── Cluster overview (combined) ──────────────────────────────────────── */

function ClusterOverviewCard() {
  const totalBytes = INDICES.reduce((s, i) => s + i.bytes, 0)
  const totalDocs = INDICES.reduce((s, i) => s + i.docs, 0)
  const tiers: Tier[] = ['hot', 'warm', 'cold', 'frozen']
  const perTier = tiers.reduce<Record<Tier, number>>(
    (acc, t) => {
      acc[t] = INDICES.filter((i) => i.tier === t).reduce((s, i) => s + i.bytes, 0)
      return acc
    },
    { hot: 0, warm: 0, cold: 0, frozen: 0 }
  )
  const errors = INDICES.filter((i) => i.status === 'error').length
  const reindexing = INDICES.filter((i) => i.status === 'reindexing').length

  // Top data views by size
  const topViews = [...VIEWS].sort((a, b) => b.bytes - a.bytes).slice(0, 5)
  const maxViewBytes = topViews[0]?.bytes ?? 1

  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 lg:grid-cols-[1.4fr_1fr]">
        {/* Storage column */}
        <div className="border-b border-border p-6 lg:border-b-0 lg:border-r">
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Cluster storage
          </div>
          <div className="mt-1 flex flex-wrap items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{formatBytes(totalBytes)}</span>
            <span className="text-sm text-muted-foreground">
              across <span className="text-foreground">{INDICES.length}</span> indices ·{' '}
              <span className="text-foreground">{formatCount(totalDocs)}</span> docs
            </span>
          </div>

          <div className="mt-4 flex h-2.5 overflow-hidden rounded-full bg-muted">
            {tiers.map((t) => {
              const pct = (perTier[t] / totalBytes) * 100
              if (pct < 0.5) return null
              return (
                <div
                  key={t}
                  className={TIER_META[t].bg}
                  style={{ width: `${pct}%` }}
                  title={`${TIER_META[t].label}: ${formatBytes(perTier[t])}`}
                />
              )
            })}
          </div>

          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1.5 text-[11px]">
            {tiers.map((t) => {
              const meta = TIER_META[t]
              const Icon = meta.icon
              return (
                <span key={t} className="inline-flex items-center gap-1.5">
                  <Icon size={11} className={meta.tone} />
                  <span className={meta.tone}>{meta.label}</span>
                  <span className="font-mono text-muted-foreground">
                    {formatBytes(perTier[t])}
                  </span>
                </span>
              )
            })}
          </div>

          <div className="mt-5 border-t border-border pt-4">
            <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
              Top data views by size
            </div>
            <ul className="space-y-1.5">
              {topViews.map((v) => {
                const pct = (v.bytes / maxViewBytes) * 100
                return (
                  <li key={v.id} className="grid grid-cols-[1fr_120px_70px] items-center gap-3 text-xs">
                    <span className="truncate font-mono text-[11px]">{v.pattern}</span>
                    <div className="h-1 overflow-hidden rounded-full bg-muted">
                      <div className="h-full rounded-full bg-foreground/40" style={{ width: `${pct}%` }} />
                    </div>
                    <span className="text-right font-mono tabular-nums text-muted-foreground">
                      {formatBytes(v.bytes)}
                    </span>
                  </li>
                )
              })}
            </ul>
          </div>
        </div>

        {/* Health column */}
        <div className="p-6">
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Cluster health
          </div>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-2xl font-semibold text-emerald-500">Green</span>
            <span className="text-sm text-muted-foreground">OpenSearch reachable</span>
          </div>

          <div className="mt-4 grid grid-cols-2 gap-3">
            <HealthStat label="Nodes" value="5" sub="3 data · 1 master · 1 ingest" />
            <HealthStat label="Active shards" value="487" sub="98.2%" />
            <HealthStat label="Unassigned" value="9" sub="being reassigned" tone="text-amber-500" />
            <HealthStat label="Pending tasks" value="2" sub="all low priority" />
          </div>

          {(errors > 0 || reindexing > 0) && (
            <div className="mt-4 space-y-2">
              {errors > 0 && (
                <div className="flex items-start gap-2 rounded-md border border-red-500/30 bg-red-500/5 p-2.5 text-[11px]">
                  <AlertTriangle size={12} className="mt-0.5 shrink-0 text-red-500" />
                  <div>
                    <span className="font-medium text-foreground">
                      {errors} index in error state.
                    </span>{' '}
                    <button className="text-primary hover:underline">Investigate</button>
                  </div>
                </div>
              )}
              {reindexing > 0 && (
                <div className="flex items-start gap-2 rounded-md border border-sky-500/30 bg-sky-500/5 p-2.5 text-[11px]">
                  <Activity size={12} className="mt-0.5 shrink-0 text-sky-500" />
                  <div>
                    <span className="font-medium text-foreground">
                      {reindexing} index reindexing
                    </span>{' '}
                    — ETA ~12 minutes.
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function HealthStat({
  label,
  value,
  sub,
  tone,
}: {
  label: string
  value: string
  sub: string
  tone?: string
}) {
  return (
    <div className="rounded-md border border-border bg-background/40 p-3">
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-0.5 font-mono text-lg font-semibold tabular-nums', tone)}>{value}</div>
      <div className="text-[10px] text-muted-foreground">{sub}</div>
    </div>
  )
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

const TABS: { id: Tab; label: string; icon: LucideIcon; count: number }[] = [
  { id: 'views', label: 'Data views', icon: LayoutGrid, count: VIEWS.length },
  { id: 'indices', label: 'Indices', icon: Database, count: INDICES.length },
]

function TabsRow({ current, onChange }: { current: Tab; onChange: (t: Tab) => void }) {
  return (
    <div className="flex items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const active = current === t.id
        const Icon = t.icon
        return (
          <button
            key={t.id}
            onClick={() => onChange(t.id)}
            className={cn(
              'relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <Icon size={13} />
            {t.label}
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {t.count}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Data views table ─────────────────────────────────────────────────── */

const VIEW_COLS = '20px 1fr 200px 80px 100px 90px 110px 36px'

function DataViewsTable({
  search,
  onOpen,
}: {
  search: string
  onOpen: (v: DataView) => void
}) {
  const filtered = useMemo(
    () =>
      VIEWS.filter((v) =>
        search
          ? (v.pattern + v.description + v.integrations.join(' ')).toLowerCase().includes(search.toLowerCase())
          : true
      ),
    [search]
  )
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: VIEW_COLS }}
      >
        <div />
        <div>Pattern</div>
        <div>Sources</div>
        <div className="text-right">Fields</div>
        <div className="text-right">Docs</div>
        <div className="text-right">Size</div>
        <div className="text-right">Updated</div>
        <div />
      </div>
      {filtered.map((v) => (
        <ViewRow key={v.id} view={v} onOpen={() => onOpen(v)} />
      ))}
      {filtered.length === 0 && (
        <div className="px-6 py-16 text-center text-sm text-muted-foreground">No data views match.</div>
      )}
    </div>
  )
}

function ViewRow({ view: v, onOpen }: { view: DataView; onOpen: () => void }) {
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: VIEW_COLS }}
    >
      <span>
        {v.starred ? (
          <Star size={11} className="fill-amber-400 text-amber-400" />
        ) : (
          <span className="block h-1 w-1" />
        )}
      </span>
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate font-mono text-sm font-medium">{v.pattern}</span>
          {v.systemOwner && (
            <span className="inline-flex items-center gap-0.5 rounded bg-muted px-1.5 py-0.5 text-[9px] uppercase text-muted-foreground">
              <Lock size={9} />
              system
            </span>
          )}
        </div>
        <div className="truncate text-[11px] text-muted-foreground">{v.description}</div>
      </div>
      <div className="truncate text-[11px] text-muted-foreground">
        {v.integrations.slice(0, 2).join(', ')}
        {v.integrations.length > 2 && ` +${v.integrations.length - 2}`}
      </div>
      <div className="text-right font-mono tabular-nums">{v.fieldCount}</div>
      <div className="text-right font-mono tabular-nums">{formatCount(v.docs)}</div>
      <div className="text-right font-mono tabular-nums">{formatBytes(v.bytes)}</div>
      <div className="text-right font-mono text-[11px] text-muted-foreground">
        {relativeTime(v.lastUpdate)}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <ChevronRight size={14} className="text-muted-foreground" />
      </div>
    </div>
  )
}

/* ─── Indices table ────────────────────────────────────────────────────── */

const IDX_COLS = '24px 1fr 90px 110px 110px 90px 70px 36px'

function IndicesTable({ search }: { search: string }) {
  const filtered = useMemo(
    () =>
      INDICES.filter((i) =>
        search ? (i.name + i.pattern).toLowerCase().includes(search.toLowerCase()) : true
      ),
    [search]
  )
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: IDX_COLS }}
      >
        <div />
        <div>Index</div>
        <div>Tier</div>
        <div>Status</div>
        <div className="text-right">Docs</div>
        <div className="text-right">Size</div>
        <div className="text-right">Age</div>
        <div />
      </div>
      {filtered.map((i) => (
        <IndexRow key={i.id} idx={i} />
      ))}
      {filtered.length === 0 && (
        <div className="px-6 py-16 text-center text-sm text-muted-foreground">No indices match.</div>
      )}
    </div>
  )
}

function IndexRow({ idx: i }: { idx: Idx }) {
  const tier = TIER_META[i.tier]
  const TIcon = tier.icon
  const st = STATUS_META[i.status]
  return (
    <div
      className={cn(
        'group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0',
        i.status === 'closed' && 'opacity-60'
      )}
      style={{ gridTemplateColumns: IDX_COLS }}
    >
      <Database size={12} className="text-muted-foreground" />
      <div className="min-w-0">
        <div className="truncate font-mono text-sm">{i.name}</div>
        <div className="truncate text-[10px] text-muted-foreground">
          <span className="font-mono">{i.shards}p · {i.replicas}r</span>
          <span> · matches </span>
          <span className="font-mono">{i.pattern}</span>
        </div>
      </div>
      <div className="inline-flex items-center gap-1.5 text-[11px]">
        <TIcon size={11} className={tier.tone} />
        <span className={tier.tone}>{tier.label}</span>
      </div>
      <div>
        <span className="inline-flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>
      <div className="text-right font-mono tabular-nums">
        {i.docs > 0 ? formatCount(i.docs) : <span className="text-muted-foreground">—</span>}
      </div>
      <div className="text-right font-mono tabular-nums">
        {i.bytes > 0 ? formatBytes(i.bytes) : <span className="text-muted-foreground">—</span>}
      </div>
      <div className="text-right font-mono text-[11px] text-muted-foreground">
        {relativeTime(i.createdAt)}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button
          onClick={(e) => e.stopPropagation()}
          className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground hover:bg-background hover:text-foreground"
        >
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}

/* ─── View drawer ──────────────────────────────────────────────────────── */

function ViewDrawer({ view: v, onClose }: { view: DataView; onClose: () => void }) {
  const matchedIndices = useMemo(
    () => INDICES.filter((i) => i.pattern === v.pattern),
    [v.pattern]
  )
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <LayoutGrid size={11} />
                <span>Data view</span>
                {v.systemOwner && (
                  <>
                    <span>·</span>
                    <span className="inline-flex items-center gap-1">
                      <Lock size={10} />
                      system
                    </span>
                  </>
                )}
                <span>·</span>
                <span>updated {relativeTime(v.lastUpdate)}</span>
              </div>
              <div className="mt-1 flex items-center gap-2">
                {v.starred && <Star size={14} className="fill-amber-400 text-amber-400" />}
                <h2 className="truncate font-mono text-lg font-semibold">{v.pattern}</h2>
              </div>
              <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{v.description}</p>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span className="text-muted-foreground">Sources:</span>
                {v.integrations.map((s) => (
                  <span key={s} className="rounded-md bg-muted px-1.5 py-0.5">
                    {s}
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
              <Search size={13} className="mr-1.5" />
              Open in Log Explorer
            </Button>
            <Button size="sm" variant="outline">
              <Copy size={13} className="mr-1.5" />
              Duplicate
            </Button>
            {!v.systemOwner && (
              <Button size="sm" variant="outline" className="ml-auto text-red-500 hover:bg-red-500/10">
                <Trash2 size={13} className="mr-1.5" />
                Delete
              </Button>
            )}
          </div>
        </header>

        <div className="flex-1 space-y-5 overflow-y-auto bg-muted/20 p-6">
          <Section title="Stats">
            <div className="grid grid-cols-4 gap-3">
              <Stat label="Documents" value={formatCount(v.docs)} icon={Zap} />
              <Stat label="Size" value={formatBytes(v.bytes)} icon={Database} />
              <Stat label="Indices" value={String(v.indices)} icon={Boxes} />
              <Stat label="Fields" value={String(v.fieldCount)} icon={LayoutGrid} />
            </div>
          </Section>

          <Section
            title="Configuration"
            subtitle="Settings used when querying this data view."
          >
            <dl className="grid grid-cols-[160px_1fr] gap-y-2 text-xs">
              <Row k="Index pattern"><span className="font-mono">{v.pattern}</span></Row>
              <Row k="Time field"><span className="font-mono">{v.timeField}</span></Row>
              <Row k="Default columns">
                <span className="font-mono text-[11px] text-muted-foreground">@timestamp, host.name, message</span>
              </Row>
              <Row k="Refresh interval">5 minutes</Row>
            </dl>
          </Section>

          <Section
            title="Fields"
            subtitle={`First ${v.fields.length} of ${v.fieldCount}. Click a field for stats and top values.`}
          >
            <div className="overflow-hidden rounded-md border border-border bg-card">
              <div className="grid grid-cols-[24px_1fr_100px_1fr] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
                <div />
                <div>Name</div>
                <div>Type</div>
                <div>Sample</div>
              </div>
              {v.fields.map((f) => (
                <div
                  key={f.name}
                  className="grid grid-cols-[24px_1fr_100px_1fr] items-center gap-3 border-b border-border/60 px-3 py-1.5 text-xs last:border-b-0 hover:bg-muted/30"
                >
                  <span className={cn('h-1.5 w-1.5 rounded-full', FIELD_TYPE_TONE[f.type])} />
                  <span className="font-mono text-[11px]">{f.name}</span>
                  <span className="font-mono text-[10px] uppercase text-muted-foreground">{f.type}</span>
                  <span className="truncate font-mono text-[11px] text-muted-foreground">
                    {f.sample ?? '—'}
                  </span>
                </div>
              ))}
            </div>
          </Section>

          <Section
            title={`Backing indices · ${matchedIndices.length}`}
            subtitle="Physical indices that match this pattern."
          >
            <div className="overflow-hidden rounded-md border border-border bg-card">
              <div className="grid grid-cols-[1fr_80px_90px_90px] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
                <div>Name</div>
                <div>Tier</div>
                <div className="text-right">Docs</div>
                <div className="text-right">Size</div>
              </div>
              {matchedIndices.map((i) => {
                const tier = TIER_META[i.tier]
                const TIcon = tier.icon
                return (
                  <div
                    key={i.id}
                    className="grid grid-cols-[1fr_80px_90px_90px] items-center gap-3 border-b border-border/60 px-3 py-1.5 text-xs last:border-b-0 hover:bg-muted/30"
                  >
                    <span className="truncate font-mono text-[11px]">{i.name}</span>
                    <span className="inline-flex items-center gap-1.5 text-[11px]">
                      <TIcon size={10} className={tier.tone} />
                      <span className={tier.tone}>{tier.label}</span>
                    </span>
                    <span className="text-right font-mono tabular-nums">{formatCount(i.docs)}</span>
                    <span className="text-right font-mono tabular-nums">{formatBytes(i.bytes)}</span>
                  </div>
                )
              })}
              {matchedIndices.length === 0 && (
                <div className="px-3 py-6 text-center text-[11px] text-muted-foreground">
                  No physical indices yet — pattern is empty.
                </div>
              )}
            </div>
          </Section>
        </div>
      </div>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children: React.ReactNode
}) {
  return (
    <section className="rounded-xl border border-border bg-card p-5">
      <div className="mb-3">
        <h2 className="text-sm font-semibold">{title}</h2>
        {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
      </div>
      {children}
    </section>
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

function Stat({ label, value, icon: Icon }: { label: string; value: string; icon: LucideIcon }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-3 py-2.5">
      <div className="flex items-center gap-1.5 text-[10px] uppercase tracking-wider text-muted-foreground">
        <Icon size={10} />
        {label}
      </div>
      <div className="mt-1 font-mono text-base font-semibold tabular-nums">{value}</div>
    </div>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  if (!iso) return '—'
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h`
  return `${Math.round(h / 24)}d`
}

function formatCount(n: number) {
  if (n >= 1_000_000_000) return `${(n / 1_000_000_000).toFixed(1)}B`
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return n.toString()
}

function formatBytes(bytes: number) {
  if (bytes >= 1024 ** 4) return `${(bytes / 1024 ** 4).toFixed(1)} TB`
  if (bytes >= 1024 ** 3) return `${(bytes / 1024 ** 3).toFixed(1)} GB`
  if (bytes >= 1024 ** 2) return `${(bytes / 1024 ** 2).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${bytes} B`
}

// Suppress unused import warnings
void Calendar
