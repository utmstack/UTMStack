import { useMemo, useState } from 'react'
import {
  AlertTriangle,
  Calendar,
  Check,
  Copy,
  Crown,
  Eye,
  Globe2,
  Key,
  KeyRound,
  ListFilter,
  Lock,
  MoreHorizontal,
  Network,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  Shield,
  Trash2,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Scope = 'org' | 'workspace'
type Role = 'reader' | 'maintainer' | 'admin'
type KeyStatus = 'active' | 'expiring' | 'expired' | 'revoked'

interface ApiKey {
  id: string
  name: string
  prefix: string // first 6 chars of the actual token, used for identification
  scope: Scope
  workspaceId?: string
  role: Role
  allowedIps: string[]
  expiresAt?: string
  createdAt: string
  createdBy: string
  lastUsedAt?: string
  status: KeyStatus
  requests24h: number
  description?: string
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-02T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()
const future = (d: number) => new Date(NOW + d * 86_400_000).toISOString()

const WORKSPACES_MOCK = [
  { id: 'ws-acme', name: 'Acme Corp' },
  { id: 'ws-globex', name: 'Globex Industries' },
  { id: 'ws-initech', name: 'Initech' },
]

const wsName = (id?: string) => (id ? WORKSPACES_MOCK.find((w) => w.id === id)?.name ?? id : '—')

const KEYS: ApiKey[] = [
  {
    id: 'k-001',
    name: 'Terraform CI',
    prefix: 'utm_4f',
    scope: 'org',
    role: 'admin',
    allowedIps: ['198.51.100.0/24'],
    expiresAt: future(45),
    createdAt: days(45),
    createdBy: 'jsmith',
    lastUsedAt: min(8),
    status: 'active',
    requests24h: 142,
    description: 'CI runner that provisions integrations and rules via the REST API.',
  },
  {
    id: 'k-002',
    name: 'Splunk export pipeline',
    prefix: 'utm_a1',
    scope: 'org',
    role: 'reader',
    allowedIps: ['10.0.0.0/8'],
    expiresAt: future(180),
    createdAt: days(120),
    createdBy: 'mthomas',
    lastUsedAt: min(2),
    status: 'active',
    requests24h: 4_812,
    description: 'Streams alerts and incidents into our Splunk pipeline for archival.',
  },
  {
    id: 'k-003',
    name: 'Slack bot · alerts',
    prefix: 'utm_9c',
    scope: 'workspace',
    workspaceId: 'ws-acme',
    role: 'reader',
    allowedIps: [],
    expiresAt: future(5),
    createdAt: days(360),
    createdBy: 'jsmith',
    lastUsedAt: min(0),
    status: 'expiring',
    requests24h: 312,
    description: 'Posts critical alerts into #soc on Slack.',
  },
  {
    id: 'k-004',
    name: 'Datadog metrics scraper',
    prefix: 'utm_d4',
    scope: 'org',
    role: 'reader',
    allowedIps: ['54.236.0.0/15'],
    expiresAt: undefined,
    createdAt: days(540),
    createdBy: 'crodriguez',
    lastUsedAt: min(0),
    status: 'active',
    requests24h: 1_204,
    description: 'Polls health metrics every minute for the Datadog dashboard.',
  },
  {
    id: 'k-005',
    name: 'Quarterly compliance export',
    prefix: 'utm_b8',
    scope: 'workspace',
    workspaceId: 'ws-acme',
    role: 'reader',
    allowedIps: ['203.0.113.50/32'],
    expiresAt: future(90),
    createdAt: days(80),
    createdBy: 'aluna',
    lastUsedAt: days(7),
    status: 'active',
    requests24h: 0,
    description: 'Auditor read-only access for Q2 compliance reports.',
  },
  {
    id: 'k-006',
    name: 'Old legacy webhook',
    prefix: 'utm_3e',
    scope: 'org',
    role: 'maintainer',
    allowedIps: [],
    expiresAt: days(2),
    createdAt: days(720),
    createdBy: 'system',
    lastUsedAt: days(45),
    status: 'expired',
    requests24h: 0,
    description: 'Legacy webhook integration. Has been replaced by the new pipeline.',
  },
  {
    id: 'k-007',
    name: 'Compromised — pentest',
    prefix: 'utm_7a',
    scope: 'org',
    role: 'admin',
    allowedIps: [],
    expiresAt: undefined,
    createdAt: days(180),
    createdBy: 'tferguson',
    lastUsedAt: days(60),
    status: 'revoked',
    requests24h: 0,
    description: 'Revoked after the August 2025 pentest finding (KEY-LEAK-2025-08-14).',
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const STATUS_META: Record<KeyStatus, { label: string; tone: string; pill: string; dot: string }> = {
  active: { label: 'Active', tone: 'text-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
  expiring: { label: 'Expiring soon', tone: 'text-amber-500', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
  expired: { label: 'Expired', tone: 'text-red-500', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500' },
  revoked: { label: 'Revoked', tone: 'text-muted-foreground', pill: 'bg-muted text-muted-foreground ring-border', dot: 'bg-muted-foreground' },
}

const ROLE_META: Record<Role, { label: string; icon: LucideIcon; tone: string; pill: string }> = {
  admin: { label: 'Admin', icon: Crown, tone: 'text-amber-500', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300' },
  maintainer: { label: 'Maintainer', icon: Shield, tone: 'text-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300' },
  reader: { label: 'Reader', icon: Eye, tone: 'text-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300' },
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

type Tab = 'all' | 'active' | 'expiring' | 'inactive' | 'revoked'

const TABS: { id: Tab; label: string; predicate: (k: ApiKey) => boolean }[] = [
  { id: 'all', label: 'All', predicate: () => true },
  { id: 'active', label: 'Active', predicate: (k) => k.status === 'active' },
  { id: 'expiring', label: 'Expiring soon', predicate: (k) => k.status === 'expiring' },
  { id: 'inactive', label: 'Unused 30+ days', predicate: (k) => !k.lastUsedAt || (NOW - new Date(k.lastUsedAt).getTime()) > 30 * 86_400_000 },
  { id: 'revoked', label: 'Revoked / expired', predicate: (k) => k.status === 'revoked' || k.status === 'expired' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function ApiKeysPage() {
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [openKey, setOpenKey] = useState<ApiKey | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [generated, setGenerated] = useState<{ name: string; token: string } | null>(null)

  const filtered = useMemo(() => {
    const t = TABS.find((x) => x.id === tab) ?? TABS[0]
    return KEYS.filter(t.predicate).filter((k) =>
      search ? (k.name + k.prefix + (k.description ?? '')).toLowerCase().includes(search.toLowerCase()) : true
    )
  }, [tab, search])

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header onCreate={() => setCreateOpen(true)} />

      <div className="mt-6">
        <StatsStrip />
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
            placeholder="Search by name, prefix, description…"
            className="h-9 pl-9"
          />
        </div>
        <Button variant="outline" size="sm">
          <ListFilter size={14} className="mr-2" />
          Filters
        </Button>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <ListHeader />
        {filtered.map((k) => (
          <KeyRow key={k.id} apiKey={k} onOpen={() => setOpenKey(k)} />
        ))}
        {filtered.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            No API keys match this view.
          </div>
        )}
      </div>

      {openKey && <KeyDrawer apiKey={openKey} onClose={() => setOpenKey(null)} />}
      {createOpen && (
        <CreateDialog
          onClose={() => setCreateOpen(false)}
          onCreated={(name, token) => {
            setCreateOpen(false)
            setGenerated({ name, token })
          }}
        />
      )}
      {generated && <GeneratedDialog name={generated.name} token={generated.token} onClose={() => setGenerated(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ onCreate }: { onCreate: () => void }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <KeyRound size={18} strokeWidth={1.75} />
          API keys
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Programmatic access to UTMStack. Each key is bound to a role and can be scoped to a
          workspace and an IP allowlist.
        </p>
      </div>
      <Button size="sm" onClick={onCreate}>
        <Plus size={14} className="mr-2" />
        New API key
      </Button>
    </header>
  )
}

/* ─── Stats strip ──────────────────────────────────────────────────────── */

function StatsStrip() {
  const active = KEYS.filter((k) => k.status === 'active').length
  const expiring = KEYS.filter((k) => k.status === 'expiring').length
  const inactive = KEYS.filter(
    (k) => !k.lastUsedAt || (NOW - new Date(k.lastUsedAt).getTime()) > 30 * 86_400_000
  ).length
  const totalReqs = KEYS.reduce((s, k) => s + k.requests24h, 0)
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-4 sm:divide-x sm:divide-y-0">
        <StripStat label="Active" value={String(active)} sub="Currently accepting requests" tone="text-emerald-500" />
        <StripStat label="Expiring soon" value={String(expiring)} sub="Within the next 7 days" tone={expiring > 0 ? 'text-amber-500' : 'text-foreground'} />
        <StripStat label="Unused 30+ days" value={String(inactive)} sub="Consider revoking" tone={inactive > 0 ? 'text-muted-foreground' : 'text-foreground'} />
        <StripStat label="Requests · 24h" value={totalReqs.toLocaleString()} sub="Across all active keys" tone="text-foreground" />
      </div>
    </section>
  )
}

function StripStat({
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
    <div className="px-5 py-4">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-1 font-mono text-2xl font-semibold tabular-nums', tone)}>{value}</div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">{sub}</div>
    </div>
  )
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

function TabsRow({ current, onChange }: { current: Tab; onChange: (t: Tab) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const active = current === t.id
        const count = KEYS.filter(t.predicate).length
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
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {count}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── List ─────────────────────────────────────────────────────────────── */

const COLS = '4px 1fr 110px 130px 90px 110px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <div />
      <div>Key</div>
      <div>Role</div>
      <div>Scope</div>
      <div className="text-right">24h</div>
      <div>Last used</div>
      <div>Expires</div>
      <div />
    </div>
  )
}

function KeyRow({ apiKey: k, onOpen }: { apiKey: ApiKey; onOpen: () => void }) {
  const st = STATUS_META[k.status]
  const role = ROLE_META[k.role]
  const RoleIcon = role.icon
  return (
    <div
      onClick={onOpen}
      className={cn(
        'group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs hover:bg-muted/40 last:border-b-0',
        (k.status === 'revoked' || k.status === 'expired') && 'opacity-60'
      )}
      style={{ gridTemplateColumns: COLS }}
    >
      <span className={cn('h-7 w-1 rounded-full', st.dot)} />
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{k.name}</span>
          <span
            className={cn(
              'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
              st.pill
            )}
          >
            {k.status === 'expired' || k.status === 'revoked' ? null : (
              <span className={cn('h-1 w-1 rounded-full', st.dot)} />
            )}
            {st.label}
          </span>
        </div>
        <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">{k.prefix}…</code>
          {k.allowedIps.length > 0 && (
            <>
              <span>·</span>
              <span className="inline-flex items-center gap-1">
                <Network size={10} />
                {k.allowedIps.length} IP{k.allowedIps.length === 1 ? '' : 's'}
              </span>
            </>
          )}
          {k.allowedIps.length === 0 && k.status === 'active' && (
            <>
              <span>·</span>
              <span className="text-amber-500">no IP restriction</span>
            </>
          )}
        </div>
      </div>
      <div>
        <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', role.pill)}>
          <RoleIcon size={10} />
          {role.label}
        </span>
      </div>
      <div className="text-[11px] text-muted-foreground">
        {k.scope === 'org' ? (
          <span className="inline-flex items-center gap-1">
            <Globe2 size={11} />
            Organization
          </span>
        ) : (
          <span className="truncate">{wsName(k.workspaceId)}</span>
        )}
      </div>
      <div className={cn('text-right tabular-nums', k.requests24h > 0 ? 'text-foreground' : 'text-muted-foreground')}>
        {k.requests24h > 0 ? k.requests24h.toLocaleString() : '—'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {k.lastUsedAt ? relativeTime(k.lastUsedAt) : 'never'}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {k.expiresAt ? (
          <>
            {relativeFromOrTo(k.expiresAt)}
            <div className={cn('text-[10px]', k.status === 'expiring' && 'text-amber-500', k.status === 'expired' && 'text-red-500')}>
              {k.status === 'expiring' && 'in danger'}
              {k.status === 'expired' && 'expired'}
            </div>
          </>
        ) : (
          'never'
        )}
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

/* ─── Drawer ───────────────────────────────────────────────────────────── */

function KeyDrawer({ apiKey: k, onClose }: { apiKey: ApiKey; onClose: () => void }) {
  const st = STATUS_META[k.status]
  const role = ROLE_META[k.role]
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
                <Key size={11} />
                <span className="font-mono">{k.id}</span>
                <span>·</span>
                <span>created {relativeTime(k.createdAt)} by <span className="font-mono">{k.createdBy}</span></span>
              </div>
              <h2 className="mt-1 truncate text-lg font-semibold">{k.name}</h2>
              {k.description && (
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{k.description}</p>
              )}
              <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', st.pill)}>
                  <span className={cn('h-1 w-1 rounded-full', st.dot)} />
                  {st.label}
                </span>
                <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', role.pill)}>
                  <role.icon size={10} />
                  {role.label}
                </span>
                <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                  {k.scope === 'org' ? (
                    <>
                      <Globe2 size={10} />
                      Organization
                    </>
                  ) : (
                    <>
                      <Lock size={10} />
                      {wsName(k.workspaceId)}
                    </>
                  )}
                </span>
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
            {k.status !== 'revoked' && k.status !== 'expired' && (
              <>
                <Button size="sm">
                  <Pencil size={13} className="mr-1.5" />
                  Edit
                </Button>
                <Button size="sm" variant="outline">
                  <RefreshCw size={13} className="mr-1.5" />
                  Rotate
                </Button>
                <Button size="sm" variant="outline" className="ml-auto text-red-500 hover:bg-red-500/10">
                  <Trash2 size={13} className="mr-1.5" />
                  Revoke
                </Button>
              </>
            )}
            {(k.status === 'revoked' || k.status === 'expired') && (
              <Button size="sm" variant="outline" className="text-red-500 hover:bg-red-500/10">
                <Trash2 size={13} className="mr-1.5" />
                Delete permanently
              </Button>
            )}
          </div>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto bg-muted/20 p-6">
          {/* Token preview */}
          <Section title="Token">
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border border-input bg-background/40 px-3 py-2 font-mono text-xs">
                {k.prefix}……………………………………………………………………{`(hidden after creation)`}
              </code>
              <Button size="sm" variant="outline">
                <RefreshCw size={13} className="mr-1.5" />
                Rotate
              </Button>
            </div>
            <p className="mt-2 text-[11px] text-muted-foreground">
              The full token value is shown only once at creation time. If lost, rotate it to get a
              new value. The first 6 characters always remain visible for identification.
            </p>
          </Section>

          {/* IP allowlist */}
          <Section title="IP allowlist">
            {k.allowedIps.length === 0 ? (
              <div className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-xs">
                <AlertTriangle size={13} className="mt-0.5 shrink-0 text-amber-500" />
                <div>
                  <span className="font-medium text-foreground">No IP restriction.</span> This key
                  can be used from any address. Consider restricting to specific IPs or CIDR ranges.
                </div>
              </div>
            ) : (
              <div className="space-y-1.5">
                {k.allowedIps.map((ip) => (
                  <div
                    key={ip}
                    className="flex items-center justify-between rounded-md border border-border bg-background/40 px-3 py-2 text-xs"
                  >
                    <span className="font-mono">{ip}</span>
                    <Network size={12} className="text-muted-foreground" />
                  </div>
                ))}
              </div>
            )}
          </Section>

          {/* Usage stats */}
          <Section title="Usage">
            <dl className="grid grid-cols-[160px_1fr] gap-y-2 text-xs">
              <Row k="Last used"><span className="font-mono">{k.lastUsedAt ? relativeTime(k.lastUsedAt) : 'never'}</span></Row>
              <Row k="Requests · 24h"><span className="font-mono tabular-nums">{k.requests24h.toLocaleString()}</span></Row>
              <Row k="Created"><span className="font-mono">{absTimestamp(k.createdAt)}</span></Row>
              {k.expiresAt && <Row k="Expires"><span className="font-mono">{absTimestamp(k.expiresAt)}</span></Row>}
            </dl>
          </Section>

          {/* Recent calls (mock) */}
          <Section title="Recent calls">
            <div className="overflow-hidden rounded-md border border-border bg-card">
              <div className="grid grid-cols-[120px_80px_1fr_80px] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
                <div>When</div>
                <div>Method</div>
                <div>Endpoint</div>
                <div className="text-right">Status</div>
              </div>
              {k.requests24h > 0 ? (
                [
                  { ts: min(2), method: 'GET', endpoint: '/api/v1/alerts?status=open&limit=50', status: 200 },
                  { ts: min(8), method: 'GET', endpoint: '/api/v1/alerts?status=open&limit=50', status: 200 },
                  { ts: min(14), method: 'POST', endpoint: '/api/v1/incidents/1142/actions/ack', status: 200 },
                  { ts: min(22), method: 'GET', endpoint: '/api/v1/datasources', status: 200 },
                  { ts: min(34), method: 'GET', endpoint: '/api/v1/alerts?status=open&limit=50', status: 200 },
                ].map((c, i) => (
                  <div
                    key={i}
                    className="grid grid-cols-[120px_80px_1fr_80px] items-center gap-3 border-b border-border/60 px-3 py-1.5 text-xs last:border-b-0"
                  >
                    <span className="font-mono text-[11px] text-muted-foreground">{relativeTime(c.ts)}</span>
                    <span className="font-mono text-[11px]">{c.method}</span>
                    <span className="truncate font-mono text-[11px]">{c.endpoint}</span>
                    <span className={cn('text-right font-mono text-[11px]', c.status >= 400 ? 'text-red-500' : 'text-emerald-500')}>
                      {c.status}
                    </span>
                  </div>
                ))
              ) : (
                <div className="px-3 py-6 text-center text-[11px] italic text-muted-foreground">
                  No requests recorded in the selected window.
                </div>
              )}
            </div>
          </Section>
        </div>
      </div>
    </div>
  )
}

/* ─── Create dialog ────────────────────────────────────────────────────── */

function CreateDialog({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated: (name: string, token: string) => void
}) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [scope, setScope] = useState<Scope>('org')
  const [workspaceId, setWorkspaceId] = useState<string>(WORKSPACES_MOCK[0].id)
  const [role, setRole] = useState<Role>('reader')
  const [ips, setIps] = useState('')
  const [expiry, setExpiry] = useState<'30d' | '90d' | '1y' | 'never'>('90d')

  const create = () => {
    // Generate a fake token
    const random = Array.from({ length: 32 }, () =>
      'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'.charAt(
        Math.floor(Math.random() * 62)
      )
    ).join('')
    const token = `utm_${random}`
    onCreated(name || 'Untitled key', token)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <KeyRound size={18} />
              New API key
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">
              We'll generate a secret token. You can only view the full value once — copy it now and
              store it securely.
            </p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="space-y-5 px-6 py-5">
          <Field label="Name" hint="A label so future-you remembers what this key is for.">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Terraform CI"
              className="h-9"
            />
          </Field>

          <Field label="Description (optional)">
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Used by GitHub Actions to provision integrations"
              className="h-9"
            />
          </Field>

          <Field label="Scope">
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => setScope('org')}
                className={cn(
                  'flex items-start gap-2.5 rounded-lg border p-3 text-left text-xs transition-colors',
                  scope === 'org' ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                )}
              >
                <Globe2 size={14} className={scope === 'org' ? 'text-primary' : 'text-muted-foreground'} />
                <div>
                  <div className={cn('font-semibold', scope === 'org' && 'text-primary')}>Organization</div>
                  <div className="text-[11px] text-muted-foreground">All workspaces.</div>
                </div>
              </button>
              <button
                onClick={() => setScope('workspace')}
                className={cn(
                  'flex items-start gap-2.5 rounded-lg border p-3 text-left text-xs transition-colors',
                  scope === 'workspace' ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                )}
              >
                <Lock size={14} className={scope === 'workspace' ? 'text-primary' : 'text-muted-foreground'} />
                <div>
                  <div className={cn('font-semibold', scope === 'workspace' && 'text-primary')}>Workspace</div>
                  <div className="text-[11px] text-muted-foreground">A single workspace.</div>
                </div>
              </button>
            </div>
            {scope === 'workspace' && (
              <select
                value={workspaceId}
                onChange={(e) => setWorkspaceId(e.target.value)}
                className="mt-2 h-9 w-full rounded-md border border-border bg-background/40 px-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
              >
                {WORKSPACES_MOCK.map((w) => (
                  <option key={w.id} value={w.id}>{w.name}</option>
                ))}
              </select>
            )}
          </Field>

          <Field label="Role">
            <div className="grid grid-cols-3 gap-2">
              {(['reader', 'maintainer', 'admin'] as Role[]).map((r) => {
                const meta = ROLE_META[r]
                const Icon = meta.icon
                const active = role === r
                return (
                  <button
                    key={r}
                    onClick={() => setRole(r)}
                    className={cn(
                      'flex flex-col items-start gap-1 rounded-lg border p-2.5 text-left text-xs transition-colors',
                      active ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                    )}
                  >
                    <Icon size={12} className={meta.tone} />
                    <span className={cn('font-semibold', meta.tone)}>{meta.label}</span>
                    <span className="text-[10px] text-muted-foreground">
                      {r === 'reader' && 'View only'}
                      {r === 'maintainer' && 'Read + ack alerts + run playbooks'}
                      {r === 'admin' && 'Full access (manage everything)'}
                    </span>
                  </button>
                )
              })}
            </div>
          </Field>

          <Field
            label="IP allowlist (optional)"
            hint="Comma-separated IPs or CIDR ranges. Leave empty to allow all."
          >
            <Input
              value={ips}
              onChange={(e) => setIps(e.target.value)}
              placeholder="198.51.100.0/24, 203.0.113.5"
              className="h-9 font-mono text-xs"
            />
          </Field>

          <Field label="Expires">
            <div className="grid grid-cols-4 gap-2">
              {[
                { id: '30d' as const, label: '30 days' },
                { id: '90d' as const, label: '90 days' },
                { id: '1y' as const, label: '1 year' },
                { id: 'never' as const, label: 'Never' },
              ].map((o) => {
                const active = expiry === o.id
                return (
                  <button
                    key={o.id}
                    onClick={() => setExpiry(o.id)}
                    className={cn(
                      'flex flex-col items-center gap-1 rounded-lg border p-2 text-xs transition-colors',
                      active ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                    )}
                  >
                    <Calendar size={12} className={active ? 'text-primary' : 'text-muted-foreground'} />
                    <span className={cn(active && 'font-semibold text-primary')}>{o.label}</span>
                  </button>
                )
              })}
            </div>
            {expiry === 'never' && (
              <div className="mt-2 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-2.5 text-[11px]">
                <AlertTriangle size={11} className="mt-0.5 shrink-0 text-amber-500" />
                <div>
                  <span className="font-medium text-foreground">Non-expiring keys are riskier.</span>{' '}
                  Set an expiration date and rotate periodically.
                </div>
              </div>
            )}
          </Field>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button size="sm" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button size="sm" onClick={create} disabled={!name.trim()}>
            Generate key
          </Button>
        </footer>
      </div>
    </div>
  )
}

/* ─── Generated dialog ─────────────────────────────────────────────────── */

function GeneratedDialog({
  name,
  token,
  onClose,
}: {
  name: string
  token: string
  onClose: () => void
}) {
  const [copied, setCopied] = useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(token)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {}
  }
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm">
      <div className="flex w-full max-w-xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl">
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-center gap-2">
            <span className="flex h-7 w-7 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-500">
              <Check size={14} strokeWidth={2.5} />
            </span>
            <h2 className="text-lg font-semibold">Key generated</h2>
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            Copy this token now. It won't be shown again.
          </p>
        </header>

        <div className="space-y-3 px-6 py-5">
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{name}</div>
          <div className="flex items-center gap-2">
            <code className="flex-1 break-all rounded-md border border-input bg-background/40 px-3 py-2.5 font-mono text-xs">
              {token}
            </code>
            <Button size="sm" variant="outline" onClick={copy}>
              {copied ? (
                <Check size={14} className="mr-1.5 text-emerald-500" />
              ) : (
                <Copy size={14} className="mr-1.5" />
              )}
              {copied ? 'Copied' : 'Copy'}
            </Button>
          </div>

          <div className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-xs">
            <AlertTriangle size={14} className="mt-0.5 shrink-0 text-amber-500" />
            <div>
              <span className="font-medium text-foreground">Treat this as a credential.</span>{' '}
              Store it in a secret manager. Never commit to source control. If you lose it, generate
              a new one — there's no recovery.
            </div>
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button size="sm" onClick={onClose} disabled={!copied}>
            {copied ? 'Done' : 'Copy first to close'}
          </Button>
        </footer>
      </div>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="rounded-lg border border-border bg-card p-4">
      <h3 className="mb-3 text-sm font-semibold">{title}</h3>
      {children}
    </section>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div>
      <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
      {hint && <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>}
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

function relativeFromOrTo(iso: string) {
  const diff = new Date(iso).getTime() - NOW
  const days = Math.round(diff / 86_400_000)
  if (days < 0) return `${Math.abs(days)}d ago`
  if (days === 0) return 'today'
  return `in ${days}d`
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
