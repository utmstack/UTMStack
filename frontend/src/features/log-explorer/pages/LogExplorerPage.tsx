import { useMemo, useState } from 'react'
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Clock,
  Code2,
  Columns3,
  Copy,
  Database,
  Download,
  Filter,
  Minus,
  Play,
  Plus,
  RefreshCw,
  Save,
  Search,
  Sparkles,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type LogLevel = 'critical' | 'error' | 'warn' | 'info' | 'debug'

interface LogEntry {
  id: string
  timestamp: string
  level: LogLevel
  source: string
  host: string
  user?: string
  ip?: string
  message: string
  raw: Record<string, unknown>
}

interface IndexPattern {
  id: string
  name: string
  count: number
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-01T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()

const INDEX_PATTERNS: IndexPattern[] = [
  { id: 'all', name: 'All logs *', count: 24_835_192 },
  { id: 'auth', name: 'auth-*', count: 4_120_044 },
  { id: 'sysmon', name: 'sysmon-*', count: 8_445_812 },
  { id: 'firewall', name: 'firewall-*', count: 6_201_318 },
  { id: 'web', name: 'web-*', count: 3_881_001 },
  { id: 'cloud', name: 'cloud-*', count: 2_187_017 },
]

const LOGS: LogEntry[] = [
  {
    id: 'l-001',
    timestamp: min(2),
    level: 'critical',
    source: 'sysmon',
    host: 'WIN-7F3K2L9Q8X',
    user: 'svc_backup',
    ip: '10.4.18.22',
    message: 'TGS-REQ ticket request for SPN MSSQLSvc/sql-prod-01 from svc_backup (RC4-HMAC)',
    raw: {
      '@timestamp': min(2),
      'event.code': 4769,
      'event.action': 'Kerberos service ticket request',
      'host.name': 'WIN-7F3K2L9Q8X',
      'source.user.name': 'svc_backup',
      'source.ip': '10.4.18.22',
      'kerberos.ticket.encryption_type': 'RC4-HMAC',
      'kerberos.ticket.spn': 'MSSQLSvc/sql-prod-01',
      'log.level': 'critical',
    },
  },
  {
    id: 'l-002',
    timestamp: min(3),
    level: 'critical',
    source: 'sysmon',
    host: 'WIN-7F3K2L9Q8X',
    user: 'svc_backup',
    ip: '10.4.18.22',
    message: 'TGS-REQ ticket request for SPN HOST/dc-01 from svc_backup (RC4-HMAC)',
    raw: { '@timestamp': min(3), 'event.code': 4769, 'host.name': 'WIN-7F3K2L9Q8X', 'kerberos.ticket.spn': 'HOST/dc-01' },
  },
  {
    id: 'l-003',
    timestamp: min(4),
    level: 'warn',
    source: 'sysmon',
    host: 'WIN-7F3K2L9Q8X',
    user: 'svc_backup',
    message: 'Process create: powershell.exe -ExecutionPolicy Bypass -enc <base64>',
    raw: { '@timestamp': min(4), 'event.code': 1, 'process.command_line': 'powershell.exe -ExecutionPolicy Bypass -enc JABz...' },
  },
  {
    id: 'l-004',
    timestamp: min(5),
    level: 'info',
    source: 'firewall',
    host: 'edge-fw-bsa-01',
    ip: '10.4.18.22',
    message: 'Outbound connection: 10.4.18.22:51442 → 198.51.100.14:443 (allowed)',
    raw: { '@timestamp': min(5), 'source.ip': '10.4.18.22', 'destination.ip': '198.51.100.14', 'destination.port': 443, 'network.transport': 'tcp' },
  },
  {
    id: 'l-005',
    timestamp: min(7),
    level: 'error',
    source: 'auth',
    host: 'mail-srv-02',
    user: 'mthomas',
    ip: '203.0.113.7',
    message: 'Failed login attempt for mthomas from 203.0.113.7 (invalid password)',
    raw: { '@timestamp': min(7), 'event.outcome': 'failure', 'user.name': 'mthomas', 'source.ip': '203.0.113.7' },
  },
  {
    id: 'l-006',
    timestamp: min(7),
    level: 'error',
    source: 'auth',
    host: 'mail-srv-02',
    user: 'mthomas',
    ip: '203.0.113.7',
    message: 'Failed login attempt for mthomas from 203.0.113.7 (invalid password)',
    raw: { '@timestamp': min(7), 'event.outcome': 'failure', 'user.name': 'mthomas', 'source.ip': '203.0.113.7' },
  },
  {
    id: 'l-007',
    timestamp: min(8),
    level: 'info',
    source: 'auth',
    host: 'mail-srv-02',
    user: 'mthomas',
    ip: '203.0.113.7',
    message: 'Successful login for mthomas from 203.0.113.7',
    raw: { '@timestamp': min(8), 'event.outcome': 'success', 'user.name': 'mthomas', 'source.ip': '203.0.113.7' },
  },
  {
    id: 'l-008',
    timestamp: min(11),
    level: 'warn',
    source: 'sysmon',
    host: 'fin-app-01',
    user: 'jsmith',
    message: 'Network connection from PowerShell.exe to 198.51.100.14:443',
    raw: { '@timestamp': min(11), 'event.code': 3, 'process.name': 'powershell.exe', 'destination.ip': '198.51.100.14' },
  },
  {
    id: 'l-009',
    timestamp: min(14),
    level: 'info',
    source: 'web',
    host: 'webapp-01',
    ip: '198.51.100.14',
    message: 'GET /api/users HTTP/1.1 200 — User-Agent: curl/7.79',
    raw: { '@timestamp': min(14), 'http.request.method': 'GET', 'url.path': '/api/users', 'http.response.status_code': 200 },
  },
  {
    id: 'l-010',
    timestamp: min(18),
    level: 'debug',
    source: 'sysmon',
    host: 'fin-app-01',
    message: 'Process exit: powershell.exe (PID 4128) exit_code 0',
    raw: { '@timestamp': min(18), 'event.code': 5, 'process.pid': 4128 },
  },
  {
    id: 'l-011',
    timestamp: min(22),
    level: 'error',
    source: 'firewall',
    host: 'edge-fw-bsa-01',
    ip: '203.0.113.7',
    message: 'Blocked: 203.0.113.7 → 10.0.4.91:22 (rate limit exceeded)',
    raw: { '@timestamp': min(22), 'event.outcome': 'denied', 'source.ip': '203.0.113.7' },
  },
  {
    id: 'l-012',
    timestamp: min(26),
    level: 'info',
    source: 'cloud',
    host: 'aws-cloudtrail',
    user: 'aluna@utmstack.com',
    message: 'AWS API call: ConsoleLogin (success) from 203.0.113.91',
    raw: { '@timestamp': min(26), 'cloud.provider': 'aws', 'event.action': 'ConsoleLogin' },
  },
  {
    id: 'l-013',
    timestamp: min(31),
    level: 'critical',
    source: 'sysmon',
    host: 'WIN-7F3K2L9Q8X',
    user: 'svc_backup',
    message: 'AV signature match: HackTool:Win32/Mimikatz!MTB at C:\\Users\\svc_backup\\m.exe',
    raw: { '@timestamp': min(31), 'event.code': 1117, 'threat.indicator.name': 'HackTool:Win32/Mimikatz' },
  },
  {
    id: 'l-014',
    timestamp: min(38),
    level: 'warn',
    source: 'sysmon',
    host: 'WIN-7F3K2L9Q8X',
    user: 'svc_backup',
    message: 'SMB file access: \\\\fileshare-01\\Finance\\Q4-payroll.xlsx (read)',
    raw: { '@timestamp': min(38), 'event.code': 5145, 'file.path': '\\\\fileshare-01\\Finance\\Q4-payroll.xlsx' },
  },
  {
    id: 'l-015',
    timestamp: min(46),
    level: 'critical',
    source: 'firewall',
    host: 'edge-fw-bsa-01',
    ip: '198.51.100.14',
    message: 'IOC match (C2): outbound to 198.51.100.14:443 from fin-app-01',
    raw: { '@timestamp': min(46), 'threat.indicator.type': 'ipv4-addr', 'threat.feed.name': 'AlienVault OTX' },
  },
  {
    id: 'l-016',
    timestamp: min(58),
    level: 'info',
    source: 'sysmon',
    host: 'fin-app-01',
    user: 'SYSTEM',
    message: 'Scheduled task created: \\Microsoft\\Windows\\update_check (runs daily)',
    raw: { '@timestamp': min(58), 'event.code': 4698 },
  },
  {
    id: 'l-017',
    timestamp: min(62),
    level: 'info',
    source: 'web',
    host: 'webapp-01',
    message: 'POST /api/login HTTP/1.1 200 (response time 142ms)',
    raw: { '@timestamp': min(62), 'http.request.method': 'POST', 'url.path': '/api/login' },
  },
  {
    id: 'l-018',
    timestamp: min(74),
    level: 'warn',
    source: 'firewall',
    host: 'edge-fw-bsa-01',
    message: 'Unusual data transfer: fin-app-01 → 45.83.66.12:443 (148 MB in 4 min)',
    raw: { '@timestamp': min(74), 'network.bytes': 154_926_080, 'destination.ip': '45.83.66.12' },
  },
  {
    id: 'l-019',
    timestamp: min(91),
    level: 'debug',
    source: 'sysmon',
    host: 'WIN-7F3K2L9Q8X',
    message: 'Driver loaded: \\Driver\\WdFilter (signed: Microsoft Corporation)',
    raw: { '@timestamp': min(91), 'event.code': 6, 'driver.signed': true },
  },
  {
    id: 'l-020',
    timestamp: min(105),
    level: 'info',
    source: 'auth',
    host: 'dc-01',
    user: 'jsmith',
    message: 'Kerberos pre-authentication successful for jsmith from 10.4.18.91',
    raw: { '@timestamp': min(105), 'event.code': 4768, 'user.name': 'jsmith' },
  },
  {
    id: 'l-021',
    timestamp: min(132),
    level: 'warn',
    source: 'auth',
    host: 'dc-01',
    user: 'tferguson',
    message: 'GPO modification by non-administrator: tferguson modified Default Domain Policy',
    raw: { '@timestamp': min(132), 'event.code': 5136, 'user.name': 'tferguson' },
  },
  {
    id: 'l-022',
    timestamp: min(180),
    level: 'info',
    source: 'cloud',
    host: 'azure-graph',
    user: 'admin@utmstack.com',
    message: 'Azure AD: New conditional access policy created (id: ca-finance-mfa)',
    raw: { '@timestamp': min(180), 'cloud.provider': 'azure' },
  },
]

/* ─── Saved views ──────────────────────────────────────────────────────── */

interface SavedView {
  id: string
  label: string
  count: number
  predicate: (l: LogEntry) => boolean
}

const SAVED_VIEWS: SavedView[] = [
  { id: 'all', label: 'All', count: LOGS.length, predicate: () => true },
  { id: 'errors', label: 'Errors', count: LOGS.filter((l) => l.level === 'error' || l.level === 'critical').length, predicate: (l) => l.level === 'error' || l.level === 'critical' },
  { id: 'auth', label: 'Auth', count: LOGS.filter((l) => l.source === 'auth').length, predicate: (l) => l.source === 'auth' },
  { id: 'network', label: 'Network', count: LOGS.filter((l) => l.source === 'firewall' || l.source === 'web').length, predicate: (l) => l.source === 'firewall' || l.source === 'web' },
  { id: 'sysmon', label: 'Sysmon', count: LOGS.filter((l) => l.source === 'sysmon').length, predicate: (l) => l.source === 'sysmon' },
  { id: 'cloud', label: 'Cloud', count: LOGS.filter((l) => l.source === 'cloud').length, predicate: (l) => l.source === 'cloud' },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const LEVEL_STYLE: Record<LogLevel, { dot: string; label: string; tone: string }> = {
  critical: { dot: 'bg-red-500', label: 'CRITICAL', tone: 'text-red-500' },
  error: { dot: 'bg-orange-500', label: 'ERROR', tone: 'text-orange-500' },
  warn: { dot: 'bg-amber-500', label: 'WARN', tone: 'text-amber-500' },
  info: { dot: 'bg-sky-500', label: 'INFO', tone: 'text-sky-500' },
  debug: { dot: 'bg-muted-foreground', label: 'DEBUG', tone: 'text-muted-foreground' },
}

/* ─── Field metadata ───────────────────────────────────────────────────── */

interface FieldStat {
  name: string
  type: 'date' | 'keyword' | 'ip' | 'number' | 'text'
  topValues: { value: string; count: number; pct: number }[]
}

const FIELDS: FieldStat[] = [
  {
    name: 'level',
    type: 'keyword',
    topValues: [
      { value: 'info', count: 9, pct: 41 },
      { value: 'warn', count: 5, pct: 23 },
      { value: 'critical', count: 4, pct: 18 },
      { value: 'error', count: 3, pct: 14 },
      { value: 'debug', count: 1, pct: 4 },
    ],
  },
  {
    name: 'source',
    type: 'keyword',
    topValues: [
      { value: 'sysmon', count: 8, pct: 36 },
      { value: 'auth', count: 5, pct: 23 },
      { value: 'firewall', count: 4, pct: 18 },
      { value: 'web', count: 2, pct: 9 },
      { value: 'cloud', count: 2, pct: 9 },
    ],
  },
  {
    name: 'host',
    type: 'keyword',
    topValues: [
      { value: 'WIN-7F3K2L9Q8X', count: 7, pct: 32 },
      { value: 'fin-app-01', count: 4, pct: 18 },
      { value: 'edge-fw-bsa-01', count: 4, pct: 18 },
      { value: 'mail-srv-02', count: 3, pct: 14 },
      { value: 'dc-01', count: 2, pct: 9 },
    ],
  },
  {
    name: 'user',
    type: 'keyword',
    topValues: [
      { value: 'svc_backup', count: 6, pct: 27 },
      { value: 'mthomas', count: 3, pct: 14 },
      { value: 'jsmith', count: 2, pct: 9 },
      { value: 'tferguson', count: 1, pct: 5 },
      { value: 'aluna@utmstack.com', count: 1, pct: 5 },
    ],
  },
  {
    name: 'ip',
    type: 'ip',
    topValues: [
      { value: '10.4.18.22', count: 6, pct: 27 },
      { value: '203.0.113.7', count: 4, pct: 18 },
      { value: '198.51.100.14', count: 3, pct: 14 },
      { value: '45.83.66.12', count: 1, pct: 5 },
    ],
  },
  { name: '@timestamp', type: 'date', topValues: [] },
  { name: 'event.code', type: 'number', topValues: [] },
  { name: 'event.action', type: 'keyword', topValues: [] },
  { name: 'message', type: 'text', topValues: [] },
]

const DEFAULT_COLUMNS = ['timestamp', 'level', 'source', 'host', 'message']

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function LogExplorerPage() {
  const [pattern, setPattern] = useState<IndexPattern>(INDEX_PATTERNS[0])
  const [query, setQuery] = useState('')
  const [view, setView] = useState<string>('all')
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [columns] = useState<string[]>(DEFAULT_COLUMNS)
  const [sqlMode, setSqlMode] = useState(false)

  const filtered = useMemo(() => {
    const v = SAVED_VIEWS.find((s) => s.id === view) ?? SAVED_VIEWS[0]
    return LOGS.filter(v.predicate).filter((l) =>
      query
        ? (l.message + l.host + (l.user ?? '') + (l.ip ?? '') + l.source)
            .toLowerCase()
            .includes(query.toLowerCase())
        : true
    )
  }, [view, query])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <Header total={filtered.length} pattern={pattern} />

      <div className="mt-4">
        <QueryBar
          pattern={pattern}
          onPatternChange={setPattern}
          query={query}
          onQueryChange={setQuery}
          sqlMode={sqlMode}
          onSqlMode={setSqlMode}
        />
      </div>

      <div className="mt-4">
        <HistogramCard total={filtered.length} />
      </div>

      <div className="mt-5">
        <SavedViewTabsWithActions current={view} onChange={setView} />
      </div>

      {/* Unified body: sidebar + results inside a single card with vertical divider */}
      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <div className="flex min-h-0">
          <FieldSidebar selected={columns} />
          <div className="min-w-0 flex-1 border-l border-border">
            <ResultsHeader />
            {filtered.map((l) => (
              <ResultRow
                key={l.id}
                log={l}
                expanded={expandedId === l.id}
                onToggle={() => setExpandedId(expandedId === l.id ? null : l.id)}
              />
            ))}
            {filtered.length === 0 && (
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                No logs match this query.
              </div>
            )}
          </div>
        </div>
      </div>

      <Pagination total={filtered.length} />
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ total, pattern }: { total: number; pattern: IndexPattern }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Log explorer</h1>
        <p className="mt-1 text-xs text-muted-foreground">
          {pattern.count.toLocaleString()} events in <span className="font-mono">{pattern.name}</span> ·
          showing <span className="font-medium text-foreground">{total}</span> matching
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <Save size={14} className="mr-2" />
          Save query
        </Button>
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          Export
        </Button>
      </div>
    </header>
  )
}

/* ─── Query bar ────────────────────────────────────────────────────────── */

function QueryBar({
  pattern,
  onPatternChange,
  query,
  onQueryChange,
  sqlMode,
  onSqlMode,
}: {
  pattern: IndexPattern
  onPatternChange: (p: IndexPattern) => void
  query: string
  onQueryChange: (q: string) => void
  sqlMode: boolean
  onSqlMode: (b: boolean) => void
}) {
  const [patternOpen, setPatternOpen] = useState(false)
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-xl border border-border bg-card p-2">
      {/* Index pattern selector */}
      <div className="relative">
        <button
          onClick={() => setPatternOpen((v) => !v)}
          className={cn(
            'flex h-9 items-center gap-2 rounded-md px-3 text-sm transition-colors',
            patternOpen ? 'bg-muted' : 'hover:bg-muted'
          )}
        >
          <Database size={13} className="text-muted-foreground" />
          <span className="font-mono">{pattern.name}</span>
          <ChevronDown size={12} className="text-muted-foreground" />
        </button>
        {patternOpen && (
          <div className="absolute left-0 top-full z-30 mt-1 w-72 overflow-hidden rounded-md border border-border bg-popover shadow-lg">
            <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
              Index patterns
            </div>
            {INDEX_PATTERNS.map((p) => (
              <button
                key={p.id}
                onClick={() => {
                  onPatternChange(p)
                  setPatternOpen(false)
                }}
                className={cn(
                  'flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm transition-colors',
                  p.id === pattern.id ? 'bg-muted/60' : 'hover:bg-muted/60'
                )}
              >
                <span className="font-mono">{p.name}</span>
                <span className="font-mono text-[11px] text-muted-foreground">
                  {p.count.toLocaleString()}
                </span>
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="h-5 w-px bg-border" />

      {/* Search input — KQL or SQL */}
      <div className="relative flex-1 min-w-[300px]">
        {sqlMode ? (
          <Input
            value={query}
            onChange={(e) => onQueryChange(e.target.value)}
            placeholder="SELECT * FROM logs WHERE level = 'critical' ORDER BY timestamp DESC"
            className="h-9 border-0 bg-transparent font-mono text-xs shadow-none focus-visible:ring-0"
          />
        ) : (
          <>
            <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => onQueryChange(e.target.value)}
              placeholder='level: critical AND source.user.name: "svc_*"   ·   KQL'
              className="h-9 border-0 bg-transparent pl-8 font-mono text-xs shadow-none focus-visible:ring-0"
            />
          </>
        )}
      </div>

      <div className="h-5 w-px bg-border" />

      <Button variant="outline" size="sm">
        <Clock size={13} className="mr-2" />
        Last 24 hours
        <ChevronDown size={11} className="ml-1.5 opacity-60" />
      </Button>

      <button
        onClick={() => onSqlMode(!sqlMode)}
        className={cn(
          'flex h-9 items-center gap-1.5 rounded-md px-2.5 text-xs transition-colors',
          sqlMode ? 'bg-violet-500/15 text-violet-600 dark:text-violet-300' : 'text-muted-foreground hover:bg-muted hover:text-foreground'
        )}
        title="Toggle SQL mode"
      >
        <Code2 size={13} />
        SQL
      </button>

      <Button size="sm">
        <Play size={12} className="mr-1.5" />
        Run
      </Button>
    </div>
  )
}

/* ─── Histogram ────────────────────────────────────────────────────────── */

function HistogramCard({ total }: { total: number }) {
  const buckets = 60
  const data = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.max(
          0,
          Math.round(
            Math.abs(Math.sin(i / 6)) * 28 +
              Math.abs(Math.sin(i / 11)) * 12 +
              (i > 50 && i < 56 ? 22 : 0) +
              4
          )
        )
      ),
    []
  )
  const errors = useMemo(
    () =>
      Array.from({ length: buckets }, (_, i) =>
        Math.max(0, Math.round(Math.abs(Math.sin(i / 8)) * 4 + (i > 50 && i < 56 ? 6 : 0)))
      ),
    []
  )
  const max = Math.max(...data.map((v, i) => v + errors[i])) * 1.1
  const w = 1200
  const h = 100
  const bw = w / buckets - 2

  return (
    <div className="rounded-xl border border-border bg-card p-5">
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            Events over time
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-2xl font-semibold tabular-nums">
              {total.toLocaleString()}
            </span>
            <span className="text-sm text-muted-foreground">events match</span>
          </div>
        </div>
        <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
          <span className="inline-flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-sm bg-sky-500/60" />
            All
          </span>
          <span className="inline-flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-sm bg-red-500/80" />
            Critical/Error
          </span>
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        {data.map((v, i) => {
          const e = errors[i]
          const baseH = (v / max) * h
          const errH = (e / max) * h
          const x = i * (w / buckets) + 1
          return (
            <g key={i}>
              <rect x={x} y={h - baseH} width={bw} height={baseH - errH} rx={1.5} className="fill-sky-500/55" />
              <rect x={x} y={h - errH} width={bw} height={errH} rx={1.5} className="fill-red-500/80" />
            </g>
          )
        })}
      </svg>

      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>24h ago</span>
        <span>18h ago</span>
        <span>12h ago</span>
        <span>6h ago</span>
        <span>now</span>
      </div>
    </div>
  )
}

/* ─── Saved view tabs (with actions on the right) ──────────────────────── */

function SavedViewTabsWithActions({
  current,
  onChange,
}: {
  current: string
  onChange: (id: string) => void
}) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border">
      <div className="flex flex-wrap items-center">
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
      <div className="flex items-center gap-1 pr-1">
        <button className="flex h-7 items-center gap-1.5 rounded-md px-2 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground" title="Add filter">
          <Filter size={12} />
          Add filter
        </button>
        <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" title="Columns">
          <Columns3 size={13} />
        </button>
        <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" title="Refresh">
          <RefreshCw size={13} />
        </button>
      </div>
    </div>
  )
}

/* ─── Field sidebar ────────────────────────────────────────────────────── */

function FieldSidebar({ selected }: { selected: string[] }) {
  const [q, setQ] = useState('')
  const [openField, setOpenField] = useState<string | null>('source')
  const filtered = FIELDS.filter((f) =>
    q ? f.name.toLowerCase().includes(q.toLowerCase()) : true
  )
  const selectedFields = filtered.filter((f) => selected.includes(f.name))
  const availableFields = filtered.filter((f) => !selected.includes(f.name))

  return (
    <aside className="hidden w-60 shrink-0 flex-col bg-muted/20 lg:flex">
      <div className="border-b border-border p-2">
        <div className="relative">
          <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Filter fields…"
            className="h-8 border-0 bg-transparent pl-7 text-[11px] shadow-none focus-visible:bg-card focus-visible:ring-1 focus-visible:ring-border"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-1.5">
        <div className="px-2 pt-1 pb-1 text-[10px] uppercase tracking-wider text-muted-foreground/70">
          Selected
        </div>
        {selectedFields.map((f) => (
          <FieldItem
            key={f.name}
            field={f}
            open={openField === f.name}
            onToggle={() => setOpenField(openField === f.name ? null : f.name)}
            selected
          />
        ))}

        <div className="mt-3 px-2 pt-1 pb-1 text-[10px] uppercase tracking-wider text-muted-foreground/70">
          Available
        </div>
        {availableFields.map((f) => (
          <FieldItem
            key={f.name}
            field={f}
            open={openField === f.name}
            onToggle={() => setOpenField(openField === f.name ? null : f.name)}
          />
        ))}
      </div>
    </aside>
  )
}

function FieldItem({
  field,
  open,
  onToggle,
  selected,
}: {
  field: FieldStat
  open: boolean
  onToggle: () => void
  selected?: boolean
}) {
  const typeColor: Record<FieldStat['type'], string> = {
    date: 'bg-violet-500',
    keyword: 'bg-sky-500',
    ip: 'bg-fuchsia-500',
    number: 'bg-amber-500',
    text: 'bg-emerald-500',
  }
  return (
    <div className={cn('rounded-md', open && 'bg-card')}>
      <button
        onClick={onToggle}
        className={cn(
          'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors',
          !open && 'hover:bg-card/60',
          selected && !open && 'bg-card/40'
        )}
      >
        <span className={cn('h-1.5 w-1.5 shrink-0 rounded-full opacity-50', typeColor[field.type])} />
        <span className="flex-1 truncate font-mono text-[11px]">{field.name}</span>
        {selected && (
          <span className="shrink-0 text-[10px] text-primary opacity-70">●</span>
        )}
      </button>
      {open && field.topValues.length > 0 && (
        <div className="space-y-1.5 px-2 pb-2.5 pt-1.5">
          {field.topValues.slice(0, 5).map((v) => (
            <div key={v.value} className="group flex items-center gap-2 text-[11px]">
              <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between gap-1.5">
                  <span className="truncate font-mono text-[10px]">{v.value}</span>
                  <span className="shrink-0 font-mono text-[9px] text-muted-foreground">
                    {v.pct}%
                  </span>
                </div>
                <div className="mt-0.5 h-0.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-foreground/40"
                    style={{ width: `${v.pct}%` }}
                  />
                </div>
              </div>
              <div className="flex shrink-0 gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                <button
                  title="Add to query"
                  className="flex h-4 w-4 items-center justify-center rounded text-emerald-500 hover:bg-emerald-500/15"
                >
                  <Plus size={10} />
                </button>
                <button
                  title="Exclude"
                  className="flex h-4 w-4 items-center justify-center rounded text-red-500 hover:bg-red-500/15"
                >
                  <Minus size={10} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

/* ─── Results table ────────────────────────────────────────────────────── */

const COLS = '20px 4px 110px 80px 110px 160px 1fr'

function ResultsHeader() {
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: COLS }}
    >
      <div />
      <div />
      <div>Time</div>
      <div>Level</div>
      <div>Source</div>
      <div>Host</div>
      <div>Message</div>
    </div>
  )
}

function ResultRow({
  log,
  expanded,
  onToggle,
}: {
  log: LogEntry
  expanded: boolean
  onToggle: () => void
}) {
  const lv = LEVEL_STYLE[log.level]
  return (
    <>
      <div
        onClick={onToggle}
        className={cn(
          'grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2 text-xs transition-colors last:border-b-0',
          expanded ? 'bg-muted/40' : 'hover:bg-muted/30'
        )}
        style={{ gridTemplateColumns: COLS }}
      >
        <ChevronRight
          size={12}
          className={cn(
            'text-muted-foreground transition-transform',
            expanded && 'rotate-90 text-foreground'
          )}
        />
        <span className={cn('h-3 w-1 rounded-full', lv.dot)} />
        <div className="font-mono tabular-nums text-muted-foreground">
          {shortTime(log.timestamp)}
        </div>
        <div className={cn('font-mono text-[10px] font-semibold tracking-wider', lv.tone)}>
          {lv.label}
        </div>
        <div className="font-mono text-muted-foreground">{log.source}</div>
        <div className="truncate font-mono text-foreground/85">{log.host}</div>
        <div className="truncate text-foreground/90">{log.message}</div>
      </div>
      {expanded && <ExpandedPanel log={log} />}
    </>
  )
}

/* ─── Expanded inline panel ────────────────────────────────────────────── */

type DetailTab = 'fields' | 'json' | 'related'

function ExpandedPanel({ log }: { log: LogEntry }) {
  const [tab, setTab] = useState<DetailTab>('fields')
  const lv = LEVEL_STYLE[log.level]
  return (
    <div
      className={cn(
        'border-b border-border bg-muted/40 last:border-b-0',
        'border-l-2',
        log.level === 'critical' && 'border-l-red-500',
        log.level === 'error' && 'border-l-orange-500',
        log.level === 'warn' && 'border-l-amber-500',
        log.level === 'info' && 'border-l-sky-500',
        log.level === 'debug' && 'border-l-muted-foreground/50'
      )}
    >
      {/* Sub header */}
      <div className="flex items-start justify-between gap-4 border-b border-border/60 px-5 py-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
            <span className="font-mono">{log.id}</span>
            <span>·</span>
            <span className="font-mono">{absTimestamp(log.timestamp)}</span>
            <span>·</span>
            <span className={cn('font-mono font-semibold', lv.tone)}>{lv.label}</span>
            <span>·</span>
            <span className="font-mono">{log.source}</span>
            <span>·</span>
            <span className="font-mono">{log.host}</span>
            {log.user && (
              <>
                <span>·</span>
                <span className="font-mono">{log.user}</span>
              </>
            )}
          </div>
          <div className="mt-1.5 font-mono text-xs leading-snug text-foreground">
            {log.message}
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <button className="flex h-7 items-center gap-1.5 rounded-md px-2 text-[11px] text-fuchsia-500 hover:bg-fuchsia-500/10">
            <Sparkles size={12} />
            Explain
          </button>
          <button className="flex h-7 items-center gap-1.5 rounded-md px-2 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground">
            <Filter size={12} />
            Pivot
          </button>
          <button className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" title="Copy">
            <Copy size={12} />
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-4 border-b border-border/60 px-5">
        <DetailTabBtn id="fields" current={tab} onChange={setTab}>
          Parsed fields
          <span className="ml-1.5 font-mono text-[10px] text-muted-foreground">
            {Object.keys(log.raw).length}
          </span>
        </DetailTabBtn>
        <DetailTabBtn id="json" current={tab} onChange={setTab}>JSON</DetailTabBtn>
        <DetailTabBtn id="related" current={tab} onChange={setTab}>Related events</DetailTabBtn>
      </div>

      {/* Tab content */}
      <div className="p-5">
        {tab === 'fields' && <ExpandedFields log={log} />}
        {tab === 'json' && <ExpandedJson log={log} />}
        {tab === 'related' && <ExpandedRelated />}
      </div>
    </div>
  )
}

function DetailTabBtn({
  id,
  current,
  onChange,
  children,
}: {
  id: DetailTab
  current: DetailTab
  onChange: (t: DetailTab) => void
  children: React.ReactNode
}) {
  const active = id === current
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative -mb-px border-b-2 py-2 text-xs transition-colors',
        active
          ? 'border-foreground text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'
      )}
    >
      {children}
    </button>
  )
}

function ExpandedFields({ log }: { log: LogEntry }) {
  const entries = Object.entries(log.raw)
  return (
    <div className="overflow-hidden rounded-md border border-border bg-card">
      {entries.map(([k, v], i) => (
        <div
          key={k}
          className={cn(
            'group grid grid-cols-[200px_1fr_70px] items-center gap-3 px-3 py-1.5 text-xs',
            i < entries.length - 1 && 'border-b border-border/60',
            'hover:bg-muted/30'
          )}
        >
          <div className="truncate font-mono text-[11px] text-muted-foreground">{k}</div>
          <div className="truncate font-mono text-[11px]">{String(v)}</div>
          <div className="flex justify-end gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
            <button
              title="Add to query"
              className="flex h-5 w-5 items-center justify-center rounded text-emerald-500 hover:bg-emerald-500/15"
            >
              <Plus size={10} />
            </button>
            <button
              title="Exclude"
              className="flex h-5 w-5 items-center justify-center rounded text-red-500 hover:bg-red-500/15"
            >
              <Minus size={10} />
            </button>
            <button
              title="Copy"
              className="flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <Copy size={10} />
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}

function ExpandedJson({ log }: { log: LogEntry }) {
  return (
    <pre className="overflow-x-auto rounded-md border border-border bg-card p-3 font-mono text-[11px] leading-relaxed">
      {JSON.stringify(log.raw, null, 2)}
    </pre>
  )
}

function ExpandedRelated() {
  const related = LOGS.slice(0, 4)
  return (
    <div className="overflow-hidden rounded-md border border-border bg-card">
      {related.map((l, i) => {
        const lv = LEVEL_STYLE[l.level]
        return (
          <div
            key={l.id}
            className={cn(
              'grid grid-cols-[100px_60px_140px_1fr] items-baseline gap-3 px-3 py-1.5 text-xs hover:bg-muted/30',
              i < related.length - 1 && 'border-b border-border/60'
            )}
          >
            <div className="font-mono text-[10px] text-muted-foreground">
              {shortTime(l.timestamp)}
            </div>
            <div className={cn('font-mono text-[9px] font-semibold tracking-wider', lv.tone)}>
              {lv.label}
            </div>
            <div className="truncate font-mono text-[10px] text-muted-foreground">{l.host}</div>
            <div className="truncate font-mono text-[11px]">{l.message}</div>
          </div>
        )
      })}
    </div>
  )
}

/* ─── Pagination ───────────────────────────────────────────────────────── */

function Pagination({ total }: { total: number }) {
  return (
    <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
      <span>
        Showing <span className="text-foreground">1–{Math.min(50, total)}</span> of{' '}
        <span className="text-foreground">{total.toLocaleString()}</span>
      </span>
      <div className="flex items-center gap-1">
        <Button variant="outline" size="sm" disabled>
          <ChevronLeft size={14} />
        </Button>
        <Button variant="outline" size="sm" disabled>
          <ChevronRight size={14} />
        </Button>
      </div>
    </div>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function shortTime(iso: string) {
  return new Date(iso).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function absTimestamp(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}
