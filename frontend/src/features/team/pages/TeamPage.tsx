import { useMemo, useState } from 'react'
import {
  AlertCircle,
  Copy,
  Crown,
  Eye,
  Globe2,
  History,
  Key,
  KeyRound,
  ListFilter,
  Lock,
  Mail,
  MoreHorizontal,
  Pencil,
  Plus,
  Search,
  Send,
  Shield,
  ShieldCheck,
  Trash2,
  UserCircle,
  UserPlus,
  UserX,
  Users,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Role = 'admin' | 'maintainer' | 'reader' | 'none'
type IdentitySource = 'local' | 'sso-okta' | 'sso-azure' | 'sso-google'
type MemberStatus = 'active' | 'invited' | 'disabled'

interface Workspace {
  id: string
  name: string
  tone: string
}

interface WorkspaceMembership {
  workspaceId: string
  role: Role
  inherited?: boolean // when role inherits from global
}

interface Member {
  id: string
  name: string
  email: string
  initials: string
  tone: string
  status: MemberStatus
  globalRole: Role
  workspaces: WorkspaceMembership[]
  source: IdentitySource
  mfa: boolean
  joinedAt: string
  lastSeen?: string
  invitedBy?: string
  invitedAt?: string
}

/* ─── Workspaces (matches Topbar mock) ─────────────────────────────────── */

const WORKSPACES: Workspace[] = [
  { id: 'ws-acme', name: 'Acme Corp', tone: 'from-fuchsia-500 to-violet-600' },
  { id: 'ws-globex', name: 'Globex Industries', tone: 'from-sky-500 to-cyan-500' },
  { id: 'ws-initech', name: 'Initech', tone: 'from-amber-500 to-orange-500' },
]

const wsName = (id: string) => WORKSPACES.find((w) => w.id === id)?.name ?? id

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const MEMBERS: Member[] = [
  {
    id: 'u-jsmith',
    name: 'Jordan Smith',
    email: 'jsmith@utmstack.com',
    initials: 'JS',
    tone: 'bg-sky-500',
    status: 'active',
    globalRole: 'admin',
    workspaces: WORKSPACES.map((w) => ({ workspaceId: w.id, role: 'admin', inherited: true })),
    source: 'sso-okta',
    mfa: true,
    joinedAt: days(540),
    lastSeen: min(11),
  },
  {
    id: 'u-mthomas',
    name: 'Mara Thomas',
    email: 'mthomas@utmstack.com',
    initials: 'MT',
    tone: 'bg-emerald-500',
    status: 'active',
    globalRole: 'maintainer',
    workspaces: [
      { workspaceId: 'ws-acme', role: 'maintainer', inherited: true },
      { workspaceId: 'ws-globex', role: 'admin' },
      { workspaceId: 'ws-initech', role: 'reader' },
    ],
    source: 'sso-okta',
    mfa: true,
    joinedAt: days(420),
    lastSeen: min(8),
  },
  {
    id: 'u-tferguson',
    name: 'Tomás Ferguson',
    email: 'tferguson@utmstack.com',
    initials: 'TF',
    tone: 'bg-amber-500',
    status: 'active',
    globalRole: 'maintainer',
    workspaces: [
      { workspaceId: 'ws-acme', role: 'maintainer', inherited: true },
      { workspaceId: 'ws-globex', role: 'maintainer', inherited: true },
    ],
    source: 'sso-okta',
    mfa: true,
    joinedAt: days(360),
    lastSeen: hr(2),
  },
  {
    id: 'u-aluna',
    name: 'Ana Luna',
    email: 'aluna@utmstack.com',
    initials: 'AL',
    tone: 'bg-violet-500',
    status: 'active',
    globalRole: 'reader',
    workspaces: [
      { workspaceId: 'ws-acme', role: 'reader', inherited: true },
      { workspaceId: 'ws-globex', role: 'maintainer' },
    ],
    source: 'sso-azure',
    mfa: true,
    joinedAt: days(180),
    lastSeen: min(26),
  },
  {
    id: 'u-kpatel',
    name: 'Kavya Patel',
    email: 'kpatel@utmstack.com',
    initials: 'KP',
    tone: 'bg-rose-500',
    status: 'active',
    globalRole: 'reader',
    workspaces: [{ workspaceId: 'ws-acme', role: 'reader', inherited: true }],
    source: 'sso-okta',
    mfa: false,
    joinedAt: days(60),
    lastSeen: hr(3),
  },
  {
    id: 'u-lpark',
    name: 'Liam Park',
    email: 'lpark@contractor-x.com',
    initials: 'LP',
    tone: 'bg-fuchsia-500',
    status: 'active',
    globalRole: 'none',
    workspaces: [{ workspaceId: 'ws-initech', role: 'reader' }],
    source: 'sso-azure',
    mfa: true,
    joinedAt: days(45),
    lastSeen: hr(8),
  },
  {
    id: 'u-rodriguez',
    name: 'Carlos Rodriguez',
    email: 'crodriguez@utmstack.com',
    initials: 'CR',
    tone: 'bg-cyan-500',
    status: 'active',
    globalRole: 'maintainer',
    workspaces: [
      { workspaceId: 'ws-acme', role: 'maintainer', inherited: true },
      { workspaceId: 'ws-globex', role: 'maintainer', inherited: true },
      { workspaceId: 'ws-initech', role: 'maintainer', inherited: true },
    ],
    source: 'sso-okta',
    mfa: false,
    joinedAt: days(220),
    lastSeen: hr(36),
  },
  {
    id: 'u-okwabi',
    name: 'Ama Okwabi',
    email: 'aokwabi@utmstack.com',
    initials: 'AO',
    tone: 'bg-orange-500',
    status: 'active',
    globalRole: 'admin',
    workspaces: WORKSPACES.map((w) => ({ workspaceId: w.id, role: 'admin', inherited: true })),
    source: 'local',
    mfa: true,
    joinedAt: days(720),
    lastSeen: min(45),
  },
  {
    id: 'u-jdoe',
    name: 'James Doe',
    email: 'jdoe@utmstack.com',
    initials: 'JD',
    tone: 'bg-zinc-500',
    status: 'disabled',
    globalRole: 'reader',
    workspaces: [{ workspaceId: 'ws-acme', role: 'reader', inherited: true }],
    source: 'sso-okta',
    mfa: false,
    joinedAt: days(900),
    lastSeen: days(60),
  },
  {
    id: 'inv-1',
    name: 'Sara Mendel',
    email: 'smendel@utmstack.com',
    initials: 'SM',
    tone: 'bg-pink-500',
    status: 'invited',
    globalRole: 'maintainer',
    workspaces: [
      { workspaceId: 'ws-acme', role: 'maintainer', inherited: true },
    ],
    source: 'sso-okta',
    mfa: false,
    joinedAt: '',
    invitedBy: 'jsmith',
    invitedAt: hr(4),
  },
  {
    id: 'inv-2',
    name: 'Daniel Cho',
    email: 'dcho@utmstack.com',
    initials: 'DC',
    tone: 'bg-teal-500',
    status: 'invited',
    globalRole: 'reader',
    workspaces: [{ workspaceId: 'ws-globex', role: 'maintainer' }],
    source: 'sso-okta',
    mfa: false,
    joinedAt: '',
    invitedBy: 'mthomas',
    invitedAt: days(2),
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const ROLE_META: Record<Role, { label: string; tone: string; pill: string; icon: LucideIcon; description: string }> = {
  admin: {
    label: 'Admin',
    tone: 'text-amber-500',
    pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300',
    icon: Crown,
    description: 'Full control. Manage members, billing, integrations, rules, and workspace settings.',
  },
  maintainer: {
    label: 'Maintainer',
    tone: 'text-sky-500',
    pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300',
    icon: Shield,
    description: 'Configure rules, integrations, dashboards, and respond to alerts. Cannot manage members or billing.',
  },
  reader: {
    label: 'Reader',
    tone: 'text-emerald-500',
    pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300',
    icon: Eye,
    description: 'View-only. See alerts, incidents, dashboards, and reports — no changes.',
  },
  none: {
    label: 'No global role',
    tone: 'text-muted-foreground',
    pill: 'bg-muted text-muted-foreground ring-border',
    icon: UserX,
    description: 'No org-level access. Only sees workspaces where they have an explicit role.',
  },
}

const SOURCE_META: Record<IdentitySource, { label: string; tone: string }> = {
  local: { label: 'Local', tone: 'text-muted-foreground' },
  'sso-okta': { label: 'SSO · Okta', tone: 'text-sky-500' },
  'sso-azure': { label: 'SSO · Entra', tone: 'text-fuchsia-500' },
  'sso-google': { label: 'SSO · Google', tone: 'text-emerald-500' },
}

const STATUS_META: Record<MemberStatus, { dot: string; label: string }> = {
  active: { dot: 'bg-emerald-500', label: 'Active' },
  invited: { dot: 'bg-amber-500', label: 'Invited' },
  disabled: { dot: 'bg-zinc-400', label: 'Disabled' },
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

type Tab = 'all' | 'admins' | 'maintainers' | 'readers' | 'invited' | 'no-mfa'

const TABS: { id: Tab; label: string; predicate: (m: Member) => boolean }[] = [
  { id: 'all', label: 'All members', predicate: (m) => m.status !== 'invited' },
  { id: 'admins', label: 'Admins', predicate: (m) => m.globalRole === 'admin' && m.status !== 'invited' },
  { id: 'maintainers', label: 'Maintainers', predicate: (m) => m.globalRole === 'maintainer' && m.status !== 'invited' },
  { id: 'readers', label: 'Readers', predicate: (m) => m.globalRole === 'reader' && m.status !== 'invited' },
  { id: 'invited', label: 'Pending invites', predicate: (m) => m.status === 'invited' },
  { id: 'no-mfa', label: 'No MFA', predicate: (m) => !m.mfa && m.status === 'active' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function TeamPage() {
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [openMember, setOpenMember] = useState<Member | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)

  const filtered = useMemo(() => {
    const t = TABS.find((x) => x.id === tab) ?? TABS[0]
    return MEMBERS.filter(t.predicate).filter((m) =>
      search ? (m.name + m.email).toLowerCase().includes(search.toLowerCase()) : true
    )
  }, [tab, search])

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header onInvite={() => setInviteOpen(true)} />

      <div className="mt-5 grid grid-cols-12 gap-4">
        <div className="col-span-12 lg:col-span-8">
          <ActivityCard />
        </div>
        <div className="col-span-12 lg:col-span-4">
          <SecurityInsightsCard />
        </div>
      </div>

      <div className="mt-5">
        <TabsRow current={tab} onChange={setTab} />
      </div>

      <Toolbar search={search} onSearch={setSearch} />

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <ListHeader />
        {filtered.map((m) => (
          <MemberRow key={m.id} member={m} onOpen={() => setOpenMember(m)} />
        ))}
        {filtered.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            No members match this view.
          </div>
        )}
      </div>

      {openMember && <MemberDrawer member={openMember} onClose={() => setOpenMember(null)} />}
      {inviteOpen && <InviteDialog onClose={() => setInviteOpen(false)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ onInvite }: { onInvite: () => void }) {
  const active = MEMBERS.filter((m) => m.status === 'active').length
  const pending = MEMBERS.filter((m) => m.status === 'invited').length
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Team</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {active} active{pending > 0 ? ` · ${pending} invited` : ''}
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Members of your organization. Roles can be assigned globally or per workspace.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <KeyRound size={14} className="mr-2" />
          SSO & SCIM
        </Button>
        <Button size="sm" onClick={onInvite}>
          <UserPlus size={14} className="mr-2" />
          Invite member
        </Button>
      </div>
    </header>
  )
}

/* ─── Activity card ────────────────────────────────────────────────────── */

function ActivityCard() {
  const buckets = 60
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.round(6 + Math.sin(i / 5) * 2 + Math.sin(i / 11) * 1.5 + (i > 50 ? 1.5 : 0))
      ),
    []
  )
  const w = 1000
  const h = 100
  const max = Math.max(...data) * 1.2
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

  const activeNow = 4
  const totalSeats = MEMBERS.filter((m) => m.status === 'active').length

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Members online · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{activeNow}</span>
            <span className="text-sm text-muted-foreground">
              of <span className="text-foreground">{totalSeats}</span> active members online now
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="teamActGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(14 165 233)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(14 165 233)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#teamActGrad)" />
        <path
          d={linePath}
          fill="none"
          stroke="rgb(14 165 233)"
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

/* ─── Security insights ────────────────────────────────────────────────── */

function SecurityInsightsCard() {
  const noMfa = MEMBERS.filter((m) => !m.mfa && m.status === 'active').length
  const stale = MEMBERS.filter((m) => {
    if (!m.lastSeen || m.status !== 'active') return false
    return new Date(m.lastSeen).getTime() < NOW - 30 * 86_400_000
  }).length
  const local = MEMBERS.filter((m) => m.source === 'local' && m.status === 'active').length
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <ShieldCheck size={14} className="text-emerald-500" />
        <h3 className="text-sm font-semibold">Security posture</h3>
      </div>

      {noMfa > 0 && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs leading-relaxed">
          <span className="font-medium text-foreground">
            {noMfa} member{noMfa === 1 ? '' : 's'} without MFA
          </span>
          {' '}— enforce MFA across the org to prevent account takeover.{' '}
          <button className="text-primary hover:underline">Enforce</button>
        </div>
      )}

      {local > 0 && (
        <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
          <span className="font-medium text-foreground">
            {local} local account{local === 1 ? '' : 's'}
          </span>
          {' '}— consider migrating to SSO for centralized lifecycle.{' '}
          <button className="text-primary hover:underline">Set up SSO</button>
        </div>
      )}

      <div className="rounded-lg border border-border bg-background/40 p-3 text-xs leading-relaxed">
        {stale === 0 ? (
          <>
            <span className="font-medium text-foreground">No stale members.</span>{' '}
            Everyone has signed in within the last 30 days.
          </>
        ) : (
          <>
            <span className="font-medium text-foreground">
              {stale} member{stale === 1 ? '' : 's'} inactive 30+ days
            </span>
            .{' '}
            <button className="text-primary hover:underline">Review</button>
          </>
        )}
      </div>
    </div>
  )
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

function TabsRow({ current, onChange }: { current: Tab; onChange: (t: Tab) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const active = current === t.id
        const count = MEMBERS.filter(t.predicate).length
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

/* ─── Toolbar ──────────────────────────────────────────────────────────── */

function Toolbar({ search, onSearch }: { search: string; onSearch: (s: string) => void }) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          placeholder="Search by name or email…"
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

/* ─── Member list ──────────────────────────────────────────────────────── */

const COLS = '40px 1fr 130px 1fr 110px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <div />
      <div>Member</div>
      <div>Global role</div>
      <div>Workspaces</div>
      <div>Source</div>
      <div>Last active</div>
      <div />
    </div>
  )
}

function MemberRow({ member: m, onOpen }: { member: Member; onOpen: () => void }) {
  const role = ROLE_META[m.globalRole]
  const status = STATUS_META[m.status]
  return (
    <div
      onClick={onOpen}
      className={cn(
        'group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs hover:bg-muted/40 last:border-b-0',
        m.status === 'disabled' && 'opacity-60'
      )}
      style={{ gridTemplateColumns: COLS }}
    >
      <div className="relative">
        <span
          className={cn(
            'flex h-9 w-9 items-center justify-center rounded-full text-[11px] font-semibold text-white',
            m.tone
          )}
        >
          {m.initials}
        </span>
        {m.status === 'invited' && (
          <span className="absolute -right-0.5 -top-0.5 h-3 w-3 rounded-full bg-amber-500 ring-2 ring-card" />
        )}
        {m.status === 'active' && !m.mfa && (
          <span
            className="absolute -right-0.5 -bottom-0.5 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-red-500 ring-2 ring-card"
            title="No MFA"
          >
            <AlertCircle size={9} className="text-white" />
          </span>
        )}
      </div>
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{m.name}</span>
          {m.status !== 'active' && (
            <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-[9px] uppercase text-muted-foreground">
              <span className={cn('h-1.5 w-1.5 rounded-full', status.dot)} />
              {status.label}
            </span>
          )}
        </div>
        <div className="truncate text-[11px] text-muted-foreground">{m.email}</div>
      </div>
      <div>
        <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', role.pill)}>
          <role.icon size={10} />
          {role.label}
        </span>
      </div>
      <div>
        <WorkspaceBadges memberships={m.workspaces} />
      </div>
      <div className="text-[11px]">
        <span className={SOURCE_META[m.source].tone}>{SOURCE_META[m.source].label}</span>
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {m.lastSeen ? relativeTime(m.lastSeen) : m.status === 'invited' ? `invited ${relativeTime(m.invitedAt!)}` : '—'}
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

function WorkspaceBadges({ memberships }: { memberships: WorkspaceMembership[] }) {
  if (memberships.length === 0) {
    return <span className="text-[11px] italic text-muted-foreground">none</span>
  }
  // Group by role
  const byRole = memberships.reduce<Record<string, WorkspaceMembership[]>>((acc, m) => {
    acc[m.role] = acc[m.role] ?? []
    acc[m.role].push(m)
    return acc
  }, {})
  const order: Role[] = ['admin', 'maintainer', 'reader']
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {order
        .filter((r) => byRole[r])
        .map((r) => {
          const items = byRole[r]
          const role = ROLE_META[r]
          return (
            <span
              key={r}
              className="inline-flex items-center gap-1 text-[11px]"
              title={items.map((i) => wsName(i.workspaceId)).join(', ')}
            >
              <role.icon size={10} className={role.tone} />
              <span className={role.tone}>{role.label}</span>
              <span className="font-mono text-[10px] text-muted-foreground">
                in {items.length}
              </span>
            </span>
          )
        })}
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type DrawerTab = 'overview' | 'permissions' | 'activity'

function MemberDrawer({ member: m, onClose }: { member: Member; onClose: () => void }) {
  const [tab, setTab] = useState<DrawerTab>('permissions')
  const role = ROLE_META[m.globalRole]
  const status = STATUS_META[m.status]
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[920px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-5">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <span
                className={cn(
                  'flex h-12 w-12 shrink-0 items-center justify-center rounded-full text-base font-semibold text-white',
                  m.tone
                )}
              >
                {m.initials}
              </span>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span className="inline-flex items-center gap-1">
                    <span className={cn('h-1.5 w-1.5 rounded-full', status.dot)} />
                    {status.label}
                  </span>
                  <span>·</span>
                  <span>{SOURCE_META[m.source].label}</span>
                  <span>·</span>
                  <span>
                    {m.status === 'invited' && m.invitedAt ? `invited ${relativeTime(m.invitedAt)}` : `joined ${relativeTime(m.joinedAt)}`}
                  </span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{m.name}</h2>
                <div className="truncate font-mono text-xs text-muted-foreground">{m.email}</div>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', role.pill)}>
                    <role.icon size={11} />
                    {role.label} (global)
                  </span>
                  {m.mfa ? (
                    <span className="inline-flex items-center gap-1 rounded-md border border-emerald-500/30 bg-emerald-500/5 px-1.5 py-0.5 text-emerald-600 dark:text-emerald-400">
                      <Lock size={10} />
                      MFA
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 rounded-md border border-red-500/30 bg-red-500/5 px-1.5 py-0.5 text-red-600 dark:text-red-300">
                      <AlertCircle size={10} />
                      No MFA
                    </span>
                  )}
                </div>
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
            {m.status === 'invited' ? (
              <>
                <Button size="sm">
                  <Send size={13} className="mr-1.5" />
                  Resend invite
                </Button>
                <Button size="sm" variant="outline">
                  <Copy size={13} className="mr-1.5" />
                  Copy invite link
                </Button>
                <Button size="sm" variant="outline" className="text-red-500 hover:bg-red-500/10">
                  <Trash2 size={13} className="mr-1.5" />
                  Revoke invite
                </Button>
              </>
            ) : (
              <>
                <Button size="sm">
                  <Pencil size={13} className="mr-1.5" />
                  Edit roles
                </Button>
                <Button size="sm" variant="outline">
                  <Mail size={13} className="mr-1.5" />
                  Email
                </Button>
                <Button size="sm" variant="outline">
                  <Key size={13} className="mr-1.5" />
                  Reset password
                </Button>
                <Button size="sm" variant="outline" className="ml-auto text-red-500 hover:bg-red-500/10">
                  <UserX size={13} className="mr-1.5" />
                  Disable
                </Button>
              </>
            )}
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTabBtn id="permissions" current={tab} onChange={setTab} icon={Shield}>
            Permissions
          </DrawerTabBtn>
          <DrawerTabBtn id="overview" current={tab} onChange={setTab} icon={UserCircle}>
            Profile
          </DrawerTabBtn>
          <DrawerTabBtn id="activity" current={tab} onChange={setTab} icon={History}>
            Activity
          </DrawerTabBtn>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'permissions' && <PermissionsTab member={m} />}
          {tab === 'overview' && <ProfileTab member={m} />}
          {tab === 'activity' && <ActivityTab member={m} />}
        </div>
      </div>
    </div>
  )
}

function DrawerTabBtn({
  id,
  current,
  onChange,
  icon: Icon,
  children,
}: {
  id: DrawerTab
  current: DrawerTab
  onChange: (t: DrawerTab) => void
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

/* ─── Permissions tab ──────────────────────────────────────────────────── */

function PermissionsTab({ member: m }: { member: Member }) {
  return (
    <div className="space-y-5 p-6">
      <Section title="Global role" subtitle="Applied to the entire organization. Workspaces inherit this role unless overridden.">
        <RolePicker current={m.globalRole} />
      </Section>

      <Section title="Workspace-level roles" subtitle="Override the global role per workspace. Adding an explicit role replaces the inherited one.">
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <div
            className="grid grid-cols-[40px_1fr_140px_100px_36px] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          >
            <div />
            <div>Workspace</div>
            <div>Role</div>
            <div>Source</div>
            <div />
          </div>
          {WORKSPACES.map((w) => {
            const m_w = m.workspaces.find((x) => x.workspaceId === w.id)
            const effective = m_w ? m_w.role : m.globalRole
            const inherited = m_w ? m_w.inherited === true : true
            const role = ROLE_META[effective]
            return (
              <div
                key={w.id}
                className="grid items-center gap-3 border-b border-border/60 px-3 py-2.5 text-xs last:border-b-0 hover:bg-muted/30"
                style={{ gridTemplateColumns: '40px 1fr 140px 100px 36px' }}
              >
                <span
                  className={cn(
                    'flex h-7 w-7 items-center justify-center rounded-md bg-gradient-to-br text-[10px] font-semibold text-white',
                    w.tone
                  )}
                >
                  {w.name
                    .split(' ')
                    .map((p) => p[0])
                    .slice(0, 2)
                    .join('')
                    .toUpperCase()}
                </span>
                <div className="font-medium">{w.name}</div>
                <div>
                  <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', role.pill)}>
                    <role.icon size={10} />
                    {role.label}
                  </span>
                </div>
                <div className="text-[11px] text-muted-foreground">
                  {inherited ? 'Inherited' : 'Explicit'}
                </div>
                <div className="flex justify-end">
                  <button className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground">
                    <Pencil size={11} />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
        <div className="mt-2 flex items-center gap-3 text-[11px]">
          <button className="inline-flex items-center gap-1 text-primary hover:underline">
            <Plus size={11} /> Override role in another workspace
          </button>
        </div>
      </Section>

      <Section title="What this role can do">
        <RoleCheatsheet />
      </Section>
    </div>
  )
}

function RolePicker({ current }: { current: Role }) {
  const roles: Role[] = ['admin', 'maintainer', 'reader', 'none']
  return (
    <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
      {roles.map((r) => {
        const meta = ROLE_META[r]
        const Icon = meta.icon
        const active = current === r
        return (
          <button
            key={r}
            className={cn(
              'flex items-start gap-3 rounded-lg border p-3 text-left transition-colors',
              active ? 'border-primary/40 bg-primary/5' : 'border-border bg-card hover:bg-muted/40'
            )}
          >
            <span
              className={cn(
                'flex h-8 w-8 shrink-0 items-center justify-center rounded-md ring-1 ring-inset',
                meta.pill
              )}
            >
              <Icon size={14} />
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <span className={cn('text-sm font-semibold', meta.tone)}>{meta.label}</span>
                {active && (
                  <span className="rounded bg-primary/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-primary">
                    current
                  </span>
                )}
              </div>
              <p className="mt-0.5 text-[11px] leading-relaxed text-muted-foreground">
                {meta.description}
              </p>
            </div>
          </button>
        )
      })}
    </div>
  )
}

function RoleCheatsheet() {
  const matrix: { capability: string; admin: boolean; maintainer: boolean; reader: boolean }[] = [
    { capability: 'View alerts, incidents, dashboards', admin: true, maintainer: true, reader: true },
    { capability: 'Acknowledge and triage alerts', admin: true, maintainer: true, reader: false },
    { capability: 'Configure alerting & tagging rules', admin: true, maintainer: true, reader: false },
    { capability: 'Connect / configure integrations', admin: true, maintainer: true, reader: false },
    { capability: 'Run SOAR playbooks', admin: true, maintainer: true, reader: false },
    { capability: 'Manage members, billing, SSO', admin: true, maintainer: false, reader: false },
    { capability: 'Create / delete workspaces', admin: true, maintainer: false, reader: false },
  ]
  return (
    <div className="overflow-hidden rounded-lg border border-border bg-card">
      <div className="grid grid-cols-[1fr_80px_100px_80px] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
        <div>Capability</div>
        <div className="text-center">Admin</div>
        <div className="text-center">Maintainer</div>
        <div className="text-center">Reader</div>
      </div>
      {matrix.map((row, i) => (
        <div
          key={i}
          className="grid grid-cols-[1fr_80px_100px_80px] items-center gap-3 border-b border-border/60 px-3 py-2 text-xs last:border-b-0"
        >
          <div>{row.capability}</div>
          <Tick value={row.admin} />
          <Tick value={row.maintainer} />
          <Tick value={row.reader} />
        </div>
      ))}
    </div>
  )
}

function Tick({ value }: { value: boolean }) {
  return (
    <div className="flex justify-center">
      {value ? (
        <span className="text-emerald-500">●</span>
      ) : (
        <span className="text-muted-foreground/40">—</span>
      )}
    </div>
  )
}

/* ─── Profile tab ──────────────────────────────────────────────────────── */

function ProfileTab({ member: m }: { member: Member }) {
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        <Section title="Profile">
          <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
            <Row k="Display name">{m.name}</Row>
            <Row k="Email"><span className="font-mono">{m.email}</span></Row>
            <Row k="User ID"><span className="font-mono text-muted-foreground">{m.id}</span></Row>
            <Row k="Identity source">
              <span className={SOURCE_META[m.source].tone}>{SOURCE_META[m.source].label}</span>
            </Row>
            <Row k="Joined">{m.joinedAt ? relativeTime(m.joinedAt) : '—'}</Row>
            <Row k="Last seen">{m.lastSeen ? relativeTime(m.lastSeen) : '—'}</Row>
          </dl>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">Security</div>
          <ul className="space-y-2 text-xs">
            <li className="flex items-center justify-between">
              <span className="inline-flex items-center gap-2">
                <Lock size={11} />
                Multi-factor auth
              </span>
              {m.mfa ? (
                <span className="text-emerald-500">Enabled</span>
              ) : (
                <span className="text-red-500">Disabled</span>
              )}
            </li>
            <li className="flex items-center justify-between">
              <span className="inline-flex items-center gap-2">
                <Globe2 size={11} />
                Active sessions
              </span>
              <span className="font-mono">{m.status === 'active' ? '2' : '0'}</span>
            </li>
            <li className="flex items-center justify-between">
              <span className="inline-flex items-center gap-2">
                <Key size={11} />
                API keys
              </span>
              <span className="font-mono">{m.globalRole === 'admin' ? '1' : '0'}</span>
            </li>
          </ul>
        </div>
      </aside>
    </div>
  )
}

/* ─── Activity tab ─────────────────────────────────────────────────────── */

function ActivityTab({ member: m }: { member: Member }) {
  const events = [
    { ts: m.lastSeen ?? min(60), what: 'Signed in', detail: 'from 10.0.4.91 · Chrome on macOS' },
    { ts: hr(2), what: 'Acknowledged alert', detail: 'a-9f04c2 — Suspected Kerberoasting' },
    { ts: hr(8), what: 'Modified rule', detail: 'r-encoded-powershell — adjusted threshold to 50' },
    { ts: days(1), what: 'Invited member', detail: 'smendel@utmstack.com as Maintainer' },
    { ts: days(3), what: 'Connected integration', detail: 'AWS · workspace Acme Corp' },
  ]
  if (m.status === 'invited') {
    return (
      <div className="p-6">
        <div className="rounded-lg border border-dashed border-border bg-card px-6 py-12 text-center text-sm text-muted-foreground">
          This member hasn't accepted the invitation yet.
        </div>
      </div>
    )
  }
  return (
    <div className="p-6">
      <ol className="relative space-y-4 border-l border-border pl-6">
        {events.map((e, i) => (
          <li key={i} className="relative">
            <span className="absolute -left-[27px] top-1 h-3 w-3 rounded-full border-2 border-card bg-primary" />
            <div className="text-xs">
              <span className="font-medium">{e.what}</span> · {e.detail}
            </div>
            <div className="text-[10px] text-muted-foreground">{relativeTime(e.ts)}</div>
          </li>
        ))}
      </ol>
    </div>
  )
}

/* ─── Invite dialog ────────────────────────────────────────────────────── */

function InviteDialog({ onClose }: { onClose: () => void }) {
  const [emails, setEmails] = useState('')
  const [globalRole, setGlobalRole] = useState<Role>('reader')
  const [perWorkspace, setPerWorkspace] = useState<Record<string, Role>>({})

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4" onClick={onClose}>
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <UserPlus size={18} />
              Invite member
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">
              We'll email an invitation. They'll join your organization once they accept.
            </p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="space-y-4 px-6 py-5">
          <div>
            <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
              Email addresses
            </label>
            <Input
              value={emails}
              onChange={(e) => setEmails(e.target.value)}
              placeholder="alex@company.com, sam@company.com"
              className="h-9 text-sm"
            />
            <p className="mt-1 text-[10px] text-muted-foreground">Separate multiple emails with commas.</p>
          </div>

          <div>
            <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
              Global role
            </label>
            <div className="grid grid-cols-3 gap-2">
              {(['admin', 'maintainer', 'reader'] as Role[]).map((r) => {
                const meta = ROLE_META[r]
                const Icon = meta.icon
                const active = globalRole === r
                return (
                  <button
                    key={r}
                    onClick={() => setGlobalRole(r)}
                    className={cn(
                      'flex flex-col items-start gap-1 rounded-lg border p-2.5 text-left text-xs transition-colors',
                      active ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                    )}
                  >
                    <Icon size={13} className={meta.tone} />
                    <span className={cn('font-medium', meta.tone)}>{meta.label}</span>
                  </button>
                )
              })}
            </div>
          </div>

          <div>
            <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
              Workspace overrides (optional)
            </label>
            <div className="space-y-1.5">
              {WORKSPACES.map((w) => {
                const role = perWorkspace[w.id] ?? globalRole
                return (
                  <div
                    key={w.id}
                    className="flex items-center justify-between gap-3 rounded-md border border-border bg-card px-3 py-2 text-xs"
                  >
                    <div className="flex items-center gap-2">
                      <span
                        className={cn(
                          'flex h-6 w-6 items-center justify-center rounded-md bg-gradient-to-br text-[10px] font-semibold text-white',
                          w.tone
                        )}
                      >
                        {w.name
                          .split(' ')
                          .map((p) => p[0])
                          .slice(0, 2)
                          .join('')
                          .toUpperCase()}
                      </span>
                      <span>{w.name}</span>
                    </div>
                    <select
                      value={role}
                      onChange={(e) =>
                        setPerWorkspace((prev) => ({ ...prev, [w.id]: e.target.value as Role }))
                      }
                      className="h-7 rounded-md border border-border bg-background px-2 text-[11px]"
                    >
                      <option value="admin">Admin</option>
                      <option value="maintainer">Maintainer</option>
                      <option value="reader">Reader</option>
                      <option value="none">No access</option>
                    </select>
                  </div>
                )
              })}
            </div>
            <p className="mt-1 text-[10px] text-muted-foreground">
              Leave at the global role to inherit. Pick a different value to override per workspace.
            </p>
          </div>
        </div>

        <footer className="flex items-center justify-between gap-3 border-t border-border px-6 py-3">
          <span className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
            <Users size={11} />
            Will use 1 seat per accepted invite
          </span>
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button size="sm">
              <Send size={13} className="mr-1.5" />
              Send invite
            </Button>
          </div>
        </footer>
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
    <section className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3">
        <h4 className="text-sm font-semibold">{title}</h4>
        {subtitle && <p className="mt-0.5 text-[11px] text-muted-foreground">{subtitle}</p>}
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
