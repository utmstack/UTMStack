import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, ChevronDown, Code2, LayoutList, Loader2, Lock, Plus, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import {
  frameworkToForm,
  frameworkFormToYaml,
  formToFramework,
  yamlToFrameworkForm,
  type FrameworkFormState,
} from '../lib/framework-yaml'
import type { Control, Framework, FrameworkSection, Requirement } from '../types/compliance.types'
import { ControlEditor } from './ControlEditor'

interface ControlOpt {
  id: string
  name: string
  locked?: boolean
}

export function FrameworkEditor({
  framework,
  creating,
  onClose,
  onSaved,
}: {
  framework?: Framework
  creating: boolean
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  // The shipped catalogue is not editable, and neither is anything the
  // licence withholds.
  const readOnly = !!framework?.system || !!framework?.locked
  const [form, setForm] = useState<FrameworkFormState>(() => frameworkToForm(framework))
  const [mode, setMode] = useState<'visual' | 'code'>('visual')
  const [yaml, setYaml] = useState('')
  const [busy, setBusy] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [controls, setControls] = useState<ControlOpt[]>([])

  const set = <K extends keyof FrameworkFormState>(k: K, v: FrameworkFormState[K]) => setForm((f) => ({ ...f, [k]: v }))

  useEffect(() => {
    complianceService
      .listControls()
      .then((cs) => setControls((cs ?? []).map((c) => ({ id: c.id, name: c.name, locked: c.locked }))))
      .catch(() => setControls([]))
  }, [])

  const toCode = () => {
    setYaml(frameworkFormToYaml(form))
    setMode('code')
  }
  const toVisual = () => {
    const r = yamlToFrameworkForm(yaml)
    if (!r.ok) {
      toast.error(t('compliance.frameworks.yamlError', { error: r.error }))
      return
    }
    setForm(r.form)
    setMode('visual')
  }

  const save = async () => {
    if (busy) return
    let f = form
    if (mode === 'code') {
      const r = yamlToFrameworkForm(yaml)
      if (!r.ok) {
        toast.error(t('compliance.frameworks.yamlError', { error: r.error }))
        return
      }
      f = r.form
      setForm(f)
    }
    const payload = formToFramework(f)
    if (!payload.key) {
      toast.error(t('compliance.frameworks.keyRequired'))
      return
    }
    if (!payload.name) {
      toast.error(t('compliance.frameworks.nameRequired'))
      return
    }
    setBusy(true)
    try {
      if (creating) await complianceService.createFramework(payload)
      else await complianceService.updateFramework(framework!.key, payload)
      toast.success(creating ? t('compliance.frameworks.toast.created') : t('compliance.frameworks.toast.saved'))
      onSaved()
    } catch (e) {
      toast.error(e instanceof ComplianceHttpError ? e.message : t('compliance.frameworks.toast.saveError'))
    } finally {
      setBusy(false)
    }
  }

  const remove = async () => {
    if (!framework) return
    setBusy(true)
    try {
      await complianceService.deleteFramework(framework.key)
      toast.success(t('compliance.frameworks.toast.deleted'))
      onSaved()
    } catch {
      toast.error(t('compliance.frameworks.toast.deleteError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="flex items-center gap-2 truncate text-lg font-semibold">
              {creating ? t('compliance.frameworks.createTitle') : framework?.name}
              {readOnly && (
                <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                  <Lock size={9} /> {t('compliance.system')}
                </span>
              )}
            </h2>
            {!creating && <p className="mt-0.5 truncate font-mono text-[11px] text-muted-foreground">{framework?.key}</p>}
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <div className="inline-flex rounded-md border border-border p-0.5">
              <button type="button" onClick={() => mode !== 'visual' && toVisual()} className={cn('inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors', mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}>
                <LayoutList size={13} /> {t('compliance.frameworks.visualTab')}
              </button>
              <button type="button" onClick={() => mode !== 'code' && toCode()} className={cn('inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors', mode === 'code' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}>
                <Code2 size={13} /> {t('compliance.frameworks.codeTab')}
              </button>
            </div>
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
              <X size={16} />
            </button>
          </div>
        </header>

        {mode === 'code' ? (
          <div className="flex min-h-0 flex-1 flex-col p-6">
            <YamlCodeEditor value={yaml} onChange={setYaml} readOnly={readOnly} />
          </div>
        ) : (
          <div className="min-h-0 flex-1 space-y-4 overflow-y-auto p-6">
            <div className="grid grid-cols-2 gap-4">
              <Field label={t('compliance.frameworks.key')}>
                <Input value={form.key} readOnly={readOnly || !creating} onChange={(e) => set('key', e.target.value)} placeholder="iso-27001-2022" className="font-mono text-xs" />
              </Field>
              <Field label={t('compliance.frameworks.name')}>
                <Input value={form.name} readOnly={readOnly} onChange={(e) => set('name', e.target.value)} placeholder="ISO/IEC 27001:2022" />
              </Field>
            </div>
            <Field label={t('compliance.frameworks.description')}>
              <textarea value={form.description} readOnly={readOnly} onChange={(e) => set('description', e.target.value)} rows={2} className="w-full resize-none rounded-md border border-input bg-background/40 p-2 text-sm focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring read-only:opacity-80" />
            </Field>
            <Field label={t('compliance.frameworks.source')}>
              <Input value={form.source} readOnly={readOnly} onChange={(e) => set('source', e.target.value)} placeholder="https://… (optional)" className="text-xs" />
            </Field>

            <SectionsEditor
              sections={form.sections}
              controls={controls}
              readOnly={readOnly}
              onChange={(s) => set('sections', s)}
              onControlCreated={(c) => setControls((cur) => (cur.some((x) => x.id === c.id) ? cur : [...cur, { id: c.id, name: c.name }]))}
              t={t}
            />
          </div>
        )}

        <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
          <div>
            {!creating && !readOnly &&
              (confirmDelete ? (
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="destructive" onClick={() => void remove()} disabled={busy}><Trash2 size={13} className="mr-1.5" /> {t('compliance.frameworks.confirmDelete')}</Button>
                  <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)} disabled={busy}>{t('compliance.frameworks.cancel')}</Button>
                </div>
              ) : (
                <button onClick={() => setConfirmDelete(true)} className="rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300">{t('compliance.frameworks.delete')}</button>
              ))}
          </div>
          {!readOnly && (
            <Button size="sm" disabled={busy} onClick={() => void save()}>
              {busy ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null} {t('compliance.frameworks.save')}
            </Button>
          )}
        </footer>
      </div>
    </div>
  )
}

function SectionsEditor({ sections, controls, readOnly, onChange, onControlCreated, t }: { sections: FrameworkSection[]; controls: ControlOpt[]; readOnly?: boolean; onChange: (s: FrameworkSection[]) => void; onControlCreated: (c: Control) => void; t: ReturnType<typeof useTranslation>['t'] }) {
  const setAt = (i: number, patch: Partial<FrameworkSection>) => onChange(sections.map((s, k) => (k === i ? { ...s, ...patch } : s)))
  return (
    <div className="space-y-2">
      <label className="block text-xs font-medium text-foreground/80">{t('compliance.frameworks.sections')}</label>
      {sections.map((sec, i) => (
        <div key={i} className="space-y-3 rounded-lg border border-border bg-card/60 p-3">
          <div className="flex items-center gap-2">
            <Input value={sec.key} readOnly={readOnly} onChange={(e) => setAt(i, { key: e.target.value })} placeholder={t('compliance.frameworks.sectionKey')} className="h-8 w-28 font-mono text-xs" />
            <Input value={sec.name} readOnly={readOnly} onChange={(e) => setAt(i, { name: e.target.value })} placeholder={t('compliance.frameworks.sectionName')} className="h-8 flex-1 text-xs" />
            {!readOnly && (
              <button type="button" onClick={() => onChange(sections.filter((_, k) => k !== i))} className="rounded p-1 text-muted-foreground hover:text-red-500"><Trash2 size={13} /></button>
            )}
          </div>
          <RequirementsEditor requirements={sec.requirements ?? []} controls={controls} readOnly={readOnly} onChange={(reqs) => setAt(i, { requirements: reqs })} onControlCreated={onControlCreated} t={t} />
        </div>
      ))}
      {!readOnly && (
        <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => onChange([...sections, { key: '', name: '', requirements: [] }])}>
          <Plus size={12} className="mr-1" /> {t('compliance.frameworks.addSection')}
        </Button>
      )}
    </div>
  )
}

function RequirementsEditor({ requirements, controls, readOnly, onChange, onControlCreated, t }: { requirements: Requirement[]; controls: ControlOpt[]; readOnly?: boolean; onChange: (r: Requirement[]) => void; onControlCreated: (c: Control) => void; t: ReturnType<typeof useTranslation>['t'] }) {
  const setAt = (i: number, patch: Partial<Requirement>) => onChange(requirements.map((r, k) => (k === i ? { ...r, ...patch } : r)))
  return (
    <div className="space-y-1.5 border-l-2 border-border pl-3">
      <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{t('compliance.frameworks.requirements')}</div>
      {requirements.map((r, i) => (
        <div key={i} className="space-y-1 rounded-md bg-background/40 p-2">
          <div className="flex items-center gap-2">
            <Input value={r.id} readOnly={readOnly} onChange={(e) => setAt(i, { id: e.target.value })} placeholder={t('compliance.frameworks.reqId')} className="h-7 w-24 font-mono text-[11px]" />
            <Input value={r.name} readOnly={readOnly} onChange={(e) => setAt(i, { name: e.target.value })} placeholder={t('compliance.frameworks.reqName')} className="h-7 flex-1 text-[11px]" />
            {!readOnly && (
              <button type="button" onClick={() => onChange(requirements.filter((_, k) => k !== i))} className="rounded p-1 text-muted-foreground hover:text-red-500"><X size={12} /></button>
            )}
          </div>
          <ControlMultiSelect options={controls} values={r.satisfiedBy ?? []} readOnly={readOnly} onChange={(v) => setAt(i, { satisfiedBy: v })} onControlCreated={onControlCreated} t={t} />
        </div>
      ))}
      {!readOnly && (
        <button type="button" onClick={() => onChange([...requirements, { id: '', name: '', satisfiedBy: [] }])} className="flex items-center gap-1 text-[11px] text-primary hover:underline">
          <Plus size={11} /> {t('compliance.frameworks.addRequirement')}
        </button>
      )}
    </div>
  )
}

function ControlMultiSelect({ options, values, readOnly, onChange, onControlCreated, t }: { options: ControlOpt[]; values: string[]; readOnly?: boolean; onChange: (v: string[]) => void; onControlCreated: (c: Control) => void; t: ReturnType<typeof useTranslation>['t'] }) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])
  const toggle = (id: string) => onChange(values.includes(id) ? values.filter((x) => x !== id) : [...values, id])
  const filtered = options.filter((o) => (q ? (o.id + ' ' + o.name).toLowerCase().includes(q.toLowerCase()) : true)).slice(0, 60)

  return (
    <div ref={ref} className="relative">
      <div className="flex min-h-7 flex-wrap items-center gap-1 rounded-md border border-input bg-background/40 p-1">
        {values.map((v) => (
          <span key={v} className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">
            {v}
            {!readOnly && (
              <button type="button" onClick={() => toggle(v)} className="text-muted-foreground hover:text-red-500"><X size={10} /></button>
            )}
          </span>
        ))}
        {!readOnly && (
          <button type="button" onClick={() => setOpen((o) => !o)} className="inline-flex items-center gap-1 px-1 text-[10px] text-muted-foreground hover:text-foreground">
            <Plus size={10} /> {values.length ? '' : t('compliance.frameworks.satisfiedBy')}
            <ChevronDown size={9} className="opacity-60" />
          </button>
        )}
      </div>
      {open && !readOnly && (
        <div className="absolute left-0 top-full z-30 mt-1 w-72 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input value={q} onChange={(e) => setQ(e.target.value)} autoFocus placeholder={t('compliance.frameworks.searchControls')} className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
          </div>
          <div className="max-h-52 overflow-y-auto">
            {filtered.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('compliance.frameworks.noControls')}</div>}
            {filtered.map((o) => {
              const on = values.includes(o.id)
              const locked = !!o.locked && !on // allow deselecting an already-set one
              return (
                <button
                  key={o.id}
                  type="button"
                  disabled={locked}
                  title={locked ? t('compliance.locked.upsell') : undefined}
                  onClick={() => toggle(o.id)}
                  className={cn('flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-transparent')}
                >
                  <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{o.id}</span>
                  <span className="min-w-0 flex-1 truncate">{o.name}</span>
                  {locked && <Lock size={11} className="shrink-0 text-amber-500" />}
                  {on && <Check size={12} className="shrink-0 text-primary" />}
                </button>
              )
            })}
          </div>
          <button
            type="button"
            onClick={() => {
              setOpen(false)
              setCreateOpen(true)
            }}
            className="flex w-full items-center gap-2 border-t border-border px-3 py-1.5 text-left text-xs font-medium text-primary hover:bg-muted"
          >
            <Plus size={12} className="shrink-0" /> {q.trim() ? t('compliance.frameworks.createNamed', { name: q.trim() }) : t('compliance.frameworks.createControl')}
          </button>
        </div>
      )}
      {createOpen && (
        <ControlEditor
          creating
          prefillName={q.trim()}
          onClose={() => setCreateOpen(false)}
          onSaved={(created) => {
            if (created) {
              onChange([...values, created.id])
              onControlCreated(created)
            }
            setCreateOpen(false)
          }}
        />
      )}
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
