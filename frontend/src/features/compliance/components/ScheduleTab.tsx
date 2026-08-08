import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, CalendarClock, Loader2, Mail, Pencil, Plus, RefreshCw, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import type { ComplianceSchedule, Framework, SaveScheduleInput } from '../types/compliance.types'

const CRON_PRESETS: { key: string; cron: string }[] = [
  { key: 'daily', cron: '0 6 * * *' },
  { key: 'weekly', cron: '0 6 * * 1' },
  { key: 'monthly', cron: '0 6 1 * *' },
]
const SELECT = 'h-9 rounded-md border border-input bg-background px-2.5 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

function cadenceLabel(cron: string, t: ReturnType<typeof useTranslation>['t']): string {
  const p = CRON_PRESETS.find((x) => x.cron === cron.trim())
  return p ? t(`compliance.schedule.cadence.${p.key}`) : t('compliance.schedule.cadence.custom')
}

export function ScheduleTab() {
  const { t } = useTranslation()
  const [schedules, setSchedules] = useState<ComplianceSchedule[]>([])
  const [frameworks, setFrameworks] = useState<Framework[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<ComplianceSchedule | 'new' | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    Promise.all([complianceService.listSchedules(), complianceService.listFrameworks()])
      .then(([sched, fws]) => {
        setSchedules(sched.data ?? [])
        setFrameworks(fws ?? [])
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [])
  useEffect(() => {
    load()
  }, [load])

  const fwName = useMemo(() => {
    const m: Record<string, string> = {}
    for (const f of frameworks) m[f.key] = f.name
    return m
  }, [frameworks])

  // Only entitled frameworks can be scheduled (the backend rejects locked ones).
  const selectable = useMemo(() => frameworks.filter((f) => !f.locked), [frameworks])

  const remove = async (s: ComplianceSchedule) => {
    try {
      await complianceService.deleteSchedule(s.id)
      toast.success(t('compliance.schedule.toast.deleted'))
      load()
    } catch {
      toast.error(t('compliance.schedule.toast.deleteError'))
    }
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="mb-3 flex shrink-0 items-center gap-2">
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('compliance.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
        <div className="ml-auto">
          <Button size="sm" onClick={() => setEditing('new')} disabled={selectable.length === 0}>
            <Plus size={14} className="mr-1.5" /> {t('compliance.schedule.new')}
          </Button>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto rounded-xl border border-border bg-card">
        {loading && schedules.length === 0 ? (
          <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('compliance.schedule.loading')}</Center>
        ) : error ? (
          <Center><AlertTriangle size={16} className="text-amber-500" /> {t('compliance.schedule.loadError')}<Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('compliance.retry')}</Button></Center>
        ) : schedules.length === 0 ? (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('compliance.schedule.empty')}</div>
        ) : (
          schedules.map((s) => (
            <div key={s.id} className="flex items-center gap-3 border-b border-border px-4 py-3 last:border-0">
              <CalendarClock size={15} className="shrink-0 text-muted-foreground" />
              <div className="min-w-0 flex-1">
                <div className="truncate text-[13px] font-medium">{fwName[s.frameworkKey] ?? s.frameworkKey}</div>
                <div className="flex flex-wrap items-center gap-x-3 text-[11px] text-muted-foreground">
                  <span>{cadenceLabel(s.scheduleString, t)} · <span className="font-mono">{s.scheduleString}</span></span>
                  <span>{t('compliance.schedule.windowDaysShort', { days: s.windowDays, defaultValue: `over ${s.windowDays}d` })}</span>
                  {s.to && <span className="inline-flex items-center gap-1"><Mail size={10} /> {s.to}{s.cc ? ` · cc ${s.cc}` : ''}</span>}
                </div>
              </div>
              <button onClick={() => setEditing(s)} className="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" title={t('compliance.schedule.edit')}>
                <Pencil size={14} />
              </button>
              <button onClick={() => void remove(s)} className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500/10 hover:text-red-500" title={t('compliance.schedule.delete')}>
                <Trash2 size={14} />
              </button>
            </div>
          ))
        )}
      </div>

      {editing && (
        <ScheduleEditor
          schedule={editing === 'new' ? null : editing}
          frameworks={selectable}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
          t={t}
        />
      )}
    </div>
  )
}

function ScheduleEditor({
  schedule,
  frameworks,
  onClose,
  onSaved,
  t,
}: {
  schedule: ComplianceSchedule | null
  frameworks: Framework[]
  onClose: () => void
  onSaved: () => void
  t: ReturnType<typeof useTranslation>['t']
}) {
  const [frameworkKey, setFrameworkKey] = useState(schedule?.frameworkKey ?? frameworks[0]?.key ?? '')
  const [cron, setCron] = useState(schedule?.scheduleString ?? CRON_PRESETS[0].cron)
  const [to, setTo] = useState(schedule?.to ?? '')
  const [cc, setCc] = useState(schedule?.cc ?? '')
  // How much the report covers, which is not how often it runs: "every morning
  // over the last 30 days" is the ordinary case.
  const [windowDays, setWindowDays] = useState(schedule?.windowDays ?? 30)
  const [busy, setBusy] = useState(false)

  const matchedPreset = CRON_PRESETS.find((p) => p.cron === cron.trim())?.key ?? 'custom'

  const save = async () => {
    if (!frameworkKey) {
      toast.error(t('compliance.schedule.frameworkRequired'))
      return
    }
    if (!cron.trim()) {
      toast.error(t('compliance.schedule.cronRequired'))
      return
    }
    if (!to.trim()) {
      toast.error(t('compliance.schedule.toRequired', { defaultValue: 'At least one recipient is required' }))
      return
    }
    const input: SaveScheduleInput = {
      frameworkKey,
      scheduleString: cron.trim(),
      windowDays,
      to: to.trim(),
      cc: cc.trim(),
    }
    if (schedule) input.id = schedule.id
    setBusy(true)
    try {
      if (schedule) await complianceService.updateSchedule(input)
      else await complianceService.createSchedule(input)
      toast.success(t('compliance.schedule.toast.saved'))
      onSaved()
    } catch (e) {
      toast.error(e instanceof ComplianceHttpError ? e.message : t('compliance.schedule.toast.saveError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[480px] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-center justify-between border-b border-border px-5 py-3.5">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <CalendarClock size={16} /> {schedule ? t('compliance.schedule.editTitle') : t('compliance.schedule.newTitle')}
          </h2>
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="space-y-4 p-5">
          <Field label={t('compliance.schedule.framework')}>
            <select value={frameworkKey} onChange={(e) => setFrameworkKey(e.target.value)} className={cn(SELECT, 'w-full')}>
              {frameworks.map((f) => (
                <option key={f.key} value={f.key}>{f.name}</option>
              ))}
            </select>
          </Field>

          <Field label={t('compliance.schedule.cadenceLabel')}>
            <select
              value={matchedPreset}
              onChange={(e) => {
                const p = CRON_PRESETS.find((x) => x.key === e.target.value)
                if (p) setCron(p.cron)
              }}
              className={cn(SELECT, 'w-full')}
            >
              {CRON_PRESETS.map((p) => (
                <option key={p.key} value={p.key}>{t(`compliance.schedule.cadence.${p.key}`)}</option>
              ))}
              <option value="custom">{t('compliance.schedule.cadence.custom')}</option>
            </select>
            <Input value={cron} onChange={(e) => setCron(e.target.value)} placeholder="0 6 1 * *" className="mt-1.5 font-mono text-xs" />
            <p className="mt-1 text-[11px] text-muted-foreground">{t('compliance.schedule.cronHint')}</p>
          </Field>

          <Field label={t('compliance.schedule.windowDays', { defaultValue: 'Reporting window (days)' })}>
            <Input
              type="number"
              min={1}
              max={3650}
              value={windowDays}
              onChange={(e) => setWindowDays(Number(e.target.value) || 30)}
            />
            <p className="mt-1 text-[11px] text-muted-foreground">
              {t('compliance.schedule.windowDaysHint', {
                defaultValue: 'How far back the report counts. Independent of how often it runs.',
              })}
            </p>
          </Field>

          <Field label={t('compliance.schedule.to', { defaultValue: 'To' })}>
            <Input value={to} onChange={(e) => setTo(e.target.value)} placeholder="soc@acme.com, ciso@acme.com" />
            <p className="mt-1 text-[11px] text-muted-foreground">{t('compliance.schedule.recipientsHint')}</p>
          </Field>

          <Field label={t('compliance.schedule.cc', { defaultValue: 'Cc' })}>
            <Input value={cc} onChange={(e) => setCc(e.target.value)} placeholder="audit@acme.com" />
          </Field>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-5 py-3">
          <Button size="sm" variant="outline" onClick={onClose} disabled={busy}>{t('compliance.schedule.cancel')}</Button>
          <Button size="sm" onClick={() => void save()} disabled={busy}>
            {busy ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null} {t('compliance.schedule.save')}
          </Button>
        </footer>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
