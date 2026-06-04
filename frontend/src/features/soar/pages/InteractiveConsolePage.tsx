import { useEffect, useMemo, useRef, useState } from 'react'
import {
  Activity,
  AlertTriangle,
  Bookmark,
  ChevronDown,
  ChevronRight,
  Circle,
  CircleDot,
  Code2,
  Copy,
  Disc,
  Download,
  Eye,
  EyeOff,
  FileText,
  History,
  Keyboard,
  Lock,
  Monitor,
  MoreHorizontal,
  Network,
  Plus,
  Search,
  Send,
  Server,
  Shield,
  Sparkles,
  Square,
  Star,
  Terminal as TerminalIcon,
  Variable,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type AgentStatus = 'online' | 'offline' | 'idle'
type OSKind = 'windows' | 'linux' | 'macos'
type Shell = 'cmd' | 'powershell' | 'bash' | 'zsh'

interface Agent {
  id: string
  hostname: string
  os: OSKind
  osVersion: string
  ip: string
  status: AgentStatus
  lastSeen: string
  riskScore: number
  group: string
  starred?: boolean
}

interface Snippet {
  id: string
  name: string
  description: string
  os: OSKind[]
  command: string
  category: 'discovery' | 'forensic' | 'network' | 'persistence' | 'response'
}

type LineKind = 'system' | 'prompt' | 'stdout' | 'stderr' | 'ai'

interface Line {
  kind: LineKind
  text: string
  ts?: string
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()

const AGENTS: Agent[] = [
  { id: 'WIN-7F3K2L9Q8X', hostname: 'WIN-7F3K2L9Q8X', os: 'windows', osVersion: 'Server 2022', ip: '10.4.18.22', status: 'online', lastSeen: min(0), riskScore: 96, group: 'Finance', starred: true },
  { id: 'fin-app-01', hostname: 'fin-app-01', os: 'linux', osVersion: 'Ubuntu 22.04', ip: '10.0.4.91', status: 'online', lastSeen: min(0), riskScore: 88, group: 'Finance' },
  { id: 'mail-srv-02', hostname: 'mail-srv-02', os: 'linux', osVersion: 'RHEL 9', ip: '10.2.1.18', status: 'online', lastSeen: min(0), riskScore: 71, group: 'IT' },
  { id: 'dc-01', hostname: 'dc-01', os: 'windows', osVersion: 'Server 2022', ip: '10.4.18.5', status: 'online', lastSeen: min(0), riskScore: 64, group: 'IT' },
  { id: 'fileshare-01', hostname: 'fileshare-01', os: 'linux', osVersion: 'Debian 12', ip: '10.4.18.7', status: 'online', lastSeen: min(2), riskScore: 41, group: 'IT' },
  { id: 'webapp-01', hostname: 'webapp-01', os: 'linux', osVersion: 'Ubuntu 22.04', ip: '10.0.4.110', status: 'idle', lastSeen: min(8), riskScore: 32, group: 'Web' },
  { id: 'sql-prod-01', hostname: 'sql-prod-01', os: 'windows', osVersion: 'Server 2019', ip: '10.0.4.45', status: 'online', lastSeen: min(0), riskScore: 28, group: 'DB' },
  { id: 'edge-fw-bsa-01', hostname: 'edge-fw-bsa-01', os: 'linux', osVersion: 'pfSense 2.7', ip: '10.0.0.1', status: 'online', lastSeen: min(0), riskScore: 12, group: 'Network' },
  { id: 'macbook-jsmith', hostname: 'macbook-jsmith', os: 'macos', osVersion: 'Sonoma 14.4', ip: '10.6.4.18', status: 'idle', lastSeen: min(15), riskScore: 8, group: 'Workstations' },
  { id: 'ws-finance-04', hostname: 'ws-finance-04', os: 'windows', osVersion: 'Win 11', ip: '10.6.4.32', status: 'offline', lastSeen: min(180), riskScore: 4, group: 'Workstations' },
]

const SNIPPETS: Snippet[] = [
  {
    id: 'snip-procs',
    name: 'List running processes',
    description: 'Show all processes with CPU and PID',
    os: ['windows', 'linux', 'macos'],
    command: 'tasklist /v',
    category: 'discovery',
  },
  {
    id: 'snip-net',
    name: 'Active network connections',
    description: 'TCP/UDP sockets with owning process',
    os: ['windows', 'linux'],
    command: 'netstat -anob',
    category: 'network',
  },
  {
    id: 'snip-users',
    name: 'Logged-in users',
    description: 'Active sessions on the host',
    os: ['windows', 'linux', 'macos'],
    command: 'query user',
    category: 'discovery',
  },
  {
    id: 'snip-tasks',
    name: 'Scheduled tasks',
    description: 'Persistence vector — list scheduled jobs',
    os: ['windows'],
    command: 'schtasks /query /fo LIST /v',
    category: 'persistence',
  },
  {
    id: 'snip-services',
    name: 'Service status',
    description: 'Windows services and start type',
    os: ['windows'],
    command: 'sc queryex type=service state=all',
    category: 'discovery',
  },
  {
    id: 'snip-isolate',
    name: 'Isolate from network',
    description: 'Block all outbound except panel',
    os: ['windows', 'linux'],
    command: 'New-NetFirewallRule -DisplayName "UTMStack-Isolate" -Direction Outbound -Action Block',
    category: 'response',
  },
  {
    id: 'snip-mem',
    name: 'Capture memory dump',
    description: 'Forensic capture for offline analysis',
    os: ['windows'],
    command: 'winpmem -o C:\\Forensics\\${HOST}-mem.aff4',
    category: 'forensic',
  },
  {
    id: 'snip-kerb',
    name: 'Inspect Kerberos tickets',
    description: 'List cached TGT/TGS — useful for Kerberoasting analysis',
    os: ['windows'],
    command: 'klist',
    category: 'forensic',
  },
]

const VARIABLES: { name: string; resolved: string; description: string }[] = [
  { name: '${HOST}', resolved: 'WIN-7F3K2L9Q8X', description: 'Selected agent hostname' },
  { name: '${IP}', resolved: '10.4.18.22', description: 'Agent primary IP' },
  { name: '${OS}', resolved: 'windows', description: 'Agent OS family' },
  { name: '${INCIDENT_ID}', resolved: '#1142', description: 'Active incident in context' },
  { name: '${ALERT_ID}', resolved: 'a-9f04c2', description: 'Linked alert' },
  { name: '${ANALYST}', resolved: 'jsmith@utmstack.com', description: 'Current operator' },
  { name: '${TS}', resolved: '2026-05-01T16:42:00Z', description: 'Now (UTC)' },
]

const HISTORY: { ts: string; cmd: string; agent: string }[] = [
  { ts: min(2), cmd: 'tasklist /v | findstr lsass', agent: 'WIN-7F3K2L9Q8X' },
  { ts: min(4), cmd: 'klist', agent: 'WIN-7F3K2L9Q8X' },
  { ts: min(8), cmd: 'Get-WmiObject -Class Win32_Process | Where { $_.Name -eq "powershell.exe" }', agent: 'WIN-7F3K2L9Q8X' },
  { ts: min(15), cmd: 'netstat -anob', agent: 'WIN-7F3K2L9Q8X' },
  { ts: min(34), cmd: 'sudo ss -tnp', agent: 'fin-app-01' },
  { ts: min(60), cmd: 'systemctl status sshd', agent: 'mail-srv-02' },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const STATUS_META: Record<AgentStatus, { dot: string; label: string; tone: string }> = {
  online: { dot: 'bg-emerald-500', label: 'Online', tone: 'text-emerald-500' },
  idle: { dot: 'bg-amber-500', label: 'Idle', tone: 'text-amber-500' },
  offline: { dot: 'bg-muted-foreground', label: 'Offline', tone: 'text-muted-foreground' },
}

const OS_ICON: Record<OSKind, LucideIcon> = {
  windows: Server,
  linux: TerminalIcon,
  macos: Monitor,
}

const OS_TONE: Record<OSKind, string> = {
  windows: 'text-sky-500',
  linux: 'text-amber-500',
  macos: 'text-violet-500',
}

const CATEGORY_TONE: Record<Snippet['category'], string> = {
  discovery: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300',
  forensic: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300',
  network: 'bg-fuchsia-500/15 text-fuchsia-600 ring-fuchsia-500/30 dark:text-fuchsia-300',
  persistence: 'bg-orange-500/15 text-orange-600 ring-orange-500/30 dark:text-orange-300',
  response: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300',
}

/* ─── Initial console state ────────────────────────────────────────────── */

interface Session {
  agentId: string
  shell: Shell
  authed: boolean
  recording: boolean
  lines: Line[]
}

function shellsFor(os: OSKind): Shell[] {
  if (os === 'windows') return ['powershell', 'cmd']
  if (os === 'macos') return ['zsh', 'bash']
  return ['bash', 'zsh']
}

function newSession(a: Agent): Session {
  const shell = shellsFor(a.os)[0]
  return {
    agentId: a.id,
    shell,
    authed: a.id === 'WIN-7F3K2L9Q8X',
    recording: true,
    lines: a.id === 'WIN-7F3K2L9Q8X' ? STARTER_LINES : [],
  }
}

const STARTER_LINES: Line[] = [
  { kind: 'system', text: 'UTMStack Interactive Console · session opened', ts: min(8) },
  { kind: 'system', text: 'Connected to WIN-7F3K2L9Q8X (Windows Server 2022, 10.4.18.22)', ts: min(8) },
  { kind: 'system', text: 'Session is being recorded for audit. Operator: jsmith@utmstack.com', ts: min(8) },
  { kind: 'ai', text: 'This host has 3 active critical alerts. Suggested first command: `klist` to inspect Kerberos tickets.', ts: min(8) },
  { kind: 'prompt', text: 'klist', ts: min(7) },
  { kind: 'stdout', text: 'Current LogonId is 0:0x3e7' },
  { kind: 'stdout', text: 'Cached Tickets: (4)' },
  { kind: 'stdout', text: '#0> Client: svc_backup @ CORP.LOCAL' },
  { kind: 'stdout', text: '    Server: krbtgt/CORP.LOCAL @ CORP.LOCAL' },
  { kind: 'stdout', text: '    KerbTicket Encryption Type: AES-256-CTS-HMAC-SHA1-96' },
  { kind: 'stdout', text: '#1> Client: svc_backup @ CORP.LOCAL' },
  { kind: 'stdout', text: '    Server: MSSQLSvc/sql-prod-01 @ CORP.LOCAL' },
  { kind: 'stdout', text: '    KerbTicket Encryption Type: RC4-HMAC-NT' },
  { kind: 'stderr', text: '⚠ RC4 ticket detected — common Kerberoasting indicator.' },
  { kind: 'prompt', text: 'tasklist /v | findstr lsass', ts: min(2) },
  { kind: 'stdout', text: 'lsass.exe                     764 Services       0    62,124 K  Local System' },
  { kind: 'system', text: 'process inspection complete · 1 match', ts: min(2) },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function InteractiveConsolePage() {
  const [openSessions, setOpenSessions] = useState<Session[]>([newSession(AGENTS[0])])
  const [activeAgent, setActiveAgent] = useState<string>(AGENTS[0].id)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | AgentStatus>('all')
  const [showRight, setShowRight] = useState(true)

  const session = openSessions.find((s) => s.agentId === activeAgent) ?? null
  const agent = AGENTS.find((a) => a.id === activeAgent) ?? null

  const filteredAgents = useMemo(() => {
    return AGENTS.filter((a) =>
      search ? (a.hostname + a.ip + a.osVersion + a.group).toLowerCase().includes(search.toLowerCase()) : true
    ).filter((a) => (statusFilter === 'all' ? true : a.status === statusFilter))
  }, [search, statusFilter])

  const grouped = useMemo(() => {
    const map = new Map<string, Agent[]>()
    filteredAgents.forEach((a) => {
      const arr = map.get(a.group) ?? []
      arr.push(a)
      map.set(a.group, arr)
    })
    return Array.from(map.entries())
  }, [filteredAgents])

  const openSessionFor = (a: Agent) => {
    setActiveAgent(a.id)
    setOpenSessions((prev) => (prev.find((s) => s.agentId === a.id) ? prev : [...prev, newSession(a)]))
  }

  const closeSession = (id: string) => {
    setOpenSessions((prev) => {
      const next = prev.filter((s) => s.agentId !== id)
      if (next.length > 0 && id === activeAgent) setActiveAgent(next[0].agentId)
      return next
    })
  }

  const updateSession = (id: string, patch: Partial<Session>) => {
    setOpenSessions((prev) => prev.map((s) => (s.agentId === id ? { ...s, ...patch } : s)))
  }

  return (
    <div className="flex h-full w-full flex-col">
      <PageHeader sessionCount={openSessions.length} showRight={showRight} onToggleRight={() => setShowRight((s) => !s)} />

      <div className="grid min-h-0 flex-1 gap-3 px-6 pb-6" style={{ gridTemplateColumns: showRight ? '280px 1fr 320px' : '280px 1fr' }}>
        <AgentSidebar
          search={search}
          onSearch={setSearch}
          statusFilter={statusFilter}
          onStatusFilter={setStatusFilter}
          grouped={grouped}
          activeAgent={activeAgent}
          openSessions={openSessions}
          onSelect={openSessionFor}
        />

        <ConsoleArea
          openSessions={openSessions}
          activeAgent={activeAgent}
          setActiveAgent={setActiveAgent}
          closeSession={closeSession}
          session={session}
          agent={agent}
          updateSession={updateSession}
        />

        {showRight && (
          <RightPanel
            onInsertSnippet={(s) => {
              if (!session || !agent) return
              updateSession(session.agentId, {
                lines: [...session.lines, { kind: 'system', text: `Snippet "${s.name}" inserted into prompt`, ts: new Date().toISOString() }],
              })
            }}
          />
        )}
      </div>
    </div>
  )
}

/* ─── Page header ──────────────────────────────────────────────────────── */

function PageHeader({
  sessionCount,
  showRight,
  onToggleRight,
}: {
  sessionCount: number
  showRight: boolean
  onToggleRight: () => void
}) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3 px-6 py-4">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Interactive console</h1>
          <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
            {sessionCount} session{sessionCount === 1 ? '' : 's'}
          </span>
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          Direct command-line access to remote agents. Every keystroke is recorded for audit.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <History size={14} className="mr-2" />
          Audit log
        </Button>
        <Button variant="outline" size="sm">
          <Download size={14} className="mr-2" />
          Export session
        </Button>
        <Button variant="outline" size="sm" onClick={onToggleRight}>
          {showRight ? <EyeOff size={14} className="mr-2" /> : <Eye size={14} className="mr-2" />}
          {showRight ? 'Hide panel' : 'Show panel'}
        </Button>
      </div>
    </header>
  )
}

/* ─── Agent sidebar ────────────────────────────────────────────────────── */

function AgentSidebar({
  search,
  onSearch,
  statusFilter,
  onStatusFilter,
  grouped,
  activeAgent,
  openSessions,
  onSelect,
}: {
  search: string
  onSearch: (v: string) => void
  statusFilter: 'all' | AgentStatus
  onStatusFilter: (v: 'all' | AgentStatus) => void
  grouped: [string, Agent[]][]
  activeAgent: string
  openSessions: Session[]
  onSelect: (a: Agent) => void
}) {
  const openIds = new Set(openSessions.map((s) => s.agentId))
  return (
    <aside className="flex min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-card">
      <div className="border-b border-border p-3">
        <div className="relative">
          <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search hostname, IP, group…"
            value={search}
            onChange={(e) => onSearch(e.target.value)}
            className="h-8 pl-8 text-xs"
          />
        </div>
        <div className="mt-2 flex items-center gap-1">
          {(['all', 'online', 'idle', 'offline'] as const).map((s) => (
            <button
              key={s}
              onClick={() => onStatusFilter(s)}
              className={cn(
                'rounded px-2 py-1 text-[10px] capitalize transition-colors',
                statusFilter === s ? 'bg-primary/15 text-primary' : 'text-muted-foreground hover:bg-muted'
              )}
            >
              {s}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {grouped.map(([group, items]) => (
          <div key={group} className="mb-3">
            <div className="mb-1 flex items-center justify-between px-2">
              <span className="text-[10px] uppercase tracking-wider text-muted-foreground">{group}</span>
              <span className="font-mono text-[10px] text-muted-foreground">{items.length}</span>
            </div>
            <div className="space-y-0.5">
              {items.map((a) => (
                <AgentItem
                  key={a.id}
                  agent={a}
                  active={activeAgent === a.id}
                  open={openIds.has(a.id)}
                  onClick={() => onSelect(a)}
                />
              ))}
            </div>
          </div>
        ))}
        {grouped.length === 0 && (
          <div className="px-3 py-6 text-center text-[11px] italic text-muted-foreground">No agents match.</div>
        )}
      </div>
    </aside>
  )
}

function AgentItem({
  agent,
  active,
  open,
  onClick,
}: {
  agent: Agent
  active: boolean
  open: boolean
  onClick: () => void
}) {
  const st = STATUS_META[agent.status]
  const OsIcon = OS_ICON[agent.os]
  return (
    <button
      onClick={onClick}
      className={cn(
        'group flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors',
        active ? 'bg-primary/10 ring-1 ring-primary/30' : 'hover:bg-muted/60'
      )}
    >
      <span className={cn('relative h-2 w-2 shrink-0 rounded-full', st.dot)}>
        {agent.status === 'online' && (
          <span className={cn('absolute inset-0 animate-ping rounded-full opacity-50', st.dot)} />
        )}
      </span>
      <OsIcon size={12} className={cn('shrink-0', OS_TONE[agent.os])} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1">
          <span className={cn('truncate font-mono text-[11px] font-medium', active && 'text-primary')}>
            {agent.hostname}
          </span>
          {agent.starred && <Star size={9} className="fill-amber-400 text-amber-400" />}
        </div>
        <div className="flex items-center gap-1 text-[10px] text-muted-foreground">
          <span className="font-mono">{agent.ip}</span>
          <span>·</span>
          <span className="truncate">{agent.osVersion}</span>
        </div>
      </div>
      {open && <CircleDot size={10} className="shrink-0 text-primary" />}
      {agent.riskScore >= 70 && (
        <span className="shrink-0 rounded bg-red-500/15 px-1 py-0.5 font-mono text-[8px] font-medium text-red-600 ring-1 ring-red-500/30 dark:text-red-300">
          {agent.riskScore}
        </span>
      )}
    </button>
  )
}

/* ─── Console area ─────────────────────────────────────────────────────── */

function ConsoleArea({
  openSessions,
  activeAgent,
  setActiveAgent,
  closeSession,
  session,
  agent,
  updateSession,
}: {
  openSessions: Session[]
  activeAgent: string
  setActiveAgent: (id: string) => void
  closeSession: (id: string) => void
  session: Session | null
  agent: Agent | null
  updateSession: (id: string, patch: Partial<Session>) => void
}) {
  if (!session || !agent) {
    return <EmptyState />
  }
  return (
    <div className="flex min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-card">
      <SessionTabs
        sessions={openSessions}
        activeAgent={activeAgent}
        onSelect={setActiveAgent}
        onClose={closeSession}
      />
      <AgentBanner agent={agent} session={session} updateSession={updateSession} />
      {session.authed ? (
        <Terminal session={session} agent={agent} updateSession={updateSession} />
      ) : (
        <AuthGate session={session} updateSession={updateSession} />
      )}
    </div>
  )
}

function SessionTabs({
  sessions,
  activeAgent,
  onSelect,
  onClose,
}: {
  sessions: Session[]
  activeAgent: string
  onSelect: (id: string) => void
  onClose: (id: string) => void
}) {
  return (
    <div className="flex items-center border-b border-border bg-muted/30">
      <div className="flex flex-1 items-center overflow-x-auto">
        {sessions.map((s) => {
          const a = AGENTS.find((x) => x.id === s.agentId)!
          const active = activeAgent === s.agentId
          return (
            <div
              key={s.agentId}
              onClick={() => onSelect(s.agentId)}
              className={cn(
                'group flex shrink-0 cursor-pointer items-center gap-2 border-r border-border px-3 py-2 text-xs',
                active ? 'bg-card text-foreground' : 'text-muted-foreground hover:bg-muted/40'
              )}
            >
              <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_META[a.status].dot)} />
              <span className="font-mono">{a.hostname}</span>
              {s.recording && <Disc size={10} className="text-red-500 animate-pulse" />}
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onClose(s.agentId)
                }}
                className="ml-1 flex h-4 w-4 items-center justify-center rounded hover:bg-muted"
              >
                <X size={11} />
              </button>
            </div>
          )
        })}
      </div>
      <button className="flex h-9 w-9 shrink-0 items-center justify-center text-muted-foreground hover:bg-muted hover:text-foreground" title="New session">
        <Plus size={14} />
      </button>
    </div>
  )
}

function AgentBanner({
  agent,
  session,
  updateSession,
}: {
  agent: Agent
  session: Session
  updateSession: (id: string, patch: Partial<Session>) => void
}) {
  const st = STATUS_META[agent.status]
  const OsIcon = OS_ICON[agent.os]
  return (
    <div className="flex flex-wrap items-center gap-3 border-b border-border bg-card px-4 py-3 text-xs">
      <div className="flex items-center gap-2">
        <span className={cn('h-2 w-2 rounded-full', st.dot)} />
        <OsIcon size={14} className={OS_TONE[agent.os]} />
        <span className="font-mono text-sm font-semibold">{agent.hostname}</span>
      </div>
      <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <Network size={11} /> {agent.ip}
        </span>
        <span className="inline-flex items-center gap-1">
          <Server size={11} /> {agent.osVersion}
        </span>
        <span className="inline-flex items-center gap-1">
          <Activity size={11} /> last seen {relativeTime(agent.lastSeen)}
        </span>
      </div>
      {agent.riskScore >= 70 && (
        <span className="inline-flex items-center gap-1 rounded bg-red-500/15 px-1.5 py-0.5 text-[10px] font-medium text-red-600 ring-1 ring-red-500/30 dark:text-red-300">
          <AlertTriangle size={10} /> Risk {agent.riskScore}
        </span>
      )}
      <div className="ml-auto flex items-center gap-1">
        <ShellSelect
          os={agent.os}
          shell={session.shell}
          onChange={(s) => updateSession(session.agentId, { shell: s })}
        />
        <button
          onClick={() => updateSession(session.agentId, { recording: !session.recording })}
          className={cn(
            'flex h-7 items-center gap-1.5 rounded-md border border-border px-2 text-[11px]',
            session.recording ? 'text-red-600 dark:text-red-300' : 'text-muted-foreground'
          )}
        >
          {session.recording ? <Disc size={11} className="animate-pulse" /> : <Square size={11} />}
          {session.recording ? 'Recording' : 'Paused'}
        </button>
        <Button size="sm" variant="outline">
          <Shield size={12} className="mr-1.5" />
          Disconnect
        </Button>
        <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" title="More">
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}

function ShellSelect({ os, shell, onChange }: { os: OSKind; shell: Shell; onChange: (s: Shell) => void }) {
  const shells = shellsFor(os)
  return (
    <div className="flex items-center gap-1 rounded-md border border-border bg-background/40 p-0.5">
      {shells.map((s) => (
        <button
          key={s}
          onClick={() => onChange(s)}
          className={cn(
            'rounded px-2 py-1 text-[10px] transition-colors',
            shell === s ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'
          )}
        >
          {s}
        </button>
      ))}
    </div>
  )
}

function AuthGate({
  session,
  updateSession,
}: {
  session: Session
  updateSession: (id: string, patch: Partial<Session>) => void
}) {
  const [pw, setPw] = useState('')
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 bg-[#0b0d12] p-6 font-mono text-sm text-zinc-200">
      <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-center text-xs">
        <Lock size={14} className="mx-auto mb-1 text-amber-400" />
        <div className="font-medium text-amber-100">Privileged session</div>
        <div className="mt-0.5 text-amber-100/70">
          Re-authenticate with your panel password to open a remote shell.
        </div>
      </div>
      <div className="w-full max-w-sm">
        <div className="flex items-center gap-2 rounded-md bg-zinc-900 px-3 py-2">
          <span className="text-emerald-400">$</span>
          <span className="text-zinc-400">password:</span>
          <input
            type="password"
            value={pw}
            onChange={(e) => setPw(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                updateSession(session.agentId, { authed: true, lines: [...STARTER_LINES] })
              }
            }}
            placeholder="••••••••"
            className="flex-1 bg-transparent text-zinc-100 outline-none placeholder:text-zinc-600"
          />
        </div>
        <div className="mt-2 text-center text-[10px] text-zinc-500">
          Press Enter to authenticate
        </div>
      </div>
    </div>
  )
}

/* ─── Terminal ─────────────────────────────────────────────────────────── */

function Terminal({
  session,
  agent,
  updateSession,
}: {
  session: Session
  agent: Agent
  updateSession: (id: string, patch: Partial<Session>) => void
}) {
  const [cmd, setCmd] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const outRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [session.agentId])

  useEffect(() => {
    outRef.current?.scrollTo({ top: outRef.current.scrollHeight })
  }, [session.lines])

  const submit = () => {
    if (!cmd.trim()) return
    const newLines: Line[] = [
      ...session.lines,
      { kind: 'prompt', text: cmd, ts: new Date().toISOString() },
      { kind: 'stdout', text: 'Output streaming…' },
    ]
    updateSession(session.agentId, { lines: newLines })
    setCmd('')
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col bg-[#0b0d12] font-mono text-[12px] text-zinc-200">
      <div ref={outRef} className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
        {session.lines.map((l, i) => (
          <TerminalLine key={i} line={l} hostname={agent.hostname} shell={session.shell} />
        ))}
      </div>
      <div className="flex items-center gap-2 border-t border-zinc-800 bg-zinc-950/40 px-4 py-2">
        <Prompt agent={agent} shell={session.shell} />
        <input
          ref={inputRef}
          value={cmd}
          onChange={(e) => setCmd(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') submit()
          }}
          spellCheck={false}
          placeholder="Type a command · Tab for variables · ↑ for history"
          className="flex-1 bg-transparent text-zinc-100 outline-none placeholder:text-zinc-600"
        />
        <button onClick={submit} className="flex h-6 w-6 items-center justify-center rounded bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30">
          <Send size={11} />
        </button>
      </div>
      <div className="flex items-center justify-between border-t border-zinc-800 bg-zinc-950 px-4 py-1.5 text-[10px] text-zinc-500">
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-1">
            <Keyboard size={10} /> Tab
            <span className="text-zinc-600">vars</span>
          </span>
          <span className="inline-flex items-center gap-1">
            <Keyboard size={10} /> ↑↓
            <span className="text-zinc-600">history</span>
          </span>
          <span className="inline-flex items-center gap-1">
            <Keyboard size={10} /> Ctrl+L
            <span className="text-zinc-600">clear</span>
          </span>
        </div>
        <button className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-fuchsia-400 hover:bg-zinc-900">
          <Sparkles size={10} />
          Ask AI for a command
        </button>
      </div>
    </div>
  )
}

function Prompt({ agent, shell }: { agent: Agent; shell: Shell }) {
  return (
    <span className="shrink-0 select-none whitespace-nowrap font-mono text-[12px]">
      <span className="text-emerald-400">jsmith@{agent.hostname}</span>
      <span className="text-zinc-500"> · </span>
      <span className="text-fuchsia-400">{shell}</span>
      <span className="text-emerald-400"> ❯</span>
    </span>
  )
}

function TerminalLine({ line, hostname, shell }: { line: Line; hostname: string; shell: Shell }) {
  if (line.kind === 'prompt') {
    return (
      <div className="mb-0.5 flex items-baseline gap-2">
        <span className="shrink-0 select-none whitespace-nowrap text-[12px]">
          <span className="text-emerald-400">jsmith@{hostname}</span>
          <span className="text-zinc-500"> · </span>
          <span className="text-fuchsia-400">{shell}</span>
          <span className="text-emerald-400"> ❯</span>
        </span>
        <span className="text-zinc-100">{line.text}</span>
      </div>
    )
  }
  if (line.kind === 'system') {
    return (
      <div className="mb-0.5 flex items-baseline gap-2 text-zinc-500">
        <span className="shrink-0 text-zinc-600">{line.ts ? formatTs(line.ts) : ''}</span>
        <span className="text-sky-400">[sys]</span>
        <span>{line.text}</span>
      </div>
    )
  }
  if (line.kind === 'ai') {
    return (
      <div className="mb-1 flex items-start gap-2 rounded border border-fuchsia-500/30 bg-fuchsia-500/5 px-2 py-1.5 text-fuchsia-100">
        <Sparkles size={11} className="mt-0.5 shrink-0 text-fuchsia-400" />
        <span>{line.text}</span>
      </div>
    )
  }
  if (line.kind === 'stderr') {
    return <div className="mb-0.5 pl-2 text-amber-300">{line.text}</div>
  }
  return <div className="mb-0.5 whitespace-pre-wrap pl-2 text-zinc-200">{line.text}</div>
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-card p-12 text-center">
      <TerminalIcon size={36} className="mb-3 text-muted-foreground" strokeWidth={1.5} />
      <h4 className="text-sm font-semibold">No agent selected</h4>
      <p className="mt-1 max-w-sm text-xs text-muted-foreground">
        Pick an agent from the left sidebar to open an interactive shell. All actions are recorded
        for audit.
      </p>
    </div>
  )
}

/* ─── Right panel ──────────────────────────────────────────────────────── */

type RightTab = 'snippets' | 'variables' | 'history' | 'files'

function RightPanel({ onInsertSnippet }: { onInsertSnippet: (s: Snippet) => void }) {
  const [tab, setTab] = useState<RightTab>('snippets')
  return (
    <aside className="flex min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-card">
      <div className="flex items-center border-b border-border">
        <RightTabBtn id="snippets" current={tab} onChange={setTab} icon={Bookmark}>
          Snippets
        </RightTabBtn>
        <RightTabBtn id="variables" current={tab} onChange={setTab} icon={Variable}>
          Vars
        </RightTabBtn>
        <RightTabBtn id="history" current={tab} onChange={setTab} icon={History}>
          History
        </RightTabBtn>
        <RightTabBtn id="files" current={tab} onChange={setTab} icon={FileText}>
          Files
        </RightTabBtn>
      </div>
      <div className="flex-1 overflow-y-auto">
        {tab === 'snippets' && <SnippetsPanel onInsert={onInsertSnippet} />}
        {tab === 'variables' && <VariablesPanel />}
        {tab === 'history' && <HistoryPanel />}
        {tab === 'files' && <FilesPanel />}
      </div>
    </aside>
  )
}

function RightTabBtn({
  id,
  current,
  onChange,
  icon: Icon,
  children,
}: {
  id: RightTab
  current: RightTab
  onChange: (id: RightTab) => void
  icon: LucideIcon
  children: React.ReactNode
}) {
  const active = id === current
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative flex flex-1 items-center justify-center gap-1.5 px-2 py-2.5 text-xs transition-colors',
        active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
      )}
    >
      <Icon size={12} />
      {children}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

function SnippetsPanel({ onInsert }: { onInsert: (s: Snippet) => void }) {
  const [q, setQ] = useState('')
  const [openCat, setOpenCat] = useState<Snippet['category'] | null>('discovery')
  const filtered = SNIPPETS.filter((s) => (q ? (s.name + s.description + s.command).toLowerCase().includes(q.toLowerCase()) : true))
  const cats = Array.from(new Set(filtered.map((s) => s.category)))
  return (
    <div>
      <div className="border-b border-border p-2">
        <div className="relative">
          <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Filter snippets…" className="h-8 pl-7 text-[11px]" />
        </div>
      </div>
      <div className="p-2">
        {cats.map((c) => (
          <div key={c} className="mb-2">
            <button
              onClick={() => setOpenCat(openCat === c ? null : c)}
              className="mb-1 flex w-full items-center gap-1 px-1.5 py-1 text-[10px] uppercase tracking-wider text-muted-foreground hover:text-foreground"
            >
              {openCat === c ? <ChevronDown size={10} /> : <ChevronRight size={10} />}
              <span>{c}</span>
              <span className="ml-auto rounded bg-muted px-1.5 py-0.5 font-mono text-muted-foreground">
                {filtered.filter((s) => s.category === c).length}
              </span>
            </button>
            {openCat === c && (
              <div className="space-y-1.5">
                {filtered.filter((s) => s.category === c).map((s) => (
                  <SnippetCard key={s.id} snippet={s} onInsert={() => onInsert(s)} />
                ))}
              </div>
            )}
          </div>
        ))}
        {filtered.length === 0 && (
          <div className="px-3 py-6 text-center text-[11px] italic text-muted-foreground">
            No snippets match.
          </div>
        )}
      </div>
    </div>
  )
}

function SnippetCard({ snippet, onInsert }: { snippet: Snippet; onInsert: () => void }) {
  return (
    <div className="rounded-md border border-border bg-background/40 p-2.5 text-xs">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="font-medium">{snippet.name}</div>
          <div className="mt-0.5 truncate text-[10px] text-muted-foreground">{snippet.description}</div>
        </div>
        <span className={cn('shrink-0 rounded px-1.5 py-0.5 text-[9px] font-medium ring-1 ring-inset', CATEGORY_TONE[snippet.category])}>
          {snippet.category}
        </span>
      </div>
      <div className="mt-2 flex items-center gap-1.5">
        {snippet.os.map((o) => {
          const Icon = OS_ICON[o]
          return (
            <Icon key={o} size={10} className={cn(OS_TONE[o])} />
          )
        })}
      </div>
      <pre className="mt-2 max-h-16 overflow-hidden rounded bg-muted/60 px-2 py-1.5 font-mono text-[10px] leading-relaxed text-foreground/90">
        {snippet.command}
      </pre>
      <div className="mt-2 flex items-center gap-1">
        <button
          onClick={onInsert}
          className="flex flex-1 items-center justify-center gap-1 rounded-md bg-primary/15 px-2 py-1 text-[11px] font-medium text-primary hover:bg-primary/25"
        >
          <Send size={10} /> Insert
        </button>
        <button className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground" title="Copy">
          <Copy size={10} />
        </button>
      </div>
    </div>
  )
}

function VariablesPanel() {
  return (
    <div className="p-3">
      <p className="mb-2 text-[10px] text-muted-foreground">
        Inserted via <kbd className="rounded border border-border bg-muted px-1 font-mono">Tab</kbd>{' '}
        in the prompt. Resolved against the active session context.
      </p>
      <div className="space-y-1">
        {VARIABLES.map((v) => (
          <button
            key={v.name}
            className="group flex w-full items-start justify-between gap-2 rounded-md border border-border bg-background/40 px-2.5 py-1.5 text-left text-xs hover:bg-muted/60"
          >
            <div className="min-w-0 flex-1">
              <div className="font-mono text-[11px] font-medium text-fuchsia-500">{v.name}</div>
              <div className="truncate text-[10px] text-muted-foreground">{v.description}</div>
            </div>
            <div className="text-right">
              <div className="truncate font-mono text-[10px]">{v.resolved}</div>
              <Code2 size={11} className="ml-auto mt-0.5 text-muted-foreground opacity-0 group-hover:opacity-100" />
            </div>
          </button>
        ))}
      </div>
    </div>
  )
}

function HistoryPanel() {
  const [q, setQ] = useState('')
  const filtered = HISTORY.filter((h) => (q ? (h.cmd + h.agent).toLowerCase().includes(q.toLowerCase()) : true))
  return (
    <div>
      <div className="border-b border-border p-2">
        <div className="relative">
          <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Filter history…" className="h-8 pl-7 text-[11px]" />
        </div>
      </div>
      <div className="space-y-0.5 p-2">
        {filtered.map((h, i) => (
          <button
            key={i}
            className="flex w-full flex-col gap-0.5 rounded-md border border-border bg-background/40 px-2.5 py-1.5 text-left text-xs hover:bg-muted/60"
          >
            <div className="flex items-center justify-between">
              <span className="truncate font-mono text-[11px]">{h.agent}</span>
              <span className="font-mono text-[10px] text-muted-foreground">{relativeTime(h.ts)}</span>
            </div>
            <span className="truncate font-mono text-[10px] text-foreground/90">{h.cmd}</span>
          </button>
        ))}
        {filtered.length === 0 && (
          <div className="px-3 py-6 text-center text-[11px] italic text-muted-foreground">
            No history matches.
          </div>
        )}
      </div>
    </div>
  )
}

function FilesPanel() {
  return (
    <div className="p-3">
      <p className="mb-3 text-[11px] text-muted-foreground">
        Drop files here to upload to the active agent, or pull files from the agent for forensic
        review.
      </p>
      <div className="rounded-lg border border-dashed border-border bg-background/40 px-4 py-8 text-center">
        <FileText size={20} className="mx-auto mb-2 text-muted-foreground" strokeWidth={1.5} />
        <div className="text-[11px] text-muted-foreground">Drag & drop or browse to upload</div>
      </div>
      <div className="mt-3 text-[10px] uppercase tracking-wider text-muted-foreground">
        Recent transfers
      </div>
      <ul className="mt-1 space-y-1 text-xs">
        <li className="flex items-center justify-between rounded-md border border-border bg-background/40 px-2.5 py-1.5">
          <span className="inline-flex items-center gap-1.5">
            <Circle size={8} className="fill-emerald-500 text-emerald-500" />
            <span className="font-mono text-[11px]">forensic-rip-1142.zip</span>
          </span>
          <span className="font-mono text-[10px] text-muted-foreground">28.4 MB</span>
        </li>
        <li className="flex items-center justify-between rounded-md border border-border bg-background/40 px-2.5 py-1.5">
          <span className="inline-flex items-center gap-1.5">
            <Circle size={8} className="fill-emerald-500 text-emerald-500" />
            <span className="font-mono text-[11px]">klist-output.txt</span>
          </span>
          <span className="font-mono text-[10px] text-muted-foreground">4.1 KB</span>
        </li>
      </ul>
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

function formatTs(iso: string) {
  return new Date(iso).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}
