import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  BadgeCheck,
  CalendarClock,
  FileBarChart,
  ListChecks,
  Loader2,
  Lock,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import type { Framework, Report } from '../types/compliance.types'
import { scoreTone } from '../components/ReportView'
import { ReportDocument } from '../components/ReportDocument'
import { ReportsTab } from '../components/ReportsTab'
import { ScheduleTab } from '../components/ScheduleTab'
import { ControlsTab } from '../components/ControlsTab'
import { FrameworkEditor } from '../components/FrameworkEditor'

type PageTab = 'frameworks' | 'controls' | 'reports' | 'schedule'

export function CompliancePage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<PageTab>('frameworks')

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 pb-6 pt-3">
      <header className="shrink-0">
        <div className="inline-flex rounded-md border border-border p-0.5">
          <TabButton active={tab === 'frameworks'} onClick={() => setTab('frameworks')} icon={ShieldCheck} label={t('compliance.tabs.frameworks')} />
          <TabButton active={tab === 'controls'} onClick={() => setTab('controls')} icon={ListChecks} label={t('compliance.tabs.controls')} />
          <TabButton active={tab === 'reports'} onClick={() => setTab('reports')} icon={FileBarChart} label={t('compliance.tabs.reports')} />
          <TabButton active={tab === 'schedule'} onClick={() => setTab('schedule')} icon={CalendarClock} label={t('compliance.tabs.schedule')} />
        </div>
      </header>

      <div className="mt-4 flex min-h-0 flex-1 flex-col">
        {tab === 'frameworks' ? <FrameworksTab /> : tab === 'controls' ? <ControlsTab /> : tab === 'reports' ? <ReportsTab /> : <ScheduleTab />}
      </div>
    </div>
  )
}

function TabButton({ active, onClick, icon: Icon, label }: { active: boolean; onClick: () => void; icon: typeof BadgeCheck; label: string }) {
  return (
    <button
      onClick={onClick}
      className={cn('inline-flex items-center gap-1.5 rounded px-3 py-1.5 text-xs transition-colors', active ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
    >
      <Icon size={13} /> {label}
    </button>
  )
}

/* ─── Frameworks tab ────────────────────────────────────────────────────── */

function controlCount(f: Framework): number {
  const ids = new Set<string>()
  for (const s of f.sections ?? []) for (const r of s.requirements ?? []) for (const c of r.satisfiedBy ?? []) ids.add(c)
  return ids.size
}

function FrameworksTab() {
  const { t } = useTranslation()
  const [frameworks, setFrameworks] = useState<Framework[]>([])
  const [scores, setScores] = useState<Record<string, number>>({})
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [open, setOpen] = useState<Framework | null>(null)
  const [editing, setEditing] = useState<{ framework?: Framework; creating: boolean } | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    complianceService
      .listFrameworks()
      .then((fws) => setFrameworks(fws ?? []))
      .catch(() => setError(true))
      .finally(() => setLoading(false))
    // Latest snapshot score per framework (one call, newest-first → dedupe).
    complianceService
      .listSnapshots(undefined, 200)
      .then((snaps) => {
        const m: Record<string, number> = {}
        for (const s of snaps ?? []) if (!(s.frameworkKey in m)) m[s.frameworkKey] = s.score
        setScores(m)
      })
      .catch(() => setScores({}))
  }, [])
  useEffect(() => {
    load()
  }, [load])

  const toggle = async (f: Framework) => {
    if (f.locked) {
      toast.error(t('compliance.locked.upsell'))
      return
    }
    const next = !f.enabled
    setFrameworks((list) => list.map((x) => (x.key === f.key ? { ...x, enabled: next } : x)))
    try {
      await complianceService.setFrameworkEnabled(f.key, next)
    } catch {
      setFrameworks((list) => list.map((x) => (x.key === f.key ? { ...x, enabled: f.enabled } : x)))
      toast.error(t('compliance.toast.toggleError'))
    }
  }

  const shown = useMemo(() => {
    const q = search.trim().toLowerCase()
    return q ? frameworks.filter((f) => (f.name + ' ' + f.key).toLowerCase().includes(q)) : frameworks
  }, [frameworks, search])

  return (
    <>
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('compliance.search')} className="w-[260px] pl-8" />
        </div>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('compliance.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
        <div className="ml-auto">
          <Button size="sm" onClick={() => setEditing({ creating: true })}>
            <Plus size={14} className="mr-1.5" /> {t('compliance.frameworks.new')}
          </Button>
        </div>
      </div>

      <div className="mt-3 min-h-0 flex-1 overflow-y-auto">
        {loading && frameworks.length === 0 ? (
          <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.loading')}</Center>
        ) : error ? (
          <Center>
            <AlertTriangle size={16} className="text-amber-500" /> {t('compliance.loadError')}
            <Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button>
          </Center>
        ) : shown.length === 0 ? (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('compliance.empty')}</div>
        ) : (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {shown.map((f) => (
              <FrameworkCard key={f.key} f={f} score={scores[f.key]} onOpen={() => setOpen(f)} onEdit={() => setEditing({ framework: f, creating: false })} onToggle={() => toggle(f)} t={t} />
            ))}
          </div>
        )}
      </div>

      {open && <FrameworkReportDrawer framework={open} onClose={() => setOpen(null)} t={t} />}
      {editing && (
        <FrameworkEditor
          framework={editing.framework}
          creating={editing.creating}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
        />
      )}
    </>
  )
}

function FrameworkCard({ f, score, onOpen, onEdit, onToggle, t }: { f: Framework; score?: number; onOpen: () => void; onEdit: () => void; onToggle: () => void; t: ReturnType<typeof useTranslation>['t'] }) {
  return (
    <div className={cn('flex cursor-pointer flex-col rounded-xl border border-border bg-card p-4 transition-colors hover:border-primary/40', (!f.enabled || f.locked) && 'opacity-60')} onClick={f.locked ? onToggle : onOpen}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-1.5">
            <h3 className="truncate text-sm font-semibold">{f.name}</h3>
            {f.system && !f.locked && <Lock size={11} className="shrink-0 text-muted-foreground/60" />}
          </div>
          <p className="mt-0.5 line-clamp-2 text-[11px] text-muted-foreground">{f.description || f.key}</p>
        </div>
        <div className="flex shrink-0 items-start gap-2">
          {!f.locked && (
            <button onClick={(e) => { e.stopPropagation(); onEdit() }} title={t('compliance.frameworks.edit')} className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground">
              <Pencil size={13} />
            </button>
          )}
          <div className="text-right">
            {f.locked ? (
              <span className="inline-flex items-center gap-1 rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 dark:text-amber-400">
                <Lock size={10} /> {t('compliance.locked.badge')}
              </span>
            ) : score != null ? (
              <div className={cn('text-2xl font-bold tabular-nums', scoreTone(score))}>{score}%</div>
            ) : (
              <div className="text-[11px] text-muted-foreground">{t('compliance.notEvaluated')}</div>
            )}
          </div>
        </div>
      </div>
      <div className="mt-3 flex items-center justify-between" onClick={(e) => e.stopPropagation()}>
        <span className="text-[11px] text-muted-foreground">{t('compliance.controlsCount', { n: controlCount(f) })}</span>
        {f.locked ? (
          <span className="text-[11px] font-medium text-amber-600 dark:text-amber-400">{t('compliance.locked.upgrade')}</span>
        ) : (
          <label className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
            {f.enabled ? t('compliance.enabled') : t('compliance.disabled')}
            <Toggle checked={f.enabled} onChange={onToggle} />
          </label>
        )}
      </div>
    </div>
  )
}

/* ─── Framework report drawer (live evaluation) ─────────────────────────── */

function FrameworkReportDrawer({ framework, onClose, t }: { framework: Framework; onClose: () => void; t: ReturnType<typeof useTranslation>['t'] }) {
  const [report, setReport] = useState<Report | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [running, setRunning] = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    complianceService
      .liveReport(framework.key)
      .then(setReport)
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [framework.key])
  useEffect(() => {
    load()
  }, [load])

  const runReport = async () => {
    setRunning(true)
    try {
      const r = await complianceService.generateReport(framework.key)
      setReport(r)
      toast.success(t('compliance.toast.reportGenerated'))
    } catch (e) {
      toast.error(e instanceof ComplianceHttpError ? e.message : t('compliance.toast.reportError'))
    } finally {
      setRunning(false)
    }
  }

  if (loading) {
    return (
      <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={onClose}>
        <div className="flex items-center gap-2 text-sm text-white"><Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.evaluating')}</div>
      </div>
    )
  }
  if (error || !report) {
    return (
      <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={onClose}>
        <div className="flex items-center gap-2 rounded-xl bg-card px-6 py-5 text-sm" onClick={(e) => e.stopPropagation()}>
          <AlertTriangle size={16} className="text-amber-500" /> {t('compliance.reportError')}
          <Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button>
          <Button variant="ghost" size="sm" onClick={onClose}>{t('compliance.frameworks.cancel')}</Button>
        </div>
      </div>
    )
  }
  return (
    <ReportDocument
      report={report}
      onClose={onClose}
      onRun={() => void runReport()}
      running={running}
      onDownloadPdf={() => complianceService.downloadFrameworkPdf(framework.key)}
    />
  )
}

function Toggle({ checked, onChange }: { checked: boolean; onChange: () => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      className={cn('relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors', checked ? 'bg-primary' : 'bg-muted-foreground/30')}
    >
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform', checked ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
