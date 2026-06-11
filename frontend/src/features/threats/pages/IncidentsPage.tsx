import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Check,
  ChevronDown,
  Flame,
  LayoutGrid,
  Link2,
  Loader2,
  RefreshCw,
  Rows3,
  Search,
  Send,
  Trash2,
  UserCircle2,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { useDateFormat } from '@/shared/lib/datetime'
import { usersHttpService } from '@/features/team/services/team-http.service'
import {
  incidentsHttpService as svc,
  IncidentsHttpError,
  type Incident,
  type IncidentAlert,
  type IncidentHistory,
  type IncidentNote,
  type IncidentStatus,
} from '../services/incidents-http.service'

/* ─── Status / severity meta ───────────────────────────────────────────── */

const STATUSES: IncidentStatus[] = ['OPEN', 'IN_REVIEW', 'COMPLETED', 'MERGED']

const ST_META: Record<IncidentStatus, { dot: string; pill: string; col: string }> = {
  OPEN: { dot: 'bg-red-500', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', col: 'border-t-red-500' },
  IN_REVIEW: { dot: 'bg-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300', col: 'border-t-sky-500' },
  COMPLETED: { dot: 'bg-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', col: 'border-t-emerald-500' },
  MERGED: { dot: 'bg-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', col: 'border-t-violet-500' },
}

function sevKey(n?: number): 'high' | 'medium' | 'low' | 'unknown' {
  if (n === 3) return 'high'
  if (n === 2) return 'medium'
  if (n === 1) return 'low'
  return 'unknown'
}
const SEV_TONE: Record<string, string> = {
  high: 'text-red-500',
  medium: 'text-amber-500',
  low: 'text-sky-500',
  unknown: 'text-muted-foreground',
}

const SELECT_CLS = 'h-9 rounded-md border border-border bg-background px-2 text-sm'
const COLS = '1fr 120px 90px 150px 70px 110px'

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function IncidentsPage() {
  const { t } = useTranslation()
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [statusFilter, setStatusFilter] = useState<IncidentStatus | 'all'>('all')
  const [assignee, setAssignee] = useState<string>('all')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [layout, setLayout] = useState<'table' | 'board'>('table')

  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [nonce, setNonce] = useState(0)

  const [counts, setCounts] = useState<Record<IncidentStatus, number>>({ OPEN: 0, IN_REVIEW: 0, COMPLETED: 0, MERGED: 0 })
  const [assigneeOptions, setAssigneeOptions] = useState<string[]>([])
  const [open, setOpen] = useState<Incident | null>(null)

  // Debounce the free-text search.
  useEffect(() => {
    const h = setTimeout(() => { setDebounced(search.trim()); setPage(0) }, 300)
    return () => clearTimeout(h)
  }, [search])

  // Board shows a grouped overview (no pager) of a larger window; table paginates.
  const isBoard = layout === 'board'
  const reqSize = isBoard ? 200 : pageSize
  const reqPage = isBoard ? 1 : page + 1

  // Main (server-side) list.
  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    svc
      .list({
        incidentName: debounced || undefined,
        incidentStatus: statusFilter === 'all' ? undefined : statusFilter,
        incidentAssignedTo: assignee === 'all' ? undefined : assignee,
        createdDateStart: dateFrom ? `${dateFrom}T00:00:00Z` : undefined,
        createdDateEnd: dateTo ? `${dateTo}T23:59:59Z` : undefined,
        page: reqPage,
        size: reqSize,
      })
      .then(({ data, total }) => {
        if (cancelled) return
        setIncidents(data ?? [])
        setTotal(total)
      })
      .catch(() => !cancelled && setError(true))
      .finally(() => !cancelled && setLoading(false))
    return () => { cancelled = true }
  }, [debounced, statusFilter, assignee, dateFrom, dateTo, reqPage, reqSize, nonce])

  // Global status counts (independent of filters) + assignee options.
  useEffect(() => {
    let cancelled = false
    Promise.all(STATUSES.map((s) => svc.list({ incidentStatus: s, size: 1 }).then((r) => [s, r.total] as const).catch(() => [s, 0] as const)))
      .then((pairs) => !cancelled && setCounts(Object.fromEntries(pairs) as Record<IncidentStatus, number>))
    svc.usersAssigned().then((u) => !cancelled && setAssigneeOptions(u.map((x) => x.login))).catch(() => {})
    return () => { cancelled = true }
  }, [nonce])

  const refresh = useCallback(() => setNonce((n) => n + 1), [])

  // After a mutation: refetch everything, then resync the open drawer from the fresh page.
  const refreshAndSync = useCallback(
    async (id: number) => {
      refresh()
      try {
        const fresh = await svc.getById(id)
        setOpen((prev) => (prev && prev.id === id ? { ...prev, ...fresh } : prev))
      } catch {
        /* keep current */
      }
    },
    [refresh],
  )

  const setStatus = (s: IncidentStatus | 'all') => { setStatusFilter(s); setPage(0) }

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 py-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-semibold">
            <Flame size={18} strokeWidth={1.75} className="text-red-500" />
            {t('incidents.title')}
            <span className="text-sm font-normal text-muted-foreground">· {total}</span>
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">{t('incidents.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2">
          <LayoutToggle value={layout} onChange={(l) => { setLayout(l); setPage(0) }} t={t} />
          <button
            onClick={refresh}
            title={t('incidents.refresh')}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
          </button>
        </div>
      </header>

      {/* Stat cards (click to filter) */}
      <div className="mt-4 grid grid-cols-2 gap-3 lg:grid-cols-4">
        {STATUSES.map((s) => (
          <button
            key={s}
            onClick={() => setStatus(statusFilter === s ? 'all' : s)}
            className={cn(
              'rounded-xl border bg-card p-4 text-left transition-colors',
              statusFilter === s ? 'border-primary' : 'border-border hover:bg-muted/30',
            )}
          >
            <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
              <span className={cn('h-2 w-2 rounded-full', ST_META[s].dot)} /> {t(`incidents.status.${s}`)}
            </div>
            <div className="mt-1 text-2xl font-semibold tabular-nums">{counts[s]}</div>
          </button>
        ))}
      </div>

      {/* Toolbar */}
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[220px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder={t('incidents.toolbar.search')} value={search} onChange={(e) => setSearch(e.target.value)} className="h-9 pl-9" />
        </div>
        <select value={statusFilter} onChange={(e) => setStatus(e.target.value as IncidentStatus | 'all')} className={SELECT_CLS}>
          <option value="all">{t('incidents.toolbar.allStatuses')}</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>{t(`incidents.status.${s}`)}</option>
          ))}
        </select>
        <select value={assignee} onChange={(e) => { setAssignee(e.target.value); setPage(0) }} className={SELECT_CLS}>
          <option value="all">{t('incidents.toolbar.allAssignees')}</option>
          {assigneeOptions.map((a) => (
            <option key={a} value={a}>{a}</option>
          ))}
        </select>
        <input type="date" value={dateFrom} onChange={(e) => { setDateFrom(e.target.value); setPage(0) }} title={t('incidents.toolbar.from')} className={SELECT_CLS} />
        <input type="date" value={dateTo} onChange={(e) => { setDateTo(e.target.value); setPage(0) }} title={t('incidents.toolbar.to')} className={SELECT_CLS} />
        {(dateFrom || dateTo || debounced || statusFilter !== 'all' || assignee !== 'all') && (
          <button onClick={() => { setSearch(''); setStatusFilter('all'); setAssignee('all'); setDateFrom(''); setDateTo(''); setPage(0) }} className="text-xs text-muted-foreground hover:text-foreground hover:underline">
            {t('incidents.toolbar.clear')}
          </button>
        )}
      </div>

      {/* Content */}
      {error ? (
        <Center>
          <AlertTriangle size={16} className="text-amber-500" /> {t('incidents.loadError')}
          <button onClick={refresh} className="ml-2 text-primary hover:underline">{t('incidents.retry')}</button>
        </Center>
      ) : loading && incidents.length === 0 ? (
        <Center><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></Center>
      ) : incidents.length === 0 ? (
        <Center>{t('incidents.empty')}</Center>
      ) : isBoard ? (
        <Board incidents={incidents} onOpen={setOpen} t={t} />
      ) : (
        <>
          <Table incidents={incidents} onOpen={setOpen} t={t} />
          <div className="mt-3 shrink-0">
            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={(s) => { setPageSize(s); setPage(0) }} />
          </div>
        </>
      )}

      {open && <IncidentDrawer incident={open} onClose={() => setOpen(null)} onChanged={refreshAndSync} t={t} />}
    </div>
  )
}

/* ─── Table ────────────────────────────────────────────────────────────── */

function Table({ incidents, onOpen, t }: { incidents: Incident[]; onOpen: (i: Incident) => void; t: TFunction }) {
  const df = useDateFormat()
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-y-auto rounded-xl border border-border">
      <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
        <div>{t('incidents.table.name')}</div>
        <div>{t('incidents.table.status')}</div>
        <div>{t('incidents.table.severity')}</div>
        <div>{t('incidents.table.assignee')}</div>
        <div className="text-center">{t('incidents.table.alerts')}</div>
        <div>{t('incidents.table.created')}</div>
      </div>
      {incidents.map((i) => (
        <button
          key={i.id}
          onClick={() => onOpen(i)}
          className="grid w-full items-center gap-3 border-b border-border/60 px-4 py-3 text-left text-sm transition-colors last:border-b-0 hover:bg-muted/30"
          style={{ gridTemplateColumns: COLS }}
        >
          <div className="min-w-0">
            <div className="truncate font-medium">{i.incidentName}</div>
            {i.incidentDescription && <div className="truncate text-xs text-muted-foreground">{i.incidentDescription}</div>}
          </div>
          <div><StatusPill status={i.incidentStatus} t={t} /></div>
          <div className={cn('text-xs font-medium', SEV_TONE[sevKey(i.incidentSeverity)])}>{t(`incidents.sev.${sevKey(i.incidentSeverity)}`)}</div>
          <div className="min-w-0"><Assignee login={i.incidentAssignedTo} t={t} /></div>
          <div className="text-center font-mono tabular-nums text-muted-foreground">{i.alertCount}</div>
          <div className="font-mono text-xs text-muted-foreground">{df.formatDate(i.incidentCreatedDate)}</div>
        </button>
      ))}
    </div>
  )
}

/* ─── Board (kanban by status) ─────────────────────────────────────────── */

function Board({ incidents, onOpen, t }: { incidents: Incident[]; onOpen: (i: Incident) => void; t: TFunction }) {
  return (
    <div className="mt-4 grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-y-auto md:grid-cols-2 xl:grid-cols-4">
      {STATUSES.map((s) => {
        const items = incidents.filter((i) => i.incidentStatus === s)
        return (
          <div key={s} className={cn('flex min-h-0 flex-col rounded-xl border border-t-2 border-border bg-card/50', ST_META[s].col)}>
            <div className="flex items-center justify-between px-3 py-2 text-xs font-medium">
              <span className="flex items-center gap-1.5"><span className={cn('h-2 w-2 rounded-full', ST_META[s].dot)} /> {t(`incidents.status.${s}`)}</span>
              <span className="font-mono text-muted-foreground">{items.length}</span>
            </div>
            <div className="min-h-0 flex-1 space-y-2 overflow-y-auto p-2">
              {items.map((i) => (
                <button key={i.id} onClick={() => onOpen(i)} className="w-full rounded-lg border border-border bg-card p-3 text-left transition-colors hover:bg-muted/40">
                  <div className="line-clamp-2 text-sm font-medium">{i.incidentName}</div>
                  <div className="mt-2 flex items-center justify-between text-[11px] text-muted-foreground">
                    <span className={cn('font-medium', SEV_TONE[sevKey(i.incidentSeverity)])}>{t(`incidents.sev.${sevKey(i.incidentSeverity)}`)}</span>
                    <span className="flex items-center gap-1"><Link2 size={11} /> {i.alertCount}</span>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

type Tab = 'overview' | 'alerts' | 'notes' | 'history'

function IncidentDrawer({ incident, onClose, onChanged, t }: { incident: Incident; onClose: () => void; onChanged: (id: number) => void; t: TFunction }) {
  const df = useDateFormat()
  const [tab, setTab] = useState<Tab>('overview')
  const [busy, setBusy] = useState(false)
  const [solution, setSolution] = useState(incident.incidentSolution ?? '')

  const changeStatus = async (status: IncidentStatus) => {
    if (busy || status === incident.incidentStatus) return
    setBusy(true)
    try {
      await svc.changeStatus({
        id: incident.id,
        incidentName: incident.incidentName,
        incidentStatus: status,
        incidentSolution: status === 'COMPLETED' && solution.trim() ? solution.trim() : incident.incidentSolution,
      })
      toast.success(t('incidents.toast.statusChanged'))
      onChanged(incident.id)
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.toast.statusError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                <StatusPill status={incident.incidentStatus} t={t} />
                <span className={cn('font-medium', SEV_TONE[sevKey(incident.incidentSeverity)])}>{t(`incidents.sev.${sevKey(incident.incidentSeverity)}`)}</span>
                <span>· #{incident.id}</span>
              </div>
              <h2 className="mt-1 text-xl font-semibold">{incident.incidentName}</h2>
              <div className="mt-1 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                <Assignee login={incident.incidentAssignedTo} t={t} />
                <span>· {t('incidents.drawer.created', { date: df.formatDateTime(incident.incidentCreatedDate) })}</span>
                <span>· {t('incidents.drawer.alertsCount', { count: incident.alertCount })}</span>
              </div>
            </div>
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
              <X size={16} />
            </button>
          </div>

          {/* Status actions + assignment */}
          <div className="mt-4 flex flex-wrap items-center gap-1.5">
            {STATUSES.map((s) => (
              <button
                key={s}
                disabled={busy || s === incident.incidentStatus}
                onClick={() => void changeStatus(s)}
                className={cn(
                  'inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs transition-colors disabled:cursor-default',
                  s === incident.incidentStatus ? cn('ring-1 ring-inset', ST_META[s].pill) : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground',
                )}
              >
                <span className={cn('h-2 w-2 rounded-full', ST_META[s].dot)} /> {t(`incidents.status.${s}`)}
              </button>
            ))}
            <span className="mx-1 h-4 w-px bg-border" />
            <AssigneePicker incident={incident} onChanged={() => onChanged(incident.id)} t={t} />
          </div>
        </header>

        <nav className="flex items-center gap-1 border-b border-border px-6">
          {(['overview', 'alerts', 'notes', 'history'] as Tab[]).map((id) => (
            <button key={id} onClick={() => setTab(id)} className={cn('relative px-3 py-2.5 text-xs transition-colors', tab === id ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}>
              {t(`incidents.drawer.tabs.${id}`)}
              {tab === id && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          ))}
        </nav>

        <div className="flex-1 overflow-y-auto bg-muted/10 p-6">
          {tab === 'overview' && (
            <div className="space-y-4">
              {incident.incidentDescription && (
                <Section title={t('incidents.drawer.section.description')}>
                  <p className="whitespace-pre-wrap text-xs leading-relaxed text-muted-foreground">{incident.incidentDescription}</p>
                </Section>
              )}
              <Section title={t('incidents.drawer.section.solution')}>
                <textarea
                  value={solution}
                  onChange={(e) => setSolution(e.target.value)}
                  rows={3}
                  placeholder={t('incidents.drawer.solutionPlaceholder')}
                  className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
                <p className="mt-1 text-[11px] text-muted-foreground">{t('incidents.drawer.solutionHint')}</p>
              </Section>
              <Section title={t('incidents.drawer.section.details')}>
                <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                  <Row k={t('incidents.drawer.details.id')}><span className="font-mono">#{incident.id}</span></Row>
                  <Row k={t('incidents.drawer.details.status')}><StatusPill status={incident.incidentStatus} t={t} /></Row>
                  <Row k={t('incidents.drawer.details.severity')}><span className={SEV_TONE[sevKey(incident.incidentSeverity)]}>{t(`incidents.sev.${sevKey(incident.incidentSeverity)}`)}</span></Row>
                  <Row k={t('incidents.drawer.details.assignee')}><Assignee login={incident.incidentAssignedTo} t={t} /></Row>
                  <Row k={t('incidents.drawer.details.created')}>{df.formatDateTime(incident.incidentCreatedDate)}</Row>
                  <Row k={t('incidents.drawer.details.alerts')}>{incident.alertCount}</Row>
                </dl>
              </Section>
            </div>
          )}
          {tab === 'alerts' && <AlertsTab incidentId={incident.id} onChanged={() => onChanged(incident.id)} t={t} />}
          {tab === 'notes' && <NotesTab incidentId={incident.id} t={t} />}
          {tab === 'history' && <HistoryTab incidentId={incident.id} t={t} />}
        </div>
      </div>
    </div>
  )
}

/* ─── Drawer tabs ──────────────────────────────────────────────────────── */

function AlertsTab({ incidentId, onChanged, t }: { incidentId: number; onChanged: () => void; t: TFunction }) {
  const [rows, setRows] = useState<IncidentAlert[] | null>(null)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setError(false)
    try {
      const { data } = await svc.alerts(incidentId)
      setRows(data ?? [])
    } catch {
      setError(true)
    }
  }, [incidentId])
  useEffect(() => { void load() }, [load])

  const remove = async (a: IncidentAlert) => {
    if (!confirm(t('incidents.alerts.removeConfirm', { name: a.alertName }))) return
    try {
      await svc.removeAlert(a.id)
      toast.success(t('incidents.alerts.removed'))
      await load()
      onChanged()
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.alerts.removeError'))
    }
  }

  if (error) return <TabError t={t} onRetry={load} />
  if (rows === null) return <TabLoader />
  if (rows.length === 0) return <TabEmpty>{t('incidents.alerts.empty')}</TabEmpty>
  return (
    <div className="overflow-hidden rounded-lg border border-border bg-card">
      {rows.map((a) => (
        <div key={a.id} className="group flex items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs last:border-b-0">
          <span className={cn('h-4 w-1 shrink-0 rounded-full', a.alertSeverity === 3 ? 'bg-red-500' : a.alertSeverity === 2 ? 'bg-amber-500' : 'bg-sky-500')} />
          <span className="min-w-0 flex-1 truncate font-medium" title={a.alertName}>{a.alertName}</span>
          <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{t(`incidents.sev.${sevKey(a.alertSeverity)}`)}</span>
          <button onClick={() => void remove(a)} title={t('incidents.alerts.remove')} className="shrink-0 rounded p-1 text-muted-foreground opacity-0 transition hover:bg-muted hover:text-red-500 group-hover:opacity-100">
            <Trash2 size={13} />
          </button>
        </div>
      ))}
    </div>
  )
}

function NotesTab({ incidentId, t }: { incidentId: number; t: TFunction }) {
  const df = useDateFormat()
  const [rows, setRows] = useState<IncidentNote[] | null>(null)
  const [error, setError] = useState(false)
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    setError(false)
    try {
      const { data } = await svc.notes(incidentId)
      setRows((data ?? []).slice().sort((a, b) => +new Date(b.noteSendDate) - +new Date(a.noteSendDate)))
    } catch {
      setError(true)
    }
  }, [incidentId])
  useEffect(() => { void load() }, [load])

  const add = async () => {
    const v = text.trim()
    if (!v || busy) return
    setBusy(true)
    try {
      await svc.addNote(incidentId, v)
      setText('')
      await load()
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.notes.addError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-3">
      <div className="rounded-lg border border-border bg-card p-3">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          rows={2}
          maxLength={1000}
          placeholder={t('incidents.notes.placeholder')}
          className="w-full resize-none rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
        <div className="mt-2 flex justify-end">
          <Button size="sm" onClick={() => void add()} disabled={!text.trim() || busy}>
            <Send size={13} className="mr-1.5" /> {t('incidents.notes.add')}
          </Button>
        </div>
      </div>
      {error ? (
        <TabError t={t} onRetry={load} />
      ) : rows === null ? (
        <TabLoader />
      ) : rows.length === 0 ? (
        <TabEmpty>{t('incidents.notes.empty')}</TabEmpty>
      ) : (
        <ul className="space-y-2">
          {rows.map((n) => (
            <li key={n.id} className="rounded-lg border border-border bg-card p-3">
              <p className="whitespace-pre-wrap text-xs leading-relaxed">{n.noteText}</p>
              <div className="mt-1.5 flex items-center gap-2 text-[10px] text-muted-foreground">
                <UserCircle2 size={11} /> {n.noteSendBy || t('incidents.unknownUser')} · {df.formatDateTime(n.noteSendDate)}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function HistoryTab({ incidentId, t }: { incidentId: number; t: TFunction }) {
  const df = useDateFormat()
  const [rows, setRows] = useState<IncidentHistory[] | null>(null)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setError(false)
    try {
      const { data } = await svc.history(incidentId)
      setRows((data ?? []).slice().sort((a, b) => +new Date(b.actionCreatedDate) - +new Date(a.actionCreatedDate)))
    } catch {
      setError(true)
    }
  }, [incidentId])
  useEffect(() => { void load() }, [load])

  if (error) return <TabError t={t} onRetry={load} />
  if (rows === null) return <TabLoader />
  if (rows.length === 0) return <TabEmpty>{t('incidents.history.empty')}</TabEmpty>
  return (
    <ol className="relative ml-2 space-y-4 border-l border-border pl-5">
      {rows.map((h) => (
        <li key={h.id} className="relative">
          <span className="absolute -left-[27px] top-0.5 flex h-4 w-4 items-center justify-center rounded-full border border-border bg-card">
            <span className="h-1.5 w-1.5 rounded-full bg-primary" />
          </span>
          <div className="text-xs font-medium">{h.action}</div>
          {h.actionDetail && <div className="mt-0.5 text-xs text-muted-foreground">{h.actionDetail}</div>}
          <div className="mt-0.5 text-[10px] text-muted-foreground">
            {h.actionCreatedBy || t('incidents.unknownUser')} · {df.formatDateTime(h.actionCreatedDate)}
          </div>
        </li>
      ))}
    </ol>
  )
}

/* ─── Small bits ───────────────────────────────────────────────────────── */

function StatusPill({ status, t }: { status: IncidentStatus; t: TFunction }) {
  return (
    <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium ring-1 ring-inset', ST_META[status].pill)}>
      <span className={cn('h-1.5 w-1.5 rounded-full', ST_META[status].dot)} /> {t(`incidents.status.${status}`)}
    </span>
  )
}

function AssigneePicker({ incident, onChanged, t }: { incident: Incident; onChanged: () => void; t: TFunction }) {
  const [open, setOpen] = useState(false)
  const [users, setUsers] = useState<string[] | null>(null)
  const [busy, setBusy] = useState(false)
  const [q, setQ] = useState('')
  const ref = useRef<HTMLDivElement>(null)
  const current = incident.incidentAssignedTo

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Lazy-load the IAM users on first open (needs users.read).
  useEffect(() => {
    if (!open || users !== null) return
    usersHttpService
      .list({ page_size: 500 })
      .then((r) => setUsers((r.data ?? []).map((u) => u.login).sort()))
      .catch(() => setUsers([]))
  }, [open, users])

  const assign = async (login: string | null) => {
    if (busy) return
    setBusy(true)
    try {
      await svc.assign(incident.id, login)
      toast.success(login ? t('incidents.assign.assigned', { user: login }) : t('incidents.assign.unassigned'))
      setOpen(false)
      onChanged()
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.assign.error'))
    } finally {
      setBusy(false)
    }
  }

  const filtered = (users ?? []).filter((u) => (q ? u.toLowerCase().includes(q.toLowerCase()) : true))

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        disabled={busy}
        className="inline-flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <UserCircle2 size={13} /> <span className="max-w-[120px] truncate">{current || t('incidents.assign.assign')}</span> <ChevronDown size={12} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-60 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              autoFocus
              placeholder={t('incidents.assign.search')}
              className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          {current && (
            <button onClick={() => void assign(null)} className="flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-xs text-muted-foreground hover:bg-muted">
              <X size={12} /> {t('incidents.assign.unassign')}
            </button>
          )}
          <div className="max-h-56 overflow-y-auto">
            {users === null && <div className="flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground"><Loader2 className="h-3.5 w-3.5 animate-spin" /></div>}
            {users !== null && filtered.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('incidents.assign.noUsers')}</div>}
            {filtered.map((u) => (
              <button key={u} onClick={() => void assign(u)} className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted">
                <span className="truncate">{u}</span>
                {u === current && <Check size={14} className="shrink-0 text-primary" />}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function Assignee({ login, t }: { login?: string; t: TFunction }) {
  if (!login) return <span className="inline-flex items-center gap-1 text-muted-foreground/70"><UserCircle2 size={12} /> {t('incidents.unassigned')}</span>
  return <span className="inline-flex min-w-0 items-center gap-1 text-muted-foreground"><UserCircle2 size={12} className="shrink-0" /><span className="truncate">{login}</span></span>
}

function LayoutToggle({ value, onChange, t }: { value: 'table' | 'board'; onChange: (v: 'table' | 'board') => void; t: TFunction }) {
  return (
    <div className="inline-flex rounded-md border border-border p-0.5">
      <button onClick={() => onChange('table')} title={t('incidents.layout.table')} className={cn('flex h-8 w-8 items-center justify-center rounded transition-colors', value === 'table' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>
        <Rows3 size={15} />
      </button>
      <button onClick={() => onChange('board')} title={t('incidents.layout.board')} className={cn('flex h-8 w-8 items-center justify-center rounded transition-colors', value === 'board' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground')}>
        <LayoutGrid size={15} />
      </button>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      {children}
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="min-w-0">{children}</dd>
    </>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="mt-4 flex flex-1 items-center justify-center gap-2 rounded-xl border border-border bg-card text-sm text-muted-foreground">{children}</div>
}

function TabLoader() {
  return <div className="flex h-32 items-center justify-center"><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></div>
}
function TabEmpty({ children }: { children: React.ReactNode }) {
  return <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">{children}</div>
}
function TabError({ t, onRetry }: { t: TFunction; onRetry: () => void }) {
  return (
    <div className="flex h-32 items-center justify-center gap-2 text-sm text-muted-foreground">
      <AlertTriangle size={15} className="text-amber-500" /> {t('incidents.loadError')}
      <button onClick={onRetry} className="text-primary hover:underline">{t('incidents.retry')}</button>
    </div>
  )
}
