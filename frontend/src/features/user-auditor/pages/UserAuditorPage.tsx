import { useMemo, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  ChevronDown,
  Clock,
  Cog,
  Crown,
  Download,
  ExternalLink,
  Eye,
  EyeOff,
  Globe2,
  ListFilter,
  ListIcon,
  LayoutGrid,
  Lock,
  MapPin,
  MoreHorizontal,
  RefreshCw,
  Search,
  Shield,
  Sparkles,
  User,
  UserX,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type UserKind = 'human' | 'service' | 'admin' | 'contractor'
type UserStatus = 'active' | 'inactive' | 'locked' | 'flagged'
type RiskLevel = 'critical' | 'high' | 'medium' | 'low'
type SourceKind = 'ad' | 'azure' | 'okta' | 'local'

interface AuditedUser {
  id: string
  name: string
  username: string
  email?: string
  kind: UserKind
  status: UserStatus
  source: { name: string; kind: SourceKind }
  department?: string
  title?: string
  groups: string[]
  riskScore: number
  riskLevel: RiskLevel
  firstSeen: string
  lastSeen: string
  lastLogin: { ts: string; ip: string; country?: string; countryCode?: string }
  anomalies: { ts: string; kind: string; detail: string }[]
  recentEvents: number
  hourlyActivity: number[]
  attributes: { key: string; value: string }[]
  recentAlerts: { id: string; name: string; severity: 'critical' | 'high' | 'medium' | 'low'; ts: string }[]
  flags: string[]
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const sineActivity = (seed: number) =>
  Array.from({ length: 24 }, (_, i) =>
    Math.max(0, Math.round(Math.abs(Math.sin((i + seed) / 2.4)) * 8 + (i > 8 && i < 18 ? 6 : 0)))
  )

const USERS: AuditedUser[] = [
  {
    id: 'u-svc_backup',
    name: 'svc_backup',
    username: 'svc_backup',
    email: undefined,
    kind: 'service',
    status: 'flagged',
    source: { name: 'CORP.LOCAL', kind: 'ad' },
    department: 'IT / Infrastructure',
    title: 'Backup service account',
    groups: ['Domain Users', 'Backup Operators', 'svc-accounts'],
    riskScore: 96,
    riskLevel: 'critical',
    firstSeen: days(420),
    lastSeen: min(2),
    lastLogin: { ts: min(2), ip: '10.4.18.22', country: 'United States', countryCode: 'US' },
    anomalies: [
      { ts: min(4), kind: 'kerberoasting', detail: 'Unusual TGS-REQ burst (12 SPNs in 4 minutes)' },
      { ts: min(38), kind: 'lateral-movement', detail: 'SMB share enumeration from non-standard host' },
      { ts: hr(2), kind: 'off-hours', detail: 'First login outside maintenance window in 90 days' },
    ],
    recentEvents: 1284,
    hourlyActivity: sineActivity(2).map((v, i) => (i === 15 || i === 16 ? v + 18 : v)),
    attributes: [
      { key: 'sAMAccountName', value: 'svc_backup' },
      { key: 'objectSid', value: 'S-1-5-21-3623811015-3361044348-30300820-1147' },
      { key: 'distinguishedName', value: 'CN=svc_backup,OU=Service Accounts,DC=corp,DC=local' },
      { key: 'pwdLastSet', value: '2025-11-14 03:00:00' },
      { key: 'lastLogon', value: min(2) },
      { key: 'memberOf', value: 'Backup Operators' },
      { key: 'servicePrincipalName', value: 'MSSQLSvc/sql-prod-01, HOST/dc-01' },
      { key: 'userAccountControl', value: '512 (NORMAL_ACCOUNT)' },
    ],
    recentAlerts: [
      { id: 'a-9f04c2', name: 'Suspected Kerberoasting against domain controller', severity: 'critical', ts: min(4) },
      { id: 'a-3e88a1', name: 'Suspicious SMB share enumeration', severity: 'medium', ts: min(38) },
      { id: 'a-44a811', name: 'AV signature match — Mimikatz variant', severity: 'high', ts: min(62) },
    ],
    flags: ['compromised-suspected', 'campaign-44'],
  },
  {
    id: 'u-jsmith',
    name: 'Jordan Smith',
    username: 'jsmith',
    email: 'jsmith@utmstack.com',
    kind: 'admin',
    status: 'active',
    source: { name: 'Azure AD', kind: 'azure' },
    department: 'Security',
    title: 'SOC Analyst Lead',
    groups: ['SOC-Admins', 'IR-Team', 'KeyVault-Reader'],
    riskScore: 22,
    riskLevel: 'low',
    firstSeen: days(380),
    lastSeen: min(11),
    lastLogin: { ts: min(11), ip: '10.0.4.91', country: 'United States', countryCode: 'US' },
    anomalies: [],
    recentEvents: 312,
    hourlyActivity: sineActivity(7),
    attributes: [
      { key: 'userPrincipalName', value: 'jsmith@utmstack.com' },
      { key: 'mfaEnabled', value: 'true (FIDO2)' },
      { key: 'lastSignIn', value: min(11) },
      { key: 'department', value: 'Security' },
      { key: 'title', value: 'SOC Analyst Lead' },
      { key: 'manager', value: 'Mara Thomas' },
    ],
    recentAlerts: [],
    flags: ['mfa-enrolled'],
  },
  {
    id: 'u-mthomas',
    name: 'Mara Thomas',
    username: 'mthomas',
    email: 'mthomas@utmstack.com',
    kind: 'human',
    status: 'flagged',
    source: { name: 'CORP.LOCAL', kind: 'ad' },
    department: 'IT',
    title: 'Sysadmin',
    groups: ['Domain Users', 'Server Operators'],
    riskScore: 71,
    riskLevel: 'high',
    firstSeen: days(580),
    lastSeen: min(8),
    lastLogin: { ts: min(8), ip: '203.0.113.7', country: 'Brazil', countryCode: 'BR' },
    anomalies: [
      { ts: min(8), kind: 'impossible-travel', detail: 'Login from Brazil 14 minutes after login from US' },
      { ts: min(23), kind: 'brute-force-target', detail: '312 failed logins preceded successful auth' },
    ],
    recentEvents: 88,
    hourlyActivity: sineActivity(11),
    attributes: [
      { key: 'sAMAccountName', value: 'mthomas' },
      { key: 'mail', value: 'mthomas@utmstack.com' },
      { key: 'department', value: 'IT' },
      { key: 'lastLogon', value: min(8) },
      { key: 'lastLogonIp', value: '203.0.113.7 (BR)' },
    ],
    recentAlerts: [
      { id: 'a-7c2d11', name: 'Multiple failed logins followed by success', severity: 'high', ts: min(23) },
    ],
    flags: ['credential-attack-target'],
  },
  {
    id: 'u-tferguson',
    name: 'Tomás Ferguson',
    username: 'tferguson',
    email: 'tferguson@utmstack.com',
    kind: 'admin',
    status: 'active',
    source: { name: 'CORP.LOCAL', kind: 'ad' },
    department: 'IT / Operations',
    title: 'Senior Sysadmin',
    groups: ['Domain Admins', 'Schema Admins', 'GPO-Editors'],
    riskScore: 58,
    riskLevel: 'medium',
    firstSeen: days(720),
    lastSeen: hr(2),
    lastLogin: { ts: hr(2), ip: '10.4.18.91', country: 'United States', countryCode: 'US' },
    anomalies: [
      { ts: hr(2), kind: 'policy-change', detail: 'GPO modification outside change window' },
    ],
    recentEvents: 41,
    hourlyActivity: sineActivity(13),
    attributes: [
      { key: 'sAMAccountName', value: 'tferguson' },
      { key: 'mail', value: 'tferguson@utmstack.com' },
      { key: 'memberOf', value: 'Domain Admins, Schema Admins' },
      { key: 'lastLogon', value: hr(2) },
    ],
    recentAlerts: [
      { id: 'a-08f3a4', name: 'GPO modification by non-administrator', severity: 'high', ts: hr(2) },
    ],
    flags: ['privileged'],
  },
  {
    id: 'u-aluna',
    name: 'Ana Luna',
    username: 'aluna',
    email: 'aluna@utmstack.com',
    kind: 'human',
    status: 'active',
    source: { name: 'Okta', kind: 'okta' },
    department: 'Engineering',
    title: 'Backend Engineer',
    groups: ['eng-backend', 'aws-developers'],
    riskScore: 14,
    riskLevel: 'low',
    firstSeen: days(180),
    lastSeen: min(26),
    lastLogin: { ts: min(26), ip: '203.0.113.91', country: 'United States', countryCode: 'US' },
    anomalies: [],
    recentEvents: 142,
    hourlyActivity: sineActivity(3),
    attributes: [
      { key: 'login', value: 'aluna@utmstack.com' },
      { key: 'mfa', value: 'TOTP, WebAuthn' },
      { key: 'lastSignIn', value: min(26) },
    ],
    recentAlerts: [],
    flags: ['mfa-enrolled'],
  },
  {
    id: 'u-svc_app',
    name: 'svc_app',
    username: 'svc_app',
    kind: 'service',
    status: 'active',
    source: { name: 'CORP.LOCAL', kind: 'ad' },
    department: 'Engineering',
    title: 'Application service account',
    groups: ['svc-accounts', 'fin-app-readers'],
    riskScore: 32,
    riskLevel: 'low',
    firstSeen: days(540),
    lastSeen: min(58),
    lastLogin: { ts: min(58), ip: '10.0.4.91' },
    anomalies: [],
    recentEvents: 5871,
    hourlyActivity: sineActivity(8),
    attributes: [
      { key: 'sAMAccountName', value: 'svc_app' },
      { key: 'pwdLastSet', value: '2024-08-22 02:00:00' },
      { key: 'pwdAge', value: '252 days' },
      { key: 'lastLogon', value: min(58) },
    ],
    recentAlerts: [],
    flags: ['stale-password'],
  },
  {
    id: 'u-contractor-dev',
    name: 'Liam Park',
    username: 'lpark',
    email: 'lpark@contractor-x.com',
    kind: 'contractor',
    status: 'active',
    source: { name: 'Azure AD', kind: 'azure' },
    department: 'Engineering (Contractor)',
    title: 'Frontend Contractor',
    groups: ['contractors', 'eng-frontend'],
    riskScore: 41,
    riskLevel: 'medium',
    firstSeen: days(45),
    lastSeen: hr(8),
    lastLogin: { ts: hr(8), ip: '203.0.113.220', country: 'Vietnam', countryCode: 'VN' },
    anomalies: [
      { ts: hr(8), kind: 'foreign-login', detail: 'Login from Vietnam — first time from this country' },
    ],
    recentEvents: 22,
    hourlyActivity: sineActivity(17),
    attributes: [
      { key: 'userPrincipalName', value: 'lpark@contractor-x.com' },
      { key: 'guestType', value: 'B2B Guest' },
      { key: 'lastSignIn', value: hr(8) },
    ],
    recentAlerts: [],
    flags: ['external-identity', 'new-country'],
  },
  {
    id: 'u-disabled',
    name: 'Robert Carter',
    username: 'rcarter',
    email: 'rcarter@utmstack.com',
    kind: 'human',
    status: 'inactive',
    source: { name: 'CORP.LOCAL', kind: 'ad' },
    department: 'Sales (former)',
    title: 'Account Executive (Former)',
    groups: ['Disabled Users'],
    riskScore: 8,
    riskLevel: 'low',
    firstSeen: days(900),
    lastSeen: days(45),
    lastLogin: { ts: days(45), ip: '10.4.18.61' },
    anomalies: [],
    recentEvents: 0,
    hourlyActivity: Array(24).fill(0),
    attributes: [
      { key: 'sAMAccountName', value: 'rcarter' },
      { key: 'userAccountControl', value: '514 (DISABLED)' },
      { key: 'disabledDate', value: '2026-03-17' },
    ],
    recentAlerts: [],
    flags: ['disabled'],
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const RISK_STYLE: Record<RiskLevel, { tone: string; ring: string; pill: string; label: string }> = {
  critical: { tone: 'text-red-500', ring: 'ring-red-500/40', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', label: 'Critical' },
  high: { tone: 'text-orange-500', ring: 'ring-orange-500/40', pill: 'bg-orange-500/15 text-orange-600 ring-orange-500/30 dark:text-orange-300', label: 'High' },
  medium: { tone: 'text-amber-500', ring: 'ring-amber-500/40', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', label: 'Medium' },
  low: { tone: 'text-emerald-500', ring: 'ring-emerald-500/40', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', label: 'Low' },
}

const STATUS_META: Record<UserStatus, { label: string; dot: string }> = {
  active: { label: 'Active', dot: 'bg-emerald-500' },
  inactive: { label: 'Disabled', dot: 'bg-muted-foreground' },
  locked: { label: 'Locked', dot: 'bg-amber-500' },
  flagged: { label: 'Flagged', dot: 'bg-red-500 animate-pulse' },
}

const KIND_META: Record<UserKind, { label: string; icon: LucideIcon; tone: string }> = {
  human: { label: 'User', icon: User, tone: 'text-sky-500' },
  admin: { label: 'Admin', icon: Crown, tone: 'text-amber-500' },
  service: { label: 'Service', icon: Cog, tone: 'text-violet-500' },
  contractor: { label: 'Contractor', icon: User, tone: 'text-fuchsia-500' },
}

const SOURCE_META: Record<SourceKind, { label: string }> = {
  ad: { label: 'Active Directory' },
  azure: { label: 'Azure AD' },
  okta: { label: 'Okta' },
  local: { label: 'Local' },
}

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (u: AuditedUser) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All users', count: USERS.length, predicate: () => true },
  { id: 'flagged', label: 'Flagged', count: USERS.filter((u) => u.status === 'flagged').length, predicate: (u) => u.status === 'flagged' },
  { id: 'admins', label: 'Admins', count: USERS.filter((u) => u.kind === 'admin').length, predicate: (u) => u.kind === 'admin' },
  { id: 'service', label: 'Service accounts', count: USERS.filter((u) => u.kind === 'service').length, predicate: (u) => u.kind === 'service' },
  { id: 'contractors', label: 'Contractors', count: USERS.filter((u) => u.kind === 'contractor').length, predicate: (u) => u.kind === 'contractor' },
  { id: 'high-risk', label: 'High risk', count: USERS.filter((u) => u.riskScore >= 60).length, predicate: (u) => u.riskScore >= 60 },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function UserAuditorPage() {
  const [view, setView] = useState<string>('all')
  const [search, setSearch] = useState('')
  const [layout, setLayout] = useState<'list' | 'cards'>('list')
  const [openUser, setOpenUser] = useState<AuditedUser | null>(null)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return USERS.filter(v.predicate).filter((u) =>
      search
        ? (u.name + u.username + (u.email ?? '') + (u.department ?? '') + u.groups.join(' '))
            .toLowerCase()
            .includes(search.toLowerCase())
        : true
    )
  }, [view, search])

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header total={filtered.length} />

      <div className="mt-5">
        <ActivityOverviewCard users={USERS} />
      </div>

      <div className="mt-5">
        <SavedViewTabs current={view} onChange={setView} />
      </div>

      <Toolbar search={search} onSearch={setSearch} layout={layout} onLayoutChange={setLayout} />

      {layout === 'list' ? (
        <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
          <ListHeader />
          {filtered.map((u) => (
            <UserListRow key={u.id} user={u} onOpen={() => setOpenUser(u)} />
          ))}
          {filtered.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">
              No users match this view.
            </div>
          )}
        </div>
      ) : (
        <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((u) => (
            <UserCard key={u.id} user={u} onOpen={() => setOpenUser(u)} />
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
              No users match this view.
            </div>
          )}
        </div>
      )}

      {openUser && <UserDrawer user={openUser} onClose={() => setOpenUser(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total }: { total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">User auditor</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {total.toLocaleString()} tracked
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Identities pulled from Active Directory, Azure AD, and Okta. Behavior anomalies
          continuously scored.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Shield size={14} className="mr-2" />
          Identity sources
        </Button>
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          Export
        </Button>
      </div>
    </header>
  )
}

/* ─── Activity overview ────────────────────────────────────────────────── */

function ActivityOverviewCard({ users }: { users: AuditedUser[] }) {
  const flagged = users.filter((u) => u.status === 'flagged').length
  const admins = users.filter((u) => u.kind === 'admin').length
  const highRisk = users.filter((u) => u.riskScore >= 60).length

  const buckets = Math.max(...users.map((u) => u.hourlyActivity.length), 24)
  const summed = Array.from({ length: buckets }, (_, i) =>
    users.reduce((s, u) => s + (u.hourlyActivity[i] ?? 0), 0)
  )
  const w = 1000
  const h = 100
  const max = Math.max(...summed) * 1.15
  const xs = summed.map((_, i) => (i * w) / (summed.length - 1))
  const ys = summed.map((v) => h - (v / max) * h)
  const linePath = summed.reduce((acc, _, i) => {
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
            User activity · last 24 hours
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">{flagged}</span>
            <span className="text-sm text-muted-foreground">
              <span className="text-foreground">flagged</span>
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{highRisk}</span> high risk
              <span className="mx-2 text-border">·</span>
              <span className="text-foreground">{admins}</span> admins
            </span>
          </div>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="userActGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(14 165 233)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(14 165 233)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#userActGrad)" />
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
  layout: 'list' | 'cards'
  onLayoutChange: (l: 'list' | 'cards') => void
}) {
  return (
    <div className="mt-3 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[280px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search by name, username, email, department, group…"
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      <Button variant="outline" size="sm">
        <Clock size={14} className="mr-2" />
        Last 24 hours
        <ChevronDown size={12} className="ml-1.5 opacity-60" />
      </Button>
      <Button variant="outline" size="sm">
        <ListFilter size={14} className="mr-2" />
        Filters
      </Button>
      <div className="ml-auto flex items-center gap-1 rounded-md border border-border p-0.5">
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
        <button
          onClick={() => onLayoutChange('cards')}
          title="Cards"
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded transition-colors',
            layout === 'cards' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <LayoutGrid size={13} />
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

/* ─── List row ─────────────────────────────────────────────────────────── */

const LIST_COLS = '40px 1fr 110px 120px 60px 80px 110px 36px'

function ListHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div />
      <div>User</div>
      <div>Source</div>
      <div>Status</div>
      <div className="text-right">Risk</div>
      <div className="text-right">Anomalies</div>
      <div>Last seen</div>
      <div />
    </div>
  )
}

function UserListRow({ user, onOpen }: { user: AuditedUser; onOpen: () => void }) {
  const risk = RISK_STYLE[user.riskLevel]
  const st = STATUS_META[user.status]
  return (
    <div
      onClick={onOpen}
      className="group grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <UserAvatar user={user} />
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{user.name}</span>
          {user.kind === 'admin' && <Crown size={11} className="text-amber-500" />}
          {user.flags.includes('compromised-suspected') && (
            <span className="rounded bg-red-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-red-600 ring-1 ring-red-500/30 dark:text-red-300">
              compromised?
            </span>
          )}
          {user.flags.includes('privileged') && (
            <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
              privileged
            </span>
          )}
        </div>
        <div className="truncate text-[11px] text-muted-foreground">
          <span className="font-mono">{user.username}</span>
          {user.email && <span> · {user.email}</span>}
          {user.department && <span> · {user.department}</span>}
        </div>
      </div>
      <div className="text-[11px] text-muted-foreground">
        {SOURCE_META[user.source.kind].label}
      </div>
      <div>
        <span className="inline-flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
      </div>
      <div className="flex justify-end">
        <RiskRing value={user.riskScore} />
      </div>
      <div className={cn('text-right tabular-nums', user.anomalies.length > 0 ? risk.tone : 'text-muted-foreground')}>
        {user.anomalies.length}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {relativeTime(user.lastSeen)}
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

/* ─── Card view ────────────────────────────────────────────────────────── */

function UserCard({ user, onOpen }: { user: AuditedUser; onOpen: () => void }) {
  const st = STATUS_META[user.status]
  return (
    <button
      onClick={onOpen}
      className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <UserAvatar user={user} large />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-semibold">{user.name}</span>
            {user.kind === 'admin' && <Crown size={11} className="text-amber-500" />}
          </div>
          <div className="truncate font-mono text-[11px] text-muted-foreground">
            {user.username}
          </div>
          <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
            {user.title ?? user.department ?? '—'}
          </div>
        </div>
        <RiskRing value={user.riskScore} />
      </div>

      <div className="flex items-center gap-2 text-[11px]">
        <span className="inline-flex items-center gap-1.5 text-muted-foreground">
          <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
          {st.label}
        </span>
        <span className="text-border">·</span>
        <span className="text-muted-foreground">{SOURCE_META[user.source.kind].label}</span>
      </div>

      <div className="grid grid-cols-3 gap-2 border-t border-border pt-3 text-[11px]">
        <CardStat label="Anomalies" value={user.anomalies.length.toString()} tone={user.anomalies.length > 0 ? 'text-red-500' : ''} />
        <CardStat label="Events" value={user.recentEvents.toLocaleString()} />
        <CardStat label="Groups" value={user.groups.length.toString()} />
      </div>

      <div className="text-[11px] text-muted-foreground">
        Last seen <span className="font-mono">{relativeTime(user.lastSeen)}</span>
      </div>
    </button>
  )
}

function CardStat({ label, value, tone }: { label: string; value: string; tone?: string }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-2 py-1.5">
      <div className="text-[9px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-0.5 font-semibold tabular-nums', tone)}>{value}</div>
    </div>
  )
}

/* ─── Avatar ───────────────────────────────────────────────────────────── */

function UserAvatar({ user, large }: { user: AuditedUser; large?: boolean }) {
  const risk = RISK_STYLE[user.riskLevel]
  const kind = KIND_META[user.kind]
  const Icon = kind.icon
  return (
    <div className="relative">
      <div
        className={cn(
          'flex shrink-0 items-center justify-center rounded-lg ring-2',
          large ? 'h-11 w-11' : 'h-7 w-7',
          'bg-muted text-foreground/80',
          risk.ring
        )}
      >
        <Icon size={large ? 18 : 13} className={kind.tone} />
      </div>
      {user.status === 'flagged' && (
        <span className="absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-card animate-pulse" />
      )}
    </div>
  )
}

function RiskRing({ value }: { value: number }) {
  const tone =
    value >= 90
      ? 'stroke-red-500'
      : value >= 70
        ? 'stroke-orange-500'
        : value >= 50
          ? 'stroke-amber-500'
          : 'stroke-emerald-500'
  const r = 9
  const c = 2 * Math.PI * r
  const off = c - (value / 100) * c
  return (
    <div className="relative inline-flex h-6 w-6 items-center justify-center">
      <svg viewBox="0 0 24 24" className="h-6 w-6 -rotate-90">
        <circle cx="12" cy="12" r={r} className="fill-none stroke-muted" strokeWidth="2.5" />
        <circle
          cx="12"
          cy="12"
          r={r}
          className={cn('fill-none', tone)}
          strokeWidth="2.5"
          strokeDasharray={c}
          strokeDashoffset={off}
          strokeLinecap="round"
        />
      </svg>
      <span className="absolute inset-0 flex items-center justify-center font-mono text-[9px] font-semibold tabular-nums">
        {value}
      </span>
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'profile' | 'activity' | 'alerts' | 'groups'

function UserDrawer({ user, onClose }: { user: AuditedUser; onClose: () => void }) {
  const [tab, setTab] = useState<Tab>('profile')
  const risk = RISK_STYLE[user.riskLevel]
  const st = STATUS_META[user.status]
  const kind = KIND_META[user.kind]

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[1000px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <UserAvatar user={user} large />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <kind.icon size={11} className={kind.tone} />
                  <span>{kind.label}</span>
                  <span>·</span>
                  <span>{SOURCE_META[user.source.kind].label}</span>
                  <span>·</span>
                  <span>first seen {relativeTime(user.firstSeen)}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{user.name}</h2>
                <div className="truncate font-mono text-xs text-muted-foreground">{user.username}{user.email && ` · ${user.email}`}</div>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', risk.pill)}>
                    Risk {user.riskScore}
                  </span>
                  <span className="inline-flex items-center gap-1.5 text-muted-foreground">
                    <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                    {st.label}
                  </span>
                  {user.flags.map((f) => (
                    <span key={f} className="rounded-md bg-muted px-1.5 py-0.5">{f}</span>
                  ))}
                </div>
              </div>
              <RiskRing value={user.riskScore} />
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
              <Lock size={14} className="mr-1.5" />
              Disable account
            </Button>
            <Button size="sm" variant="outline">
              <Eye size={14} className="mr-1.5" />
              Watch
            </Button>
            <Button size="sm" variant="outline">
              <EyeOff size={14} className="mr-1.5" />
              Suppress alerts
            </Button>
            <Button size="sm" variant="outline">
              <AlertTriangle size={14} className="mr-1.5 text-amber-500" />
              Open incident
            </Button>
            <Button size="sm" variant="ghost" className="ml-auto">
              <MoreHorizontal size={14} />
            </Button>
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          <DrawerTab id="profile" current={tab} onChange={setTab} icon={User}>Profile</DrawerTab>
          <DrawerTab id="activity" current={tab} onChange={setTab} icon={Activity}>
            Activity
            {user.anomalies.length > 0 && (
              <span className="ml-1.5 rounded bg-red-500/15 px-1.5 py-0.5 font-mono text-[10px] text-red-600 dark:text-red-300">
                {user.anomalies.length} anomalies
              </span>
            )}
          </DrawerTab>
          <DrawerTab id="alerts" current={tab} onChange={setTab} icon={AlertTriangle}>
            Alerts
            {user.recentAlerts.length > 0 && (
              <span className="ml-1.5 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
                {user.recentAlerts.length}
              </span>
            )}
          </DrawerTab>
          <DrawerTab id="groups" current={tab} onChange={setTab} icon={Shield}>Groups</DrawerTab>
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/20">
          {tab === 'profile' && <ProfileTab user={user} />}
          {tab === 'activity' && <ActivityTab user={user} />}
          {tab === 'alerts' && <AlertsTab user={user} />}
          {tab === 'groups' && <GroupsTab user={user} />}
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

/* ─── Profile tab ──────────────────────────────────────────────────────── */

function ProfileTab({ user }: { user: AuditedUser }) {
  const showAi = user.anomalies.length > 0 || user.riskScore >= 60
  return (
    <div className="grid grid-cols-3 gap-4 p-6">
      <div className="col-span-2 space-y-4">
        {showAi && (
          <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
            <div className="mb-2 flex items-center gap-2 text-xs font-medium">
              <Sparkles size={14} className="text-fuchsia-500" />
              SOC AI behavioral note
            </div>
            <p className="text-sm leading-relaxed">
              {user.kind === 'service' ? (
                <>
                  Service account <span className="font-mono">{user.username}</span> shows{' '}
                  <strong>{user.anomalies.length} behavioral anomalies</strong> in the last 24h —
                  unusual for a process account. Activity pattern suggests possible compromise.
                </>
              ) : (
                <>
                  <span className="font-medium">{user.name}</span> has{' '}
                  <strong>{user.anomalies.length} unusual signals</strong> in the last 24h.
                  Consider reviewing recent privilege use and login geography.
                </>
              )}
            </p>
            <div className="mt-3 flex flex-wrap gap-2 text-[11px]">
              <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
                Compare to baseline
              </button>
              <button className="rounded-md border border-border bg-background/60 px-2 py-1 hover:bg-background">
                Hunt similar accounts
              </button>
            </div>
          </div>
        )}

        <Section title="Identity attributes">
          <dl className="grid grid-cols-[180px_1fr] gap-y-2 text-xs">
            {user.attributes.map((a) => (
              <Row key={a.key} k={a.key}>
                <span className="font-mono break-all">{a.value}</span>
              </Row>
            ))}
          </dl>
        </Section>

        <Section title="Last login">
          <div className="grid grid-cols-[120px_1fr] gap-y-2 text-xs">
            <Row k="Time">
              <span className="font-mono">{absTimestamp(user.lastLogin.ts)}</span>
            </Row>
            <Row k="IP">
              <span className="font-mono">{user.lastLogin.ip}</span>
            </Row>
            {user.lastLogin.country && (
              <Row k="Location">
                <span className="inline-flex items-center gap-1.5">
                  {user.lastLogin.countryCode && (
                    <span aria-hidden className="text-sm leading-none">
                      {flagEmoji(user.lastLogin.countryCode)}
                    </span>
                  )}
                  {user.lastLogin.country}
                </span>
              </Row>
            )}
          </div>
        </Section>
      </div>

      <aside className="col-span-1 space-y-4">
        <KeyValueCard
          title="Account"
          rows={[
            ['Source', SOURCE_META[user.source.kind].label],
            ['Domain', user.source.name],
            ['Department', user.department ?? '—'],
            ['Title', user.title ?? '—'],
            ['First seen', relativeTime(user.firstSeen)],
            ['Last seen', relativeTime(user.lastSeen)],
          ]}
        />

        <div className="rounded-lg border border-border bg-card p-4">
          <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
            Risk breakdown
          </div>
          <RiskBars
            authentication={user.riskScore >= 70 ? 0.85 : 0.2}
            privilege={user.kind === 'admin' || user.kind === 'service' ? 0.7 : 0.15}
            anomaly={Math.min(1, user.anomalies.length / 4)}
          />
        </div>
      </aside>
    </div>
  )
}

function RiskBars({
  authentication,
  privilege,
  anomaly,
}: {
  authentication: number
  privilege: number
  anomaly: number
}) {
  const items: [string, number, string][] = [
    ['Authentication', authentication, 'from-red-500 to-rose-500'],
    ['Privilege', privilege, 'from-orange-500 to-amber-500'],
    ['Anomaly', anomaly, 'from-fuchsia-500 to-violet-500'],
  ]
  return (
    <ul className="space-y-2">
      {items.map(([label, v, grad]) => (
        <li key={label}>
          <div className="mb-1 flex items-center justify-between text-[11px]">
            <span>{label}</span>
            <span className="font-mono tabular-nums text-muted-foreground">
              {Math.round(v * 100)}
            </span>
          </div>
          <div className="h-1.5 overflow-hidden rounded-full bg-muted">
            <div
              className={cn('h-full rounded-full bg-gradient-to-r', grad)}
              style={{ width: `${v * 100}%` }}
            />
          </div>
        </li>
      ))}
    </ul>
  )
}

/* ─── Activity tab ─────────────────────────────────────────────────────── */

function ActivityTab({ user }: { user: AuditedUser }) {
  return (
    <div className="space-y-4 p-6">
      <Section title="Activity over 24 hours">
        <ActivityChart data={user.hourlyActivity} />
      </Section>

      {user.anomalies.length > 0 && (
        <Section title="Detected anomalies">
          <ul className="space-y-2">
            {user.anomalies.map((a, i) => (
              <li
                key={i}
                className="flex items-start gap-3 rounded-lg border border-red-500/20 bg-red-500/5 p-3 text-xs"
              >
                <AlertTriangle size={13} className="mt-0.5 shrink-0 text-red-500" />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-[10px] uppercase tracking-wider text-red-600 dark:text-red-300">
                      {a.kind}
                    </span>
                    <span className="text-muted-foreground">{relativeTime(a.ts)}</span>
                  </div>
                  <div className="mt-1">{a.detail}</div>
                </div>
              </li>
            ))}
          </ul>
        </Section>
      )}

      <Section title="Recent events">
        <div className="text-sm text-muted-foreground">
          <span className="font-mono text-foreground">{user.recentEvents.toLocaleString()}</span>{' '}
          events recorded in the last 24h.{' '}
          <button className="text-primary hover:underline">Open in log explorer</button>
        </div>
      </Section>
    </div>
  )
}

function ActivityChart({ data }: { data: number[] }) {
  const w = 600
  const h = 100
  const max = Math.max(...data, 1)
  const bw = w / data.length
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-24 w-full">
      {[0.5, 1].map((p) => (
        <line
          key={p}
          x1="0"
          x2={w}
          y1={h - h * p * 0.92 - 4}
          y2={h - h * p * 0.92 - 4}
          className="stroke-border/40"
          strokeDasharray="2 4"
        />
      ))}
      {data.map((v, i) => {
        const bh = (v / max) * (h - 16)
        return (
          <rect
            key={i}
            x={i * bw + 1}
            y={h - bh - 12}
            width={bw - 2}
            height={bh}
            rx={1}
            className={v > max * 0.7 ? 'fill-red-500/80' : 'fill-sky-500/60'}
          />
        )
      })}
      <g className="fill-muted-foreground" fontSize="9">
        {[0, 6, 12, 18, 23].map((i) => (
          <text key={i} x={i * bw + bw / 2} y={h - 2} textAnchor="middle">
            {`${String(i).padStart(2, '0')}h`}
          </text>
        ))}
      </g>
    </svg>
  )
}

/* ─── Alerts tab ───────────────────────────────────────────────────────── */

function AlertsTab({ user }: { user: AuditedUser }) {
  if (user.recentAlerts.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center p-6">
        <div className="text-center">
          <div className="text-sm text-muted-foreground">No alerts in the last 24 hours.</div>
          <div className="mt-1 text-[11px] text-muted-foreground">
            <button className="text-primary hover:underline">View 30-day history</button>
          </div>
        </div>
      </div>
    )
  }
  const sevTone: Record<'critical' | 'high' | 'medium' | 'low', string> = {
    critical: 'bg-red-500',
    high: 'bg-orange-500',
    medium: 'bg-amber-500',
    low: 'bg-sky-500',
  }
  return (
    <div className="p-6">
      <ul className="overflow-hidden rounded-lg border border-border bg-card">
        {user.recentAlerts.map((a) => (
          <li
            key={a.id}
            className="flex cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
          >
            <span className={cn('h-3 w-1 shrink-0 rounded-full', sevTone[a.severity])} />
            <div className="min-w-0 flex-1">
              <div className="truncate font-medium">{a.name}</div>
              <div className="mt-0.5 flex items-center gap-2 text-[11px] text-muted-foreground">
                <span className="font-mono">{a.id}</span>
                <span>·</span>
                <span className="capitalize">{a.severity}</span>
              </div>
            </div>
            <div className="font-mono text-[11px] text-muted-foreground">
              {relativeTime(a.ts)}
            </div>
            <ExternalLink size={12} className="shrink-0 text-muted-foreground" />
          </li>
        ))}
      </ul>
    </div>
  )
}

/* ─── Groups tab ───────────────────────────────────────────────────────── */

function GroupsTab({ user }: { user: AuditedUser }) {
  return (
    <div className="space-y-4 p-6">
      <Section title="Group memberships">
        <ul className="space-y-1.5">
          {user.groups.map((g) => {
            const privileged = /admin|operator|root/i.test(g)
            return (
              <li
                key={g}
                className={cn(
                  'flex items-center justify-between rounded-md border px-3 py-2 text-xs',
                  privileged
                    ? 'border-amber-500/30 bg-amber-500/5'
                    : 'border-border bg-background/40'
                )}
              >
                <div className="flex items-center gap-2">
                  {privileged ? (
                    <Crown size={12} className="text-amber-500" />
                  ) : (
                    <Shield size={12} className="text-muted-foreground" />
                  )}
                  <span className="font-mono">{g}</span>
                </div>
                {privileged && (
                  <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
                    privileged
                  </span>
                )}
              </li>
            )
          })}
        </ul>
      </Section>
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

function flagEmoji(cc: string) {
  if (!cc || cc.length !== 2) return ''
  const A = 0x1f1e6
  return String.fromCodePoint(
    A - 65 + cc.toUpperCase().charCodeAt(0),
    A - 65 + cc.toUpperCase().charCodeAt(1),
  )
}

// Suppress potential unused warning
void Globe2
void UserX
void MapPin
